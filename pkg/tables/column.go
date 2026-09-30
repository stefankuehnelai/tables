package tables

import (
	"context"
	"fmt"
	"net/http"
)

// Column is a Nextcloud Tables column.
type Column struct {
	ID            int64  `json:"id"`
	UUID          string `json:"uuid,omitempty"`
	Title         string `json:"title"`
	TechnicalName string `json:"technicalName,omitempty"`
	TableID       int64  `json:"tableId"`
	Type          string `json:"type"`
	Subtype       string `json:"subtype,omitempty"`
	Mandatory     bool   `json:"mandatory"`
	Description   string `json:"description,omitempty"`
}

// ListColumnsOptions contains future-compatible list controls.
type ListColumnsOptions struct{}

// CreateColumnOptions defines a supported generic column creation payload.
type CreateColumnOptions struct {
	Title                    string
	TechnicalName            string
	Type                     string
	Subtype                  string
	Mandatory                bool
	Description              string
	NumberDefault            *float64
	NumberMin                *float64
	NumberMax                *float64
	NumberDecimals           *int
	NumberPrefix             string
	NumberSuffix             string
	TextDefault              string
	TextAllowedPattern       string
	TextMaxLength            *int
	TextUnique               *bool
	SelectionOptions         string
	SelectionDefault         string
	DatetimeDefault          string
	UsergroupDefault         string
	UsergroupMultipleItems   *bool
	UsergroupSelectUsers     *bool
	UsergroupSelectGroups    *bool
	UsergroupSelectTeams     *bool
	UsergroupShowUserStatus  *bool
}

// UpdateColumnOptions defines mutable column fields.
type UpdateColumnOptions struct {
	Title                    *string
	TechnicalName            *string
	Subtype                  *string
	Mandatory                *bool
	Description              *string
	NumberDefault            *float64
	NumberMin                *float64
	NumberMax                *float64
	NumberDecimals           *int
	NumberPrefix             *string
	NumberSuffix             *string
	TextDefault              *string
	TextAllowedPattern       *string
	TextMaxLength            *int
	TextUnique               *bool
	SelectionOptions         *string
	SelectionDefault         *string
	DatetimeDefault          *string
	UsergroupDefault         *string
	UsergroupMultipleItems   *bool
	UsergroupSelectUsers     *bool
	UsergroupSelectGroups    *bool
	UsergroupSelectTeams     *bool
	UsergroupShowUserStatus  *bool
}

// Columns is a column API scoped to one table.
type Columns struct {
	client  *Client
	tableID int64
}

// List returns all columns in the scoped table.
func (c Columns) List(ctx context.Context, _ ListColumnsOptions) ([]Column, error) {
	if err := tableID(c.tableID, "table ID"); err != nil {
		return nil, err
	}
	var result []Column
	if err := c.client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/tables/%d/columns", c.tableID), nil, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Create creates a column using the API v1 generic column endpoint.
func (c Columns) Create(ctx context.Context, options CreateColumnOptions) (Column, error) {
	if err := tableID(c.tableID, "table ID"); err != nil {
		return Column{}, err
	}
	if options.Title == "" || options.Type == "" {
		return Column{}, NewError(CodeInvalidArgument, "column title and type must not be empty")
	}
	payload := createColumnPayload(c.tableID, options)
	var result Column
	err := c.client.doJSON(ctx, http.MethodPost, "/columns", nil, payload, &result)
	return result, err
}

// Update updates a column. tableID remains part of the scope to reject mismatched CLI targeting early.
func (c Columns) Update(ctx context.Context, columnID int64, options UpdateColumnOptions) (Column, error) {
	if err := tableID(c.tableID, "table ID"); err != nil {
		return Column{}, err
	}
	if err := tableID(columnID, "column ID"); err != nil {
		return Column{}, err
	}
	var result Column
	err := c.client.doJSON(ctx, http.MethodPut, fmt.Sprintf("/columns/%d", columnID), nil, updateColumnPayload(options), &result)
	return result, err
}

// Delete deletes a column and returns the deleted representation.
func (c Columns) Delete(ctx context.Context, columnID int64) (Column, error) {
	if err := tableID(c.tableID, "table ID"); err != nil {
		return Column{}, err
	}
	if err := tableID(columnID, "column ID"); err != nil {
		return Column{}, err
	}
	var result Column
	err := c.client.doJSON(ctx, http.MethodDelete, fmt.Sprintf("/columns/%d", columnID), nil, nil, &result)
	return result, err
}

// createColumnPayload supplies nullable API v1 parameters explicitly for compatibility.
func createColumnPayload(tableID int64, o CreateColumnOptions) map[string]any {
	return map[string]any{
		"tableId": tableID, "viewId": nil, "title": o.Title, "technicalName": nullableString(o.TechnicalName),
		"type": o.Type, "subtype": nullableString(o.Subtype), "mandatory": o.Mandatory, "description": nullableString(o.Description),
		"numberPrefix": nullableString(o.NumberPrefix), "numberSuffix": nullableString(o.NumberSuffix), "numberDefault": o.NumberDefault,
		"numberMin": o.NumberMin, "numberMax": o.NumberMax, "numberDecimals": o.NumberDecimals,
		"textDefault": nullableString(o.TextDefault), "textAllowedPattern": nullableString(o.TextAllowedPattern), "textMaxLength": o.TextMaxLength,
		"textUnique": o.TextUnique, "selectionOptions": o.SelectionOptions, "selectionDefault": o.SelectionDefault,
		"datetimeDefault": o.DatetimeDefault, "usergroupDefault": o.UsergroupDefault,
		"usergroupMultipleItems": o.UsergroupMultipleItems, "usergroupSelectUsers": o.UsergroupSelectUsers,
		"usergroupSelectGroups": o.UsergroupSelectGroups, "usergroupSelectTeams": o.UsergroupSelectTeams,
		"usergroupShowUserStatus": o.UsergroupShowUserStatus, "selectedViewIds": []int64{}, "customSettings": map[string]any{},
	}
}

// updateColumnPayload supplies all nullable update parameters explicitly.
func updateColumnPayload(o UpdateColumnOptions) map[string]any {
	return map[string]any{
		"title": o.Title, "technicalName": o.TechnicalName, "subtype": o.Subtype, "mandatory": o.Mandatory, "description": o.Description,
		"numberPrefix": o.NumberPrefix, "numberSuffix": o.NumberSuffix, "numberDefault": o.NumberDefault, "numberMin": o.NumberMin,
		"numberMax": o.NumberMax, "numberDecimals": o.NumberDecimals, "textDefault": o.TextDefault,
		"textAllowedPattern": o.TextAllowedPattern, "textMaxLength": o.TextMaxLength, "textUnique": o.TextUnique,
		"selectionOptions": o.SelectionOptions, "selectionDefault": o.SelectionDefault, "datetimeDefault": o.DatetimeDefault,
		"usergroupDefault": o.UsergroupDefault, "usergroupMultipleItems": o.UsergroupMultipleItems,
		"usergroupSelectUsers": o.UsergroupSelectUsers, "usergroupSelectGroups": o.UsergroupSelectGroups,
		"usergroupSelectTeams": o.UsergroupSelectTeams, "usergroupShowUserStatus": o.UsergroupShowUserStatus,
		"customSettings": map[string]any{},
	}
}

// nullableString converts an empty optional string into JSON null.
func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
