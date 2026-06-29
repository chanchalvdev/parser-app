package repositories

import (
	"encoding/json"
	"testing"
)

func TestAsJSON(t *testing.T) {
	t.Run("empty payload", func(t *testing.T) {
		got, err := asJSON(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "{}" {
			t.Fatalf("expected {}, got %s", got)
		}
	})

	t.Run("invalid payload", func(t *testing.T) {
		invalid := json.RawMessage(`{bad json}`)
		if _, err := asJSON(invalid); err == nil {
			t.Fatalf("expected error for invalid payload")
		}
	})
}

func TestScanJSON(t *testing.T) {
	t.Run("json bytes", func(t *testing.T) {
		input := []byte(`{"a":1}`)
		got, err := scanJSON(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != `{"a":1}` {
			t.Fatalf("unexpected parsed value: %s", got)
		}
	})
}

