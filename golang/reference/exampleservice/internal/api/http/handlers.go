package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/example/exampleservice/internal/core"
)

// --- Wire DTOs -------------------------------------------------------------
//
// These are dedicated transport types with explicit snake_case json tags, per
// golang/foundations/serialization.md. They are mapped to/from core.Widget at
// the boundary; core types are never serialized directly.

// createWidgetRequest is the POST /widgets body.
type createWidgetRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// widgetResponse is the wire representation of a widget.
type widgetResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// listWidgetsResponse is the keyset-pagination envelope mandated by
// golang/services/http-services.md ### List Endpoints And Pagination: an items
// array plus an opaque next_cursor. next_cursor is "" on the last page. The
// shape is a stable wire contract shared across list endpoints.
type listWidgetsResponse struct {
	Items      []widgetResponse `json:"items"`
	NextCursor string           `json:"next_cursor"`
}

func toWidgetResponse(w core.Widget) widgetResponse {
	return widgetResponse{
		ID:        w.ID,
		Name:      w.Name,
		CreatedAt: w.CreatedAt.UTC(),
	}
}

// --- Handlers --------------------------------------------------------------
//
// Each handler follows the five-step contract: decode -> validate -> call one
// core method -> map error -> encode.

// handleCreateWidget handles POST /widgets.
func (s *Server) handleCreateWidget(w http.ResponseWriter, r *http.Request) {
	req, err := decodeJSON[createWidgetRequest](w, r, s.maxBodyBytes)
	if err != nil {
		// A decode failure is a client error; surface a 400 with the reason.
		writeError(w, r, s.logger, fmt.Errorf("%w: %s", core.ErrInvalidWidget, err.Error()))
		return
	}

	widget, err := s.svc.CreateWidget(r.Context(), req.ID, req.Name)
	if err != nil {
		writeError(w, r, s.logger, err)
		return
	}

	s.metrics.IncWidgetCreated()
	writeJSON(w, r, s.logger, http.StatusCreated, toWidgetResponse(widget))
}

// handleGetWidget handles GET /widgets/{id}.
func (s *Server) handleGetWidget(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	widget, err := s.svc.GetWidget(r.Context(), id)
	if err != nil {
		writeError(w, r, s.logger, err)
		return
	}
	writeJSON(w, r, s.logger, http.StatusOK, toWidgetResponse(widget))
}

// handleListWidgets handles GET /widgets with keyset pagination. It reads an
// optional page_size (clamped server-side) and an opaque cursor query
// parameter, decodes the cursor, asks the service for one page, and returns the
// {items, next_cursor} envelope. A malformed cursor or non-numeric page_size is
// a client error (400) mapped at the boundary; an oversized page_size is clamped
// by the service, never rejected.
func (s *Server) handleListWidgets(w http.ResponseWriter, r *http.Request) {
	pageSize, err := parsePageSize(r.URL.Query().Get("page_size"))
	if err != nil {
		writeError(w, r, s.logger, err)
		return
	}

	cursor, err := core.DecodeCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(w, r, s.logger, err)
		return
	}

	page, err := s.svc.ListWidgetsPage(r.Context(), cursor, pageSize)
	if err != nil {
		writeError(w, r, s.logger, err)
		return
	}

	resp := listWidgetsResponse{
		Items:      make([]widgetResponse, 0, len(page.Widgets)),
		NextCursor: core.EncodeCursor(page.NextCursor),
	}
	for _, wg := range page.Widgets {
		resp.Items = append(resp.Items, toWidgetResponse(wg))
	}
	writeJSON(w, r, s.logger, http.StatusOK, resp)
}

// parsePageSize reads the optional page_size query parameter. An empty value
// means "use the server default" and yields 0 (which the service treats as the
// default). A present-but-non-numeric value is a client error mapped to 400 via
// ErrInvalidWidget; the service clamps an oversized numeric value rather than
// rejecting it.
func parsePageSize(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%w: page_size must be an integer", core.ErrInvalidWidget)
	}
	return n, nil
}

// handleLivez reports process liveness: if this handler runs, the process is
// alive. It must NOT depend on downstream readiness, or the platform would kill
// a draining pod.
func (s *Server) handleLivez(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	s.writePlain(w, r, "ok")
}

// handleReadyz reports readiness to receive traffic. It reflects the readiness
// flag, which the shutdown sequence flips to false before draining.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	if !s.readiness.Ready() {
		w.WriteHeader(http.StatusServiceUnavailable)
		s.writePlain(w, r, "not ready")
		return
	}
	w.WriteHeader(http.StatusOK)
	s.writePlain(w, r, "ready")
}

// writePlain writes a short text probe body. The status is already committed,
// so a write error is unactionable for the client; it is logged once so a
// broken probe connection is observable rather than silently dropped.
func (s *Server) writePlain(w http.ResponseWriter, r *http.Request, body string) {
	if _, err := io.WriteString(w, body); err != nil {
		s.logger.WarnContext(r.Context(), "write probe body failed",
			"route", routePattern(r),
			"error", err.Error(),
		)
	}
}

// decodeJSON bounds the body, strictly decodes a single JSON value of type T,
// and rejects unknown fields. This is an internal/versioned surface, so the
// strict policy (DisallowUnknownFields) applies per the serialization doc.
func decodeJSON[T any](w http.ResponseWriter, r *http.Request, maxBytes int64) (T, error) {
	var v T
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return v, fmt.Errorf("decode %T: %w", v, err)
	}
	// Reject trailing data so two concatenated objects are not silently
	// accepted as one.
	if err := dec.Decode(new(struct{})); !errors.Is(err, io.EOF) {
		return v, errors.New("request body must contain a single JSON object")
	}
	return v, nil
}
