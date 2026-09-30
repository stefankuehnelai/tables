package tables

import (
	"context"
	"fmt"
	"net/http"
)

// Table is a Nextcloud Tables table.
type Table struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid,omitempty"`
	Title       string `json:"title"`
	Emoji       string `json:"emoji,omitempty"`
	Description string `json:"description,omitempty"`
	Archived    bool   `json:"archived,omitempty"`
	RowsCount   int64  `json:"rowsCount,omitempty"`
}

// ListTablesOptions contains future-compatible list controls.
type ListTablesOptions struct{}

// CreateTableOptions is the payload for creating a table.
type CreateTableOptions struct {
	Title       string `json:"title"`
	Emoji       string `json:"emoji,omitempty"`
	Description string `json:"description,omitempty"`
	Template    string `json:"template,omitempty"`
}

// UpdateTableOptions is the payload for updating table metadata.
type UpdateTableOptions struct {
	Title       *string `json:"title,omitempty"`
	Emoji       *string `json:"emoji,omitempty"`
	Description *string `json:"description,omitempty"`
	Archived    *bool   `json:"archived,omitempty"`
}

// TableScope binds operations to one authenticated table.
type TableScope struct {
	client *Client
	id     int64
}

// Table creates an immutable scope for tableID.
func (c *Client) Table(tableID int64) TableScope {
	return TableScope{client: c, id: tableID}
}

// ListTables returns all tables visible to the authenticated account.
func (c *Client) ListTables(ctx context.Context, _ ListTablesOptions) ([]Table, error) {
	var result []Table
	if err := c.doOCS(ctx, c.httpClient, http.MethodGet, "/tables", nil, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetTable returns one table by ID.
func (c *Client) GetTable(ctx context.Context, id int64) (Table, error) {
	if err := tableID(id, "table ID"); err != nil {
		return Table{}, err
	}
	var result Table
	err := c.doOCS(ctx, c.httpClient, http.MethodGet, fmt.Sprintf("/tables/%d", id), nil, nil, &result)
	return result, err
}

// CreateTable creates a table.
func (c *Client) CreateTable(ctx context.Context, options CreateTableOptions) (Table, error) {
	if options.Title == "" {
		return Table{}, NewError(CodeInvalidArgument, "table title must not be empty")
	}
	if options.Template == "" {
		options.Template = "custom"
	}
	var result Table
	err := c.doOCS(ctx, c.httpClient, http.MethodPost, "/tables", nil, options, &result)
	return result, err
}

// UpdateTable updates table metadata.
func (c *Client) UpdateTable(ctx context.Context, id int64, options UpdateTableOptions) (Table, error) {
	if err := tableID(id, "table ID"); err != nil {
		return Table{}, err
	}
	var result Table
	err := c.doOCS(ctx, c.httpClient, http.MethodPut, fmt.Sprintf("/tables/%d", id), nil, options, &result)
	return result, err
}

// DeleteTable deletes a table and returns the deleted representation.
func (c *Client) DeleteTable(ctx context.Context, id int64) (Table, error) {
	if err := tableID(id, "table ID"); err != nil {
		return Table{}, err
	}
	var result Table
	err := c.doOCS(ctx, c.httpClient, http.MethodDelete, fmt.Sprintf("/tables/%d", id), nil, nil, &result)
	return result, err
}

// Rows returns the shared row API scoped to this table.
func (s TableScope) Rows() Rows {
	return Rows{client: s.client, httpClient: s.client.httpClient, tableID: s.id}
}

// Columns returns a column API scoped to this table.
func (s TableScope) Columns() Columns {
	return Columns{client: s.client, tableID: s.id}
}
