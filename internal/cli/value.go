package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

func parseRowValues(values []string) (tables.RowValues, error) {
	result := tables.RowValues{}
	for _, raw := range values {
		parts := strings.SplitN(raw, "=", 2)
		if len(parts) != 2 {
			return nil, &tables.Error{ErrorCode: tables.CodeInvalidArgument, Message: fmt.Sprintf("invalid row value %q; expected COLUMN_ID=VALUE", raw)}
		}
		columnID, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil || columnID <= 0 {
			return nil, &tables.Error{ErrorCode: tables.CodeInvalidArgument, Message: fmt.Sprintf("invalid column ID %q", parts[0]), Cause: err}
		}
		result[columnID] = parseValue(parts[1])
	}
	return result, nil
}

func parseValue(raw string) any {
	var value any
	if json.Unmarshal([]byte(raw), &value) == nil {
		return value
	}
	return raw
}
