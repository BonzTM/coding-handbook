// Package core holds the widgets domain logic and the contracts it consumes.
// It depends only on the standard library and narrowly scoped seams; it imports
// no transport or storage-implementation packages, per
// golang/foundations/package-design.md.
//
// The Store interface is defined HERE, in the consumer, and names what the
// service needs ("Store"), not what an implementer is. The in-memory
// implementation lives in this package (memory.go) and satisfies it
// structurally.
package core

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Sentinel errors callers (notably the gRPC server) branch on with errors.Is.
// They are part of the package contract; the transport maps them to codes.*.
var (
	// ErrNotFound is returned when a widget does not exist.
	ErrNotFound = errors.New("widget not found")
	// ErrAlreadyExists is returned when creating a widget whose ID is taken.
	ErrAlreadyExists = errors.New("widget already exists")
	// ErrInvalidWidget is returned when a widget fails validation.
	ErrInvalidWidget = errors.New("invalid widget")
	// ErrInvalidCursor is returned when a pagination page token cannot be decoded.
	// It is mapped to InvalidArgument at the transport boundary.
	ErrInvalidCursor = errors.New("invalid page token")
)

// Pagination bounds for List endpoints: the server enforces a default and a
// maximum page size and clamps an oversized request rather than rejecting it.
const (
	// DefaultPageSize is used when a caller requests no (or a non-positive)
	// page size.
	DefaultPageSize = 20
	// MaxPageSize is the hard server-side ceiling; larger requests are clamped
	// down to it, never honored unbounded.
	MaxPageSize = 100
)

// Cursor is the opaque keyset position a List page resumes after. It carries
// the stable sort key — (CreatedAt, ID) — of the LAST row of the previous page,
// so the next page starts strictly after it. A zero Cursor means "from the
// beginning". Callers treat it as opaque; only EncodeCursor/DecodeCursor and the
// store interpret it. On the wire it is the page_token / next_page_token field.
type Cursor struct {
	CreatedAt time.Time
	ID        string
}

// IsZero reports whether the cursor is the start-of-collection sentinel.
func (c Cursor) IsZero() bool { return c.ID == "" && c.CreatedAt.IsZero() }

// cursorWire is the JSON shape encoded inside the opaque base64 token. It is an
// internal detail; the keys are short and the time is RFC3339Nano for a stable,
// total order.
type cursorWire struct {
	CreatedAt time.Time `json:"t"`
	ID        string    `json:"i"`
}

// EncodeCursor renders a Cursor as an opaque, URL-safe base64 token. The zero
// cursor encodes to the empty string so a last page reports next_page_token "".
func EncodeCursor(c Cursor) string {
	if c.IsZero() {
		return ""
	}
	raw, err := json.Marshal(cursorWire{CreatedAt: c.CreatedAt.UTC(), ID: c.ID})
	if err != nil {
		// cursorWire is a fixed, always-marshalable shape; an error here is a
		// programming error, not a runtime condition.
		panic(fmt.Sprintf("core: marshal cursor: %v", err))
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeCursor parses an opaque token produced by EncodeCursor. The empty string
// decodes to the zero (start) cursor. A malformed token returns ErrInvalidCursor
// so the transport can map it to InvalidArgument instead of Internal.
func DecodeCursor(token string) (Cursor, error) {
	if token == "" {
		return Cursor{}, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return Cursor{}, fmt.Errorf("%w: %s", ErrInvalidCursor, err.Error())
	}
	var w cursorWire
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&w); err != nil {
		return Cursor{}, fmt.Errorf("%w: %s", ErrInvalidCursor, err.Error())
	}
	if w.ID == "" {
		return Cursor{}, fmt.Errorf("%w: cursor id must not be empty", ErrInvalidCursor)
	}
	return Cursor{CreatedAt: w.CreatedAt.UTC(), ID: w.ID}, nil
}

// Page is one keyset page of widgets plus the cursor to fetch the next page.
// NextCursor is the zero Cursor on the last page, which EncodeCursor renders as
// the empty string in the response's next_page_token.
type Page struct {
	Widgets    []Widget
	NextCursor Cursor
}

// Widget is the widgets domain entity. It is a plain value type with no wire
// concerns: serialization lives on the generated proto messages in the
// transport package.
type Widget struct {
	ID        string
	TenantID  string
	Name      string
	CreatedAt time.Time
}

// Clock is the source of time for the service so behavior is deterministic in
// tests. Production wires a real clock in main; tests pass a fake. The
// interface is kept to the single method the service actually needs.
type Clock interface {
	// Now returns the current time. Implementations return UTC.
	Now() time.Time
}

// Store is the persistence contract the service consumes. It is intentionally
// narrow (interface-at-consumer); the in-memory implementation lives in
// memory.go.
//
// Every method is scoped by tenantID: the store filters on the tenant so one
// tenant can never observe another's rows. Multi-tenancy is enforced in the
// storage layer, not only at the edge, so a missing boundary check cannot leak
// cross-tenant data. See golang/services/database.md.
type Store interface {
	// Create persists a new widget within w.TenantID. It returns ErrAlreadyExists
	// if a widget with the same (tenant_id, id) is already stored.
	Create(ctx context.Context, w Widget) error
	// Get returns the widget with the given ID within tenantID, or ErrNotFound.
	// A widget that exists under a different tenant is reported as ErrNotFound.
	Get(ctx context.Context, tenantID, id string) (Widget, error)
	// ListPage returns up to limit widgets WITHIN tenantID ordered by the stable
	// keyset (CreatedAt, ID), starting strictly after the given cursor. A zero
	// cursor starts at the beginning. It returns at most limit widgets; the caller
	// derives the next cursor from the last returned widget.
	ListPage(ctx context.Context, tenantID string, after Cursor, limit int) ([]Widget, error)
	// Close releases store resources. The in-memory store holds none; the method
	// exists so main can sequence an ordered shutdown uniformly.
	Close() error
}

// Service is the widgets domain service. It is constructed with its
// dependencies and threads context explicitly; it never stores a context and
// never reads the wall clock directly.
type Service struct {
	store Store
	clock Clock
}

// NewService constructs a Service. store and clock are required; passing nil is
// a programming error and will panic on first use rather than silently
// misbehave.
func NewService(store Store, clock Clock) *Service {
	return &Service{store: store, clock: clock}
}

const maxNameLen = 128

// CreateWidget validates the input, stamps a UTC creation time from the
// injected clock, and persists the widget under the caller's tenant. The
// authenticated principal is resolved from ctx: a missing principal is
// ErrUnauthenticated and a principal without RoleWriter is ErrForbidden. The
// widget's TenantID is taken from the principal, never from the request body,
// so a caller cannot write into another tenant. Validation failures return
// ErrInvalidWidget; a duplicate ID surfaces the store's ErrAlreadyExists.
func (s *Service) CreateWidget(ctx context.Context, id, name string) (Widget, error) {
	p, ok := PrincipalFrom(ctx)
	if !ok {
		return Widget{}, ErrUnauthenticated
	}
	if !p.HasRole(RoleWriter) {
		return Widget{}, fmt.Errorf("%w: %s requires role %s", ErrForbidden, p.Subject, RoleWriter)
	}
	if id == "" {
		return Widget{}, errInvalidf("id", "id must not be empty")
	}
	if name == "" {
		return Widget{}, errInvalidf("name", "name must not be empty")
	}
	if len(name) > maxNameLen {
		return Widget{}, errInvalidf("name", "name must be at most %d bytes", maxNameLen)
	}

	w := Widget{
		ID:        id,
		TenantID:  p.TenantID,
		Name:      name,
		CreatedAt: s.clock.Now().UTC(),
	}
	if err := s.store.Create(ctx, w); err != nil {
		// Return as-is so callers can still match ErrAlreadyExists / context.
		return Widget{}, err
	}
	return w, nil
}

// GetWidget returns the widget with the given ID within the caller's tenant. A
// missing principal is ErrUnauthenticated and a principal without RoleReader is
// ErrForbidden. The store is scoped by the principal's tenant, so a widget that
// belongs to another tenant surfaces ErrNotFound (a cross-tenant read is
// indistinguishable from a missing row, by design).
func (s *Service) GetWidget(ctx context.Context, id string) (Widget, error) {
	p, ok := PrincipalFrom(ctx)
	if !ok {
		return Widget{}, ErrUnauthenticated
	}
	if !p.HasRole(RoleReader) {
		return Widget{}, fmt.Errorf("%w: %s requires role %s", ErrForbidden, p.Subject, RoleReader)
	}
	if id == "" {
		return Widget{}, errInvalidf("id", "id must not be empty")
	}
	return s.store.Get(ctx, p.TenantID, id)
}

// ClampPageSize applies the server-side pagination policy: a non-positive
// request falls back to DefaultPageSize, and any request above MaxPageSize is
// clamped down to it rather than rejected.
func ClampPageSize(requested int) int {
	switch {
	case requested <= 0:
		return DefaultPageSize
	case requested > MaxPageSize:
		return MaxPageSize
	default:
		return requested
	}
}

// ListWidgetsPage returns one keyset page of widgets after the given cursor. It
// clamps the page size, asks the store for limit+1 rows to detect whether more
// pages exist without a separate COUNT, and trims the extra row. NextCursor is
// the zero Cursor on the last page (which encodes to "").
func (s *Service) ListWidgetsPage(ctx context.Context, after Cursor, pageSize int) (Page, error) {
	p, ok := PrincipalFrom(ctx)
	if !ok {
		return Page{}, ErrUnauthenticated
	}
	if !p.HasRole(RoleReader) {
		return Page{}, fmt.Errorf("%w: %s requires role %s", ErrForbidden, p.Subject, RoleReader)
	}
	limit := ClampPageSize(pageSize)

	// Fetch one extra row: if the store returns it, there is at least one more
	// page and the trimmed last row becomes the next cursor. This avoids an
	// exact COUNT on a potentially large table.
	widgets, err := s.store.ListPage(ctx, p.TenantID, after, limit+1)
	if err != nil {
		return Page{}, err
	}

	page := Page{Widgets: widgets}
	if len(widgets) > limit {
		page.Widgets = widgets[:limit]
		last := page.Widgets[limit-1]
		page.NextCursor = Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return page, nil
}

// FieldViolation names a single request field that failed validation together
// with a human-readable reason. It is the structured, transport-agnostic carrier
// the gRPC boundary renders as a google.rpc.BadRequest field violation; core
// stays free of any wire/proto dependency.
type FieldViolation struct {
	// Field is the offending request field (e.g. "id", "name").
	Field string
	// Description is the actionable reason the field is invalid.
	Description string
}

// FieldViolations extracts the per-field validation violations carried by an
// ErrInvalidWidget error, or nil if err is not a structured validation error.
// The transport uses it to attach a google.rpc.BadRequest detail without
// importing core's unexported error type or parsing messages.
func FieldViolations(err error) []FieldViolation {
	var ie *invalidError
	if errors.As(err, &ie) && ie.field != "" {
		return []FieldViolation{{Field: ie.field, Description: ie.reason}}
	}
	return nil
}

// errInvalidf builds an ErrInvalidWidget-wrapping error for the named field with
// a specific reason so the message is actionable, the field is structured for a
// BadRequest detail, and errors.Is(err, ErrInvalidWidget) holds.
func errInvalidf(field, format string, args ...any) error {
	return &invalidError{field: field, reason: fmt.Sprintf(format, args...)}
}

type invalidError struct {
	field  string
	reason string
}

func (e *invalidError) Error() string { return "invalid widget: " + e.reason }

// Unwrap lets errors.Is(err, ErrInvalidWidget) succeed.
func (e *invalidError) Unwrap() error { return ErrInvalidWidget }
