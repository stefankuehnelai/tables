package output

import (
	"bytes"
	"fmt"
	"io"

	ghjq "github.com/cli/go-gh/v2/pkg/jq"
)

func writeJQ(writer io.Writer, data []byte, expression string) error {
	if err := ghjq.Evaluate(bytes.NewReader(data), writer, expression); err != nil {
		return fmt.Errorf("evaluate jq expression: %w", err)
	}
	return nil
}
