package output

import (
	"io"
)

// Format represents the output format type
type Format string

const (
	FormatJSON  Format = "json"
	FormatTable Format = "table"
)

// Formatter is the interface for output formatters
type Formatter interface {
	Format(w io.Writer, data any) error
}

// NewFormatter creates a new formatter based on the format type
func NewFormatter(format Format) Formatter {
	switch format {
	case FormatJSON:
		return &JSONFormatter{}
	case FormatTable:
		return &TableFormatter{}
	default:
		return &TableFormatter{}
	}
}
