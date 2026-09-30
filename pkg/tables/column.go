package tables

import (
	"context"
	"fmt"
	"net/http"
)

// Column describes a Nextcloud Tables column.
type Column struct {
	ID            int64  `json:"id"`
	UUID          string `json:"uuid,omitempty"`
	Title         string `json:"title"`
	TechnicalName string `json:"technicalName,omitempty"`
	TableID       int64  `json:"tableId,omitempty"`
	Type          string `json:"type"`
	Subtype       string `json:"subtype,omitempty"`
	Mandatory     bool   `json:"mandatory,omitempty"`
	Description   string `json:"description,omitempty"`
}

// ListColumnsOptions configures column listing.
type ListColumnsOptions struct{}

// CreateColumnOptions contains generic column creation values supported by Tables API v1.
type CreateColumnOptions struct {
	Title         string `json:"title"`
	TechnicalName string `json:"technicalName,omitempty"`
	Type          string `json:"type"`
	Subtype       string `json:"subtype,omitempty"`
	Mandatory     bool   `json:"mandatory,omitempty"`
	Description   string `json:"description,omitempty"`
}

// UpdateColumnOptions contains generic column fields to update.
type UpdateColumnOptions struct {
	Title         *string `json:"title,omitempty"`
	TechnicalName *string `json:"technicalName,omitempty"`
	Subtype       *string `json:"subtype,omitempty"`
	Mandatory     *bool   `json:"mandatory,omitempty"`
	Description   *string `json:"description,omitempty"`
}

// Columns exposes table-scoped column operations.
type Columns struct {
	client  *Client
	tableID int64
}

// List lists columns for the scope table.
func (c Columns) List(ctx context.Context, _ ListColumnsOptions) ([]Column, error) {
	if err := validateID("table ID", c.tableID); err != nil {
		return nil, err
	}
	var result []Column
	apiPath := fmt.Sprintf("/ocs/v2.php/apps/tables/api/2/columns/table/%d", c.tableID)
	if err := c.client.doOCS(ctx, c.client.httpClient, c.client.auth, http.MethodGet, apiPath, nil, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Create creates a column in the scope table.
func (c Columns) Create(ctx context.Context, options CreateColumnOptions) (*Column, error) {
	if err := validateID("table ID", c.tableID); err != nil {
		return nil, err
	}
	if options.Title == "" || options.Type == "" {
		return nil, &Error{ErrorCode: CodeInvalidArgument, Message: "column title and type must not be empty"}
	}
	var result Column
	apiPath := fmt.Sprintf("/index.php/apps/tables/api/1/tables/%d/columns", c.tableID)
	if err := c.client.doJSON(ctx, c.client.httpClient, c.client.auth, http.MethodPost, apiPath, nil, legacyCreateColumnPayload(options), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates a column in the scope table.
func (c Columns) Update(ctx context.Context, columnID int64, options UpdateColumnOptions) (*Column, error) {
	if err := validateID("table ID", c.tableID); err != nil {
		return nil, err
	}
	if err := validateID("column ID", columnID); err != nil {
		return nil, err
	}
	var result Column
	apiPath := fmt.Sprintf("/index.php/apps/tables/api/1/columns/%d", columnID)
	if err := c.client.doJSON(ctx, c.client.httpClient, c.client.auth, http.MethodPut, apiPath, nil, legacyUpdateColumnPayload(options), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a column in the scope table.
func (c Columns) Delete(ctx context.Context, columnID int64) error {
	if err := validateID("table ID", c.tableID); err != nil {
		return err
	}
	if err := validateID("column ID", columnID); err != nil {
		return err
	}
	apiPath := fmt.Sprintf("/index.php/apps/tables/api/1/columns/%d", columnID)
	return c.client.doJSON(ctx, c.client.httpClient, c.client.auth, http.MethodDelete, apiPath, nil, nil, nil)
}

// legacyCreateColumnPayload supplies nullable parameters required by the v1 generic column route.
func legacyCreateColumnPayload(options CreateColumnOptions) map[string]any {
	payload := legacyColumnSettings()
	payload["title"] = options.Title
	payload["technicalName"] = nilIfEmpty(options.TechnicalName)
	payload["type"] = options.Type
	payload["subtype"] = nilIfEmpty(options.Subtype)
	payload["mandatory"] = options.Mandatory
	payload["description"] = nilIfEmpty(options.Description)
	return payload
}

// legacyUpdateColumnPayload supplies nullable parameters required by the v1 update route.
func legacyUpdateColumnPayload(options UpdateColumnOptions) map[string]any {
	payload := legacyColumnSettings()
	payload["title"] = options.Title
	payload["technicalName"] = options.TechnicalName
	payload["subtype"] = options.Subtype
	payload["mandatory"] = options.Mandatory
	payload["description"] = options.Description
	return payload
}

// legacyColumnSettings returns the type-specific fields accepted by the generic v1 API.
func legacyColumnSettings() map[string]any {
	return map[string]any{
		"numberPrefix":            nil,
		"numberSuffix":            nil,
		"numberDefault":           nil,
		"numberMin":               nil,
		"numberMax":               nil,
		"numberDecimals":          nil,
		"textDefault":             nil,
		"textAllowedPattern":      nil,
		"textMaxLength":           nil,
		"textUnique":              nil,
		"selectionOptions":        nil,
		"selectionDefault":        nil,
		"datetimeDefault":         nil,
		"usergroupDefault":        nil,
		"usergroupMultipleItems":  nil,
		"usergroupSelectUsers":    nil,
		"usergroupSelectGroups":   nil,
		"usergroupSelectTeams":    nil,
		"usergroupShowUserStatus": nil,
		"customSettings":          nil,
	}
}

// nilIfEmpty converts an empty optional string to JSON null.
func nilIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
