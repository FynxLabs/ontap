package output

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v2"
)

// Format represents an output format
type Format string

const (
	// FormatJSON represents JSON output
	FormatJSON Format = "json"

	// FormatYAML represents YAML output
	FormatYAML Format = "yaml"

	// FormatCSV represents CSV output
	FormatCSV Format = "csv"

	// FormatText represents plain text output
	FormatText Format = "text"

	// FormatTable represents table output
	FormatTable Format = "table"
)

// Formatter is the interface for formatting output
type Formatter interface {
	// Format formats the data
	Format(data interface{}) ([]byte, error)

	// GetFormat returns the format
	GetFormat() Format
}

// JSONFormatter formats output as JSON
type JSONFormatter struct {
	// Pretty indicates whether to pretty-print the JSON
	Pretty bool
}

// NewJSONFormatter creates a new JSONFormatter
func NewJSONFormatter(pretty bool) *JSONFormatter {
	return &JSONFormatter{
		Pretty: pretty,
	}
}

// Format formats the data as JSON
func (f *JSONFormatter) Format(data interface{}) ([]byte, error) {
	if f.Pretty {
		return json.MarshalIndent(data, "", "  ")
	}
	return json.Marshal(data)
}

// GetFormat returns the format
func (f *JSONFormatter) GetFormat() Format {
	return FormatJSON
}

// YAMLFormatter formats output as YAML
type YAMLFormatter struct{}

// NewYAMLFormatter creates a new YAMLFormatter
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{}
}

// Format formats the data as YAML
func (f *YAMLFormatter) Format(data interface{}) ([]byte, error) {
	return yaml.Marshal(data)
}

// GetFormat returns the format
func (f *YAMLFormatter) GetFormat() Format {
	return FormatYAML
}

// CSVFormatter formats output as CSV
type CSVFormatter struct {
	// Header indicates whether to include a header row
	Header bool

	// Delimiter is the CSV delimiter
	Delimiter rune
}

// NewCSVFormatter creates a new CSVFormatter
func NewCSVFormatter(header bool, delimiter rune) *CSVFormatter {
	if delimiter == 0 {
		delimiter = ','
	}
	return &CSVFormatter{
		Header:    header,
		Delimiter: delimiter,
	}
}

// Format formats the data as CSV
func (f *CSVFormatter) Format(data interface{}) ([]byte, error) {
	// Convert the data to a slice of maps
	var rows []map[string]interface{}
	switch v := data.(type) {
	case []map[string]interface{}:
		rows = v
	case map[string]interface{}:
		rows = []map[string]interface{}{v}
	case []interface{}:
		rows = make([]map[string]interface{}, len(v))
		for i, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				rows[i] = m
			} else {
				rows[i] = map[string]interface{}{"value": item}
			}
		}
	default:
		return nil, fmt.Errorf("unsupported data type for CSV: %T", data)
	}

	// Get all the column names
	columns := make(map[string]bool)
	for _, row := range rows {
		for k := range row {
			columns[k] = true
		}
	}

	// Convert the column names to a slice
	var columnNames []string
	for k := range columns {
		columnNames = append(columnNames, k)
	}

	// Create a buffer to write the CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Comma = f.Delimiter

	// Write the header row
	if f.Header {
		if err := writer.Write(columnNames); err != nil {
			return nil, fmt.Errorf("failed to write CSV header: %w", err)
		}
	}

	// Write the data rows
	for _, row := range rows {
		var values []string
		for _, col := range columnNames {
			val := row[col]
			values = append(values, fmt.Sprintf("%v", val))
		}
		if err := writer.Write(values); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	return buf.Bytes(), nil
}

// GetFormat returns the format
func (f *CSVFormatter) GetFormat() Format {
	return FormatCSV
}

// TextFormatter formats output as plain text
type TextFormatter struct{}

// NewTextFormatter creates a new TextFormatter
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

// Format formats the data as plain text
func (f *TextFormatter) Format(data interface{}) ([]byte, error) {
	return []byte(fmt.Sprintf("%v", data)), nil
}

// GetFormat returns the format
func (f *TextFormatter) GetFormat() Format {
	return FormatText
}

// TableFormatter formats output as a table
type TableFormatter struct {
	// Header indicates whether to include a header row
	Header bool

	// Columns are the columns to include in the table
	Columns []string
}

// NewTableFormatter creates a new TableFormatter
func NewTableFormatter(header bool, columns []string) *TableFormatter {
	return &TableFormatter{
		Header:  header,
		Columns: columns,
	}
}

// Format formats the data as a table
func (f *TableFormatter) Format(data interface{}) ([]byte, error) {
	// For now, just use the CSV formatter
	csvFormatter := NewCSVFormatter(f.Header, '|')
	return csvFormatter.Format(data)
}

// GetFormat returns the format
func (f *TableFormatter) GetFormat() Format {
	return FormatTable
}

// NewFormatter creates a new Formatter based on the format
func NewFormatter(format string) (Formatter, error) {
	switch strings.ToLower(format) {
	case "json":
		return NewJSONFormatter(true), nil
	case "yaml", "yml":
		return NewYAMLFormatter(), nil
	case "csv":
		return NewCSVFormatter(true, ','), nil
	case "text", "txt":
		return NewTextFormatter(), nil
	case "table":
		return NewTableFormatter(true, nil), nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}

// FormatData formats data using the specified formatter
func FormatData(data interface{}, formatter Formatter) ([]byte, error) {
	return formatter.Format(data)
}

// WriteOutput writes formatted data to the specified output
func WriteOutput(data interface{}, formatter Formatter, output string) error {
	// Format the data
	formattedData, err := formatter.Format(data)
	if err != nil {
		return fmt.Errorf("failed to format data: %w", err)
	}

	// Write the data to the output
	if output == "" || output == "-" {
		// Write to stdout
		_, err = os.Stdout.Write(formattedData)
		if err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
	} else {
		// Write to file
		err = os.WriteFile(output, formattedData, 0644)
		if err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
		log.Info("Output written to file", "path", output)
	}

	return nil
}

// ExtractFields extracts fields from data using a dot notation path
func ExtractFields(data interface{}, fields []string) (interface{}, error) {
	if len(fields) == 0 {
		return data, nil
	}

	// Convert the data to JSON and back to a map to ensure a consistent structure
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var mapData interface{}
	if err := json.Unmarshal(jsonData, &mapData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Extract the fields
	result := make(map[string]interface{})
	for _, field := range fields {
		value, err := extractField(mapData, field)
		if err != nil {
			log.Warn("Failed to extract field", "field", field, "error", err)
			continue
		}
		result[field] = value
	}

	return result, nil
}

// extractField extracts a field from data using a dot notation path
func extractField(data interface{}, field string) (interface{}, error) {
	parts := strings.Split(field, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("field not found: %s", part)
			}
		case []interface{}:
			// If the part is a number, use it as an index
			if part == "[]" {
				// Return the entire array
				return v, nil
			}
			return nil, fmt.Errorf("array indexing not supported: %s", part)
		default:
			return nil, fmt.Errorf("cannot access field %s of %T", part, current)
		}
	}

	return current, nil
}

// FilterData filters data using a JQ-like syntax
func FilterData(data interface{}, filter string) (interface{}, error) {
	if filter == "" {
		return data, nil
	}

	// Convert the data to JSON and back to a map to ensure a consistent structure
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	var mapData interface{}
	if err := json.Unmarshal(jsonData, &mapData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// For now, just support simple field extraction
	return extractField(mapData, filter)
}

// PrettyPrint prints data in a pretty format
func PrettyPrint(data interface{}, w io.Writer) error {
	// Convert the data to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Write the JSON to the writer
	_, err = w.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}
