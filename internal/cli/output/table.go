package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/cli/go-gh/v2/pkg/tableprinter"
)

func writeTable(writer io.Writer, value any) error {
	rows, columns, err := normalizedRows(value)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	printer := tableprinter.New(writer, true, 120)
	for _, column := range columns {
		printer.AddField(strings.ToUpper(column))
	}
	printer.EndRow()
	for _, row := range rows {
		for _, column := range columns {
			printer.AddField(displayValue(row[column]))
		}
		printer.EndRow()
	}
	if err := printer.Render(); err != nil {
		return fmt.Errorf("render table: %w", err)
	}
	return nil
}

func normalizedRows(value any) ([]map[string]any, []string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal table output: %w", err)
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, nil, fmt.Errorf("decode table output: %w", err)
	}
	var items []any
	switch typed := decoded.(type) {
	case []any:
		items = typed
	case map[string]any:
		items = []any{typed}
	case nil:
		return nil, nil, nil
	default:
		return nil, nil, fmt.Errorf("table output requires an object or array of objects, got %s", reflect.TypeOf(decoded))
	}
	rows := make([]map[string]any, 0, len(items))
	columnSet := map[string]struct{}{}
	for _, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			return nil, nil, fmt.Errorf("table output requires objects")
		}
		rows = append(rows, object)
		for key := range object {
			if isScalar(object[key]) {
				columnSet[key] = struct{}{}
			}
		}
	}
	columns := make([]string, 0, len(columnSet))
	for key := range columnSet {
		columns = append(columns, key)
	}
	sort.Slice(columns, func(i, j int) bool {
		rank := func(key string) int {
			switch strings.ToLower(key) {
			case "id":
				return 0
			case "title", "name":
				return 1
			default:
				return 2
			}
		}
		ri, rj := rank(columns[i]), rank(columns[j])
		if ri != rj {
			return ri < rj
		}
		return columns[i] < columns[j]
	})
	return rows, columns, nil
}

func isScalar(value any) bool {
	switch value.(type) {
	case nil, string, float64, bool:
		return true
	default:
		return false
	}
}

func displayValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
