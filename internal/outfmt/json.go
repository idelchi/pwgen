package outfmt

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/idelchi/pwgen/internal/generate"
)

// JSONFormatter formats output as JSON.
type JSONFormatter struct {
	writer io.Writer
	pretty bool
}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter(writer io.Writer, pretty bool) *JSONFormatter {
	return &JSONFormatter{
		writer: writer,
		pretty: pretty,
	}
}

// FormatResults formats generation results as JSON.
func (f *JSONFormatter) FormatResults(results []generate.Result) error {
	var data any

	if len(results) == 1 {
		data = results[0]
	} else {
		data = results
	}

	var (
		output []byte
		err    error
	)

	if f.pretty {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	_, err = f.writer.Write(output)
	if err != nil {
		return fmt.Errorf("writing JSON: %w", err)
	}

	// Add newline for better terminal output
	_, err = f.writer.Write([]byte("\n"))

	return err
}

// FormatAnalysis formats entropy analysis as JSON.
func (f *JSONFormatter) FormatAnalysis(analysis generate.AnalysisResult) error {
	var (
		output []byte
		err    error
	)

	if f.pretty {
		output, err = json.MarshalIndent(analysis, "", "  ")
	} else {
		output, err = json.Marshal(analysis)
	}

	if err != nil {
		return fmt.Errorf("marshaling analysis JSON: %w", err)
	}

	_, err = f.writer.Write(output)
	if err != nil {
		return fmt.Errorf("writing analysis JSON: %w", err)
	}

	// Add newline
	_, err = f.writer.Write([]byte("\n"))

	return err
}

// FormatDictionaries formats dictionary information as JSON.
func (f *JSONFormatter) FormatDictionaries(dicts []DictionaryInfo) error {
	var (
		output []byte
		err    error
	)

	if f.pretty {
		output, err = json.MarshalIndent(dicts, "", "  ")
	} else {
		output, err = json.Marshal(dicts)
	}

	if err != nil {
		return fmt.Errorf("marshaling dictionaries JSON: %w", err)
	}

	_, err = f.writer.Write(output)
	if err != nil {
		return fmt.Errorf("writing dictionaries JSON: %w", err)
	}

	// Add newline
	_, err = f.writer.Write([]byte("\n"))

	return err
}
