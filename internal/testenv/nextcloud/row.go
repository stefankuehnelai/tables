package nextcloud

import (
	"context"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

// CreateRow creates a row for integration-test arrangement.
func (n *Nextcloud) CreateRow(ctx context.Context, username, password string, tableID int64, values tables.RowValues) (tables.Row, error) {
	client, err := n.Client(username, password)
	if err != nil {
		return tables.Row{}, err
	}
	return client.Table(tableID).Rows().Create(ctx, tables.CreateRowOptions{Values: values})
}
