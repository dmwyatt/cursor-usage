package output

import (
	"encoding/json"
	"io"
)

// RenderJSON writes v as indented JSON to w.
func RenderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
