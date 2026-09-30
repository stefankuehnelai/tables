package output

import (
	"bytes"
	"fmt"
	"io"

	ghtemplate "github.com/cli/go-gh/v2/pkg/template"
)

func writeTemplate(writer io.Writer, data []byte, expression string) error {
	tmpl := ghtemplate.New(writer, 80, false)
	if err := tmpl.Parse(expression); err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	if err := tmpl.Execute(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	if err := tmpl.Flush(); err != nil {
		return fmt.Errorf("flush template: %w", err)
	}
	return nil
}
