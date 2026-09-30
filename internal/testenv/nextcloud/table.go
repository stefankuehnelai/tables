package nextcloud

import (
	"context"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

// CreateTable creates a table for integration-test arrangement.
func (n *Nextcloud) CreateTable(ctx context.Context, username, password, title string) (tables.Table, error) {
	client, err := n.Client(username, password)
	if err != nil {
		return tables.Table{}, err
	}
	return client.CreateTable(ctx, tables.CreateTableOptions{Title: title})
}
