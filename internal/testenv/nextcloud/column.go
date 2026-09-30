package nextcloud

import (
	"context"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

// CreateColumn creates a column for integration-test arrangement.
func (n *Nextcloud) CreateColumn(ctx context.Context, username, password string, tableID int64, options tables.CreateColumnOptions) (*tables.Column, error) {
	client, err := n.Client(username, password)
	if err != nil {
		return nil, err
	}
	return client.Table(tableID).Columns().Create(ctx, options)
}
