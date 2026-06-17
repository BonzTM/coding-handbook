package core_test

import (
	"errors"
	"testing"
	"time"

	"github.com/example/exampleservice/internal/core"
)

// TestCursorRoundTrip proves a non-zero cursor survives encode/decode intact,
// to nanosecond precision, so a page boundary resumes exactly where it left off.
func TestCursorRoundTrip(t *testing.T) {
	want := core.Cursor{
		CreatedAt: time.Unix(1700000000, 123456789).UTC(),
		ID:        "w-42",
	}
	token := core.EncodeCursor(want)
	if token == "" {
		t.Fatal("EncodeCursor of a non-zero cursor returned empty token")
	}
	got, err := core.DecodeCursor(token)
	if err != nil {
		t.Fatalf("DecodeCursor: %v", err)
	}
	if !got.CreatedAt.Equal(want.CreatedAt) || got.ID != want.ID {
		t.Errorf("round-trip = %+v, want %+v", got, want)
	}
}

// TestZeroCursorEncodesEmpty pins the last-page contract: the zero cursor
// encodes to "" and "" decodes back to the zero cursor.
func TestZeroCursorEncodesEmpty(t *testing.T) {
	if tok := core.EncodeCursor(core.Cursor{}); tok != "" {
		t.Errorf("EncodeCursor(zero) = %q, want empty", tok)
	}
	got, err := core.DecodeCursor("")
	if err != nil {
		t.Fatalf("DecodeCursor(\"\"): %v", err)
	}
	if !got.IsZero() {
		t.Errorf("DecodeCursor(\"\") = %+v, want zero cursor", got)
	}
}

// TestDecodeCursorInvalid proves a malformed token is a typed ErrInvalidCursor
// (mapped to 400 at the boundary), never a panic or a silent zero cursor.
func TestDecodeCursorInvalid(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"not base64", "!!!not-base64!!!"},
		{"base64 but not json", "Zm9vYmFy"},                           // "foobar"
		{"json missing id", "eyJ0IjoiMjAyMy0wMS0wMVQwMDowMDowMFoifQ"}, // {"t":"2023-01-01T00:00:00Z"}
		{"unknown field", "eyJ4IjoxfQ"},                               // {"x":1}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := core.DecodeCursor(tt.token)
			if !errors.Is(err, core.ErrInvalidCursor) {
				t.Errorf("DecodeCursor(%q) error = %v, want ErrInvalidCursor", tt.token, err)
			}
		})
	}
}
