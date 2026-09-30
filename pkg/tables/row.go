package tables

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Row describes a Tables row.
type Row struct {
	ID         int64  `json:"id"`
	TableID    int64  `json:"tableId,omitempty"`
	CreatedBy  string `json:"createdBy,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
	LastEditBy string `json:"lastEditBy,omitempty"`
	LastEditAt string `json:"lastEditAt,omitempty"`
	Data       []Cell `json:"data,omitempty"`
}

// Cell is one row value associated with a column.
type Cell struct {
	ColumnID int64 `json:"columnId"`
	Value    any   `json:"value"`
}

// RowValues maps column IDs to values.
type RowValues map[int64]any

// ListRowsOptions configures row listing.
type ListRowsOptions struct {
	Limit  int
	Offset int
}

// CreateRowOptions contains values for a new row.
type CreateRowOptions struct {
	Values RowValues
}

// UpdateRowOptions contains row values to update.
type UpdateRowOptions struct {
	Values RowValues
}

// Rows exposes row CRUD for either an authenticated table or a public share.
type Rows struct {
	client  *Client
	tableID int64
	share   *ShareScope
}

// List returns rows from the scope.
func (r Rows) List(ctx context.Context, options ListRowsOptions) ([]Row, error) {
	if options.Limit < 0 || options.Offset < 0 {
		return nil, &Error{ErrorCode: CodeInvalidArgument, Message: "row limit and offset must not be negative"}
	}
	query := make(url.Values)
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	if options.Offset > 0 {
		query.Set("offset", strconv.Itoa(options.Offset))
	}
	if r.client == nil {
		return nil, &Error{ErrorCode: CodeInvalidConfiguration, Message: "row scope has no client"}
	}
	var result []Row
	if r.share == nil {
		if err := validateID("table ID", r.tableID); err != nil {
			return nil, err
		}
		apiPath := fmt.Sprintf("/index.php/apps/tables/api/1/tables/%d/rows", r.tableID)
		if err := r.client.doJSON(ctx, r.client.httpClient, r.client.auth, http.MethodGet, apiPath, query, nil, &result); err != nil {
			return nil, err
		}
		return result, nil
	}
	apiPath, httpClient, auth, err := r.endpoint(ctx, 0)
	if err != nil {
		return nil, err
	}
	if err := r.client.doOCS(ctx, httpClient, auth, http.MethodGet, apiPath, query, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Create creates a row in the scope.
func (r Rows) Create(ctx context.Context, options CreateRowOptions) (*Row, error) {
	if len(options.Values) == 0 {
		return nil, &Error{ErrorCode: CodeInvalidArgument, Message: "row values must not be empty"}
	}
	apiPath, httpClient, auth, err := r.endpoint(ctx, 0)
	if err != nil {
		return nil, err
	}
	var result Row
	if err := r.client.doOCS(ctx, httpClient, auth, http.MethodPost, apiPath, nil, map[string]any{"data": options.Values}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates a row in the scope.
func (r Rows) Update(ctx context.Context, rowID int64, options UpdateRowOptions) (*Row, error) {
	if err := validateID("row ID", rowID); err != nil {
		return nil, err
	}
	if len(options.Values) == 0 {
		return nil, &Error{ErrorCode: CodeInvalidArgument, Message: "row values must not be empty"}
	}
	apiPath, httpClient, auth, err := r.endpoint(ctx, rowID)
	if err != nil {
		return nil, err
	}
	var result Row
	if err := r.client.doOCS(ctx, httpClient, auth, http.MethodPut, apiPath, nil, map[string]any{"data": options.Values}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a row in the scope.
func (r Rows) Delete(ctx context.Context, rowID int64) error {
	if err := validateID("row ID", rowID); err != nil {
		return err
	}
	apiPath, httpClient, auth, err := r.endpoint(ctx, rowID)
	if err != nil {
		return err
	}
	return r.client.doOCS(ctx, httpClient, auth, http.MethodDelete, apiPath, nil, nil, nil)
}

// Value returns a value for columnID or nil when the row has no such cell.
func (r Row) Value(columnID int64) any {
	for _, cell := range r.Data {
		if cell.ColumnID == columnID {
			return cell.Value
		}
	}
	return nil
}

// endpoint resolves the row endpoint, transport, and authentication for this scope.
func (r Rows) endpoint(ctx context.Context, rowID int64) (string, *http.Client, Auth, error) {
	if r.client == nil {
		return "", nil, nil, &Error{ErrorCode: CodeInvalidConfiguration, Message: "row scope has no client"}
	}
	if r.share != nil {
		if err := r.share.ensureAuthenticated(ctx); err != nil {
			return "", nil, nil, err
		}
		base := fmt.Sprintf("/ocs/v2.php/apps/tables/api/2/public/%s/rows", url.PathEscape(string(r.share.share.Token)))
		if rowID > 0 {
			base += fmt.Sprintf("/%d", rowID)
		}
		return base, r.share.httpClient, nil, nil
	}
	if err := validateID("table ID", r.tableID); err != nil {
		return "", nil, nil, err
	}
	if rowID == 0 {
		return fmt.Sprintf("/ocs/v2.php/apps/tables/api/2/tables/%d/rows", r.tableID), r.client.httpClient, r.client.auth, nil
	}
	return fmt.Sprintf("/ocs/v2.php/apps/tables/api/2/tables/%d/rows/%d", r.tableID, rowID), r.client.httpClient, r.client.auth, nil
}
