package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestRenderJSON(t *testing.T) {
	data := map[string]any{
		"name":  "test",
		"count": 42,
	}

	var buf bytes.Buffer
	if err := RenderJSON(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if parsed["name"] != "test" {
		t.Errorf("expected name=test, got %v", parsed["name"])
	}

	// Verify it's indented (contains newlines beyond just the trailing one)
	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("\n  ")) {
		t.Error("expected indented output")
	}
}
