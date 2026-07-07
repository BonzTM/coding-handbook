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

// FuzzDecodeCursor fuzzes the untrusted-input edge of pagination: the token is
// base64 straight off the wire, exactly the shape fuzzing is for, per
// golang/quality/testing.md ### Fuzzing. It asserts invariants, not exact
// outputs: every failure is the typed ErrInvalidCursor with a zero cursor
// (reject, don't crash), and every success round-trips through EncodeCursor.
// The f.Add seeds plus the committed corpus under testdata/fuzz/FuzzDecodeCursor/
// replay as ordinary deterministic tests in `go test` (so inside make verify);
// exploration runs on demand: go test -fuzz=FuzzDecodeCursor -fuzztime=30s ./internal/core/
func FuzzDecodeCursor(f *testing.F) {
	// One valid token plus a near-miss for each rejection branch of DecodeCursor.
	f.Add(core.EncodeCursor(core.Cursor{CreatedAt: time.Unix(1700000000, 123456789).UTC(), ID: "w-42"}))
	f.Add("")                                       // start-of-collection sentinel
	f.Add("!!!not-base64!!!")                       // rejected before JSON decoding
	f.Add("Zm9vYmFy")                               // "foobar": base64 but not JSON
	f.Add("eyJ0IjoiMjAyMy0wMS0wMVQwMDowMDowMFoifQ") // {"t":"2023-01-01T00:00:00Z"}: missing id

	f.Fuzz(func(t *testing.T, token string) {
		got, err := core.DecodeCursor(token)
		if err != nil {
			if !errors.Is(err, core.ErrInvalidCursor) {
				t.Errorf("DecodeCursor(%q) error = %v, want ErrInvalidCursor", token, err)
			}
			if !got.IsZero() {
				t.Errorf("DecodeCursor(%q) = %+v alongside an error, want zero cursor", token, got)
			}
			return
		}
		// Round-trip: whatever decoded must survive encode/decode unchanged, so a
		// client echoing the token back resumes at exactly the same position. This
		// also proves EncodeCursor cannot be driven to its marshal panic by any
		// value DecodeCursor accepts.
		again, err := core.DecodeCursor(core.EncodeCursor(got))
		if err != nil {
			t.Fatalf("re-decode of re-encoded cursor %+v: %v", got, err)
		}
		if !again.CreatedAt.Equal(got.CreatedAt) || again.ID != got.ID {
			t.Errorf("round-trip = %+v, want %+v", again, got)
		}
	})
}
