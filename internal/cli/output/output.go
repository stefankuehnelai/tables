// Package output formats CLI results for human and structured output.
package output

import (
	"fmt"
	"io"
)

// Options configures a Formatter.
type Options struct {
	JSONFields []string
	JQ         string
	Template   string
}

// Formatter renders values according to the requested output mode.
type Formatter struct {
	options Options
}

// New constructs a Formatter.
func New(options Options) (*Formatter, error) {
	if options.JQ != "" && len(options.JSONFields) == 0 {
		return nil, fmt.Errorf("--jq requires --json")
	}
	if options.Template != "" && len(options.JSONFields) == 0 {
		return nil, fmt.Errorf("--template requires --json")
	}
	if options.JQ != "" && options.Template != "" {
		return nil, fmt.Errorf("--jq and --template are mutually exclusive")
	}
	return &Formatter{options: options}, nil
}

// Write renders value to writer.
func (f *Formatter) Write(writer io.Writer, value any) error {
	if len(f.options.JSONFields) > 0 {
		data, err := marshalSelectedJSON(value, f.options.JSONFields)
		if err != nil {
			return err
		}
		switch {
		case f.options.JQ != "":
			return writeJQ(writer, data, f.options.JQ)
		case f.options.Template != "":
			return writeTemplate(writer, data, f.options.Template)
		default:
			return writeJSON(writer, data)
		}
	}
	return writeTable(writer, value)
}
