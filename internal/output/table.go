package output

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

// TableFormatter formats output as a table
type TableFormatter struct {
	NoColor bool
}

// Format writes data as a table
func (f *TableFormatter) Format(w io.Writer, data any) error {
	v := reflect.ValueOf(data)

	// Handle pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice:
		return f.formatSlice(w, v)
	case reflect.Struct:
		return f.formatStruct(w, v)
	case reflect.Map:
		return f.formatMap(w, v)
	default:
		// For other types, just print the value
		_, err := fmt.Fprintln(w, data)
		return err
	}
}

func (f *TableFormatter) formatSlice(w io.Writer, v reflect.Value) error {
	if v.Len() == 0 {
		fmt.Fprintln(w, "No results found.")
		return nil
	}

	table := tablewriter.NewWriter(w)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	// Get headers from first element
	first := v.Index(0)
	if first.Kind() == reflect.Ptr {
		first = first.Elem()
	}

	headers, fields := getHeadersAndFields(first.Type())
	table.SetHeader(headers)

	// Add rows
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		row := getRowValues(elem, fields)
		table.Append(row)
	}

	table.Render()
	return nil
}

func (f *TableFormatter) formatStruct(w io.Writer, v reflect.Value) error {
	table := tablewriter.NewWriter(w)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator(":")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name := getFieldDisplayName(field)
		value := formatValue(v.Field(i))
		table.Append([]string{name, value})
	}

	table.Render()
	return nil
}

func (f *TableFormatter) formatMap(w io.Writer, v reflect.Value) error {
	table := tablewriter.NewWriter(w)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator(":")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")

	for _, key := range v.MapKeys() {
		value := v.MapIndex(key)
		table.Append([]string{fmt.Sprint(key.Interface()), formatValue(value)})
	}

	table.Render()
	return nil
}

func getHeadersAndFields(t reflect.Type) ([]string, []int) {
	var headers []string
	var fields []int

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// Skip complex types for table display
		if field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Map ||
			(field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{})) {
			continue
		}

		name := getFieldDisplayName(field)
		headers = append(headers, strings.ToUpper(name))
		fields = append(fields, i)
	}

	return headers, fields
}

func getFieldDisplayName(field reflect.StructField) string {
	// Check for json tag first
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" && parts[0] != "-" {
			return parts[0]
		}
	}
	return field.Name
}

func getRowValues(v reflect.Value, fields []int) []string {
	var values []string
	for _, i := range fields {
		values = append(values, formatValue(v.Field(i)))
	}
	return values
}

func formatValue(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}

	// Handle pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	// Handle special types
	if t, ok := v.Interface().(time.Time); ok {
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02 15:04:05")
	}

	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case reflect.Slice, reflect.Map:
		return fmt.Sprintf("[%d items]", v.Len())
	default:
		return fmt.Sprint(v.Interface())
	}
}
