package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func marshalSelectedJSON(value any, fields []string) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal output: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode output: %w", err)
	}

	selected, err := selectJSONFields(decoded, fields)
	if err != nil {
		return nil, err
	}
	result, err := json.Marshal(selected)
	if err != nil {
		return nil, fmt.Errorf("marshal selected output: %w", err)
	}
	return result, nil
}

func selectJSONFields(value any, fields []string) (any, error) {
	switch typed := value.(type) {
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			selected, err := selectJSONObject(item, fields)
			if err != nil {
				return nil, err
			}
			result = append(result, selected)
		}
		return result, nil
	default:
		return selectJSONObject(value, fields)
	}
}

func selectJSONObject(value any, fields []string) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("JSON field selection requires an object or array of objects")
	}
	result := make(map[string]any, len(fields))
	for _, rawField := range fields {
		field := strings.TrimSpace(rawField)
		if field == "" {
			return nil, fmt.Errorf("JSON field name must not be empty")
		}
		actual, found := findField(object, field)
		if !found {
			return nil, fmt.Errorf("unknown JSON field %q", field)
		}
		result[field] = object[actual]
	}
	return result, nil
}

func findField(object map[string]any, requested string) (string, bool) {
	if _, ok := object[requested]; ok {
		return requested, true
	}
	for key := range object {
		if strings.EqualFold(key, requested) {
			return key, true
		}
	}
	return "", false
}

func writeJSON(writer io.Writer, data []byte) error {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("decode JSON output: %w", err)
	}
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("write JSON output: %w", err)
	}
	return nil
}
