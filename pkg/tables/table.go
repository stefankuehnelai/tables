package tables

import (
	"context"
	"fmt"
	"net/http"
)

// Table describes a Nextcloud Tables table.
type Table struct {
	ID           int64  `json:"id"`
	UUID         string `json:"uuid,omitempty"`
	Title        string `json:"title"`
	Emoji        string `json:"emoji,omitempty"`
	Description  string `json:"description,omitempty"`
	Archived     bool   `json:"archived,omitempty"`
	Favorite     bool   `json:"favorite,omitempty"`
	RowsCount    int    `json:"rowsCount,omitempty"`
	ColumnsCount int    `json:"columnsCount,omitempty"`
}

// ListTablesOptions configures table listing.
type ListTablesOptions struct{}

// CreateTableOptions contains table creation values.
type CreateTableOptions struct {
	Title       string `json:"title"`
	Emoji       string `json:"emoji,omitempty"`
	Description string `json:"description,omitempty"`
	Template    string `json:"template,omitempty"`
}

// UpdateTableOptions contains table fields to update.
type UpdateTableOptions struct {
	Title       *string `json:"title,omitempty"`
	Emoji       *string `json:"emoji,omitempty"`
	Description *string `json:"description,omitempty"`
	Archived    *bool   `json:"archived,omitempty"`
}

// TableScope identifies one table and exposes scoped resources.
type TableScope struct {
	client *Client
	id     int64
}

// Table returns an immutable scope for id.
func (c *Client) Table(id int64) TableScope {
	return TableScope{client: c, id: id}
}

// ID returns the table identifier.
func (s TableScope) ID() int64 {
	return s.id
}

// Rows returns row operations scoped to the table.
func (s TableScope) Rows() Rows {
	return Rows{client: s.client, tableID: s.id}
}

// Columns returns column operations scoped to the table.
func (s TableScope) Columns() Columns {
	return Columns{client: s.client, tableID: s.id}
}

// ListTables returns all accessible tables.
func (c *Client) ListTables(ctx context.Context, _ ListTablesOptions) ([]Table, error) {
	var result []Table
	if err := c.doOCS(ctx, c.httpClient, c.auth, http.MethodGet, "/ocs/v2.php/apps/tables/api/2/tables", nil, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetTable returns one table.
func (c *Client) GetTable(ctx context.Context, id int64) (*Table, error) {
	if err := validateID("table ID", id); err != nil {
		return nil, err
	}
	var result Table
	apiPath := fmt.Sprintf("/ocs/v2.php/apps/tables/api/2/tables/%d", id)
	if err := c.doOCS(ctx, c.httpClient, c.auth, http.MethodGet, apiPath, nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateTable creates a table.
func (c *Client) CreateTable(ctx context.Context, options CreateTableOptions) (*Table, error) {
	if options.Title == "" {
		return nil, &Error{ErrorCode: CodeInvalidArgument, Message: "table title must not be empty"}
	}
	if options.Template == "" {
		options.Template = "custom"
	}
	var result Table
	if err := c.doOCS(ctx, c.httpClient, c.auth, http.MethodPost, "/ocs/v2.php/apps/tables/api/2/tables", nil, options, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateTable updates a table.
func (c *Client) UpdateTable(ctx context.Context, id int64, options UpdateTableOptions) (*Table, error) {
	if err := validateID("table ID", id); err != nil {
		return nil, err
	}
	var result Table
	apiPath := fmt.Sprintf("/ocs/v2.php/apps/tables/api/2/tables/%d", id)
	if err := c.doOCS(ctx, c.httpClient, c.auth, http.MethodPut, apiPath, nil, options, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteTable deletes a table.
func (c *Client) DeleteTable(ctx context.Context, id int64) error {
	if err := validateID("table ID", id); err != nil {
		return err
	}
	apiPath := fmt.Sprintf("/ocs/v2.php/apps/tables/api/2/tables/%d", id)
	return c.doOCS(ctx, c.httpClient, c.auth, http.MethodDelete, apiPath, nil, nil, nil)
}
