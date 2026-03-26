package output

import (
	"encoding/json"
	"io"
)

// JSONFormatter formats output as JSON
type JSONFormatter struct{}

// Format writes data as formatted JSON
func (f *JSONFormatter) Format(w io.Writer, data any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
