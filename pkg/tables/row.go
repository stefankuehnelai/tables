package tables

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
)

// Cell is one column value within a row.
type Cell struct {
	ColumnID int64 `json:"columnId"`
	Value    any   `json:"value"`
}

// Row is a Nextcloud Tables row.
type Row struct {
	ID          int64           `json:"id"`
	TableID     int64           `json:"tableId,omitempty"`
	CreatedBy   string          `json:"createdBy,omitempty"`
	CreatedAt   string          `json:"createdAt,omitempty"`
	LastEditBy  string          `json:"lastEditBy,omitempty"`
	LastEditAt  string          `json:"lastEditAt,omitempty"`
	Data        []Cell          `json:"data"`
	DataByAlias map[string]Cell `json:"dataByAlias,omitempty"`
}

// Value returns a cell value and whether the requested column exists.
func (r Row) Value(columnID int64) (any, bool) {
	for _, cell := range r.Data {
		if cell.ColumnID == columnID {
			return cell.Value, true
		}
	}
	return nil, false
}

// RowValues maps column identifiers to values for row mutations.
type RowValues map[int64]any

// ListRowsOptions controls pagination.
type ListRowsOptions struct {
	Limit  int
	Offset int
}

// CreateRowOptions is the payload for creating a row.
type CreateRowOptions struct {
	Values RowValues
}

// UpdateRowOptions is the payload for updating a row.
type UpdateRowOptions struct {
	Values RowValues
}

// Rows is the common row CRUD API used by table and public-share scopes.
type Rows struct {
	client     *Client
	httpClient *http.Client
	tableID    int64
	share      *shareState
}

// List returns rows in the current table or share scope.
func (r Rows) List(ctx context.Context, options ListRowsOptions) ([]Row, error) {
	if options.Limit < 0 || options.Offset < 0 {
		return nil, NewError(CodeInvalidArgument, "row limit and offset must not be negative")
	}
	if err := r.prepare(ctx); err != nil {
		return nil, err
	}
	query := url.Values{}
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	if options.Offset > 0 {
		query.Set("offset", strconv.Itoa(options.Offset))
	}
	var result []Row
	if err := r.client.doOCS(ctx, r.httpClient, http.MethodGet, r.route(""), query, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Create creates a row in the current scope.
func (r Rows) Create(ctx context.Context, options CreateRowOptions) (Row, error) {
	if err := validateRowValues(options.Values); err != nil {
		return Row{}, err
	}
	if err := r.prepare(ctx); err != nil {
		return Row{}, err
	}
	var result Row
	err := r.client.doOCS(ctx, r.httpClient, http.MethodPost, r.route(""), nil, rowPayload(options.Values), &result)
	return result, err
}

// Update updates a row in the current scope.
func (r Rows) Update(ctx context.Context, rowID int64, options UpdateRowOptions) (Row, error) {
	if err := tableID(rowID, "row ID"); err != nil {
		return Row{}, err
	}
	if err := validateRowValues(options.Values); err != nil {
		return Row{}, err
	}
	if err := r.prepare(ctx); err != nil {
		return Row{}, err
	}
	var result Row
	err := r.client.doOCS(ctx, r.httpClient, http.MethodPut, r.route(fmt.Sprintf("/%d", rowID)), nil, rowPayload(options.Values), &result)
	return result, err
}

// Delete deletes a row in the current scope.
func (r Rows) Delete(ctx context.Context, rowID int64) (Row, error) {
	if err := tableID(rowID, "row ID"); err != nil {
		return Row{}, err
	}
	if err := r.prepare(ctx); err != nil {
		return Row{}, err
	}
	var result Row
	err := r.client.doOCS(ctx, r.httpClient, http.MethodDelete, r.route(fmt.Sprintf("/%d", rowID)), nil, nil, &result)
	return result, err
}

// route resolves the row collection for authenticated or public-share use.
func (r Rows) route(suffix string) string {
	if r.share != nil {
		return "/public/" + url.PathEscape(string(r.share.share.Token)) + "/rows" + suffix
	}
	return fmt.Sprintf("/tables/%d/rows%s", r.tableID, suffix)
}

// prepare validates the scope and authenticates a protected share once.
func (r Rows) prepare(ctx context.Context) error {
	if r.share != nil {
		return r.share.authenticate(ctx, r.client)
	}
	return tableID(r.tableID, "table ID")
}

// rowPayload converts the map representation into the API's deterministic cell list.
func rowPayload(values RowValues) map[string]any {
	ids := make([]int64, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i int, j int) bool { return ids[i] < ids[j] })
	cells := make([]Cell, 0, len(ids))
	for _, id := range ids {
		cells = append(cells, Cell{ColumnID: id, Value: values[id]})
	}
	return map[string]any{"data": cells}
}

// validateRowValues ensures mutation payloads are nonempty and use valid column IDs.
func validateRowValues(values RowValues) error {
	if len(values) == 0 {
		return NewError(CodeInvalidArgument, "row values must not be empty")
	}
	for id := range values {
		if id <= 0 {
			return NewError(CodeInvalidArgument, "row column IDs must be positive")
		}
	}
	return nil
}
