package cli

import (
	"github.com/spf13/cobra"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

type rowTargetOptions struct {
	shareToken    string
	sharePassword string
}

func (a *application) newRowsCommand() *cobra.Command {
	command := &cobra.Command{Use: "rows", Short: "Work with table rows"}
	command.AddCommand(a.newRowsListCommand(), a.newRowsCreateCommand(), a.newRowsUpdateCommand(), a.newRowsDeleteCommand())
	return command
}

func (a *application) addShareFlags(command *cobra.Command, options *rowTargetOptions) {
	command.Flags().StringVar(&options.shareToken, "share-token", "", "Public share token")
	command.Flags().StringVar(&options.sharePassword, "share-password", "", "Public share password")
}

func (a *application) rowScope(args []string, options rowTargetOptions, rowIDPosition bool) (tables.Rows, int64, error) {
	token := firstNonEmpty(options.shareToken, a.deps.getenv(envShareToken))
	password := firstNonEmpty(options.sharePassword, a.deps.getenv(envSharePassword))
	if token != "" {
		expected := 0
		if rowIDPosition {
			expected = 1
		}
		if len(args) != expected {
			return tables.Rows{}, 0, &tables.Error{ErrorCode: tables.CodeInvalidArgument, Message: "do not combine a table ID with --share-token"}
		}
		client, err := a.shareClient()
		if err != nil {
			return tables.Rows{}, 0, err
		}
		rows := client.Share(tables.Share{Token: tables.ShareToken(token), Password: tables.SharePassword(password)}).Rows()
		if !rowIDPosition {
			return rows, 0, nil
		}
		rowID, err := parseID("row ID", args[0])
		return rows, rowID, err
	}

	expected := 1
	if rowIDPosition {
		expected = 2
	}
	if len(args) != expected {
		return tables.Rows{}, 0, &tables.Error{ErrorCode: tables.CodeInvalidArgument, Message: "a table ID is required when --share-token is not set"}
	}
	tableID, err := parseID("table ID", args[0])
	if err != nil {
		return tables.Rows{}, 0, err
	}
	client, err := a.client()
	if err != nil {
		return tables.Rows{}, 0, err
	}
	rows := client.Table(tableID).Rows()
	if !rowIDPosition {
		return rows, 0, nil
	}
	rowID, err := parseID("row ID", args[1])
	return rows, rowID, err
}

func (a *application) shareClient() (*tables.Client, error) {
	connection, err := a.resolveConnectionWithoutCredential()
	if err != nil {
		return nil, err
	}
	return tables.NewClient(connection.server)
}

func (a *application) resolveConnectionWithoutCredential() (connection, error) {
	cfg, err := a.deps.config.Load()
	if err != nil {
		return connection{}, err
	}
	hostname := firstNonEmpty(a.root.hostname, a.deps.getenv(envHostname), cfg.ActiveHostname)
	server := firstNonEmpty(a.root.server, a.deps.getenv(envServer))
	if server == "" && hostname != "" {
		if host, ok := cfg.Hosts[hostname]; ok {
			server = host.Server
		}
		if server == "" {
			server = "https://" + hostname
		}
	}
	if server == "" {
		return connection{}, &tables.Error{ErrorCode: tables.CodeInvalidConfiguration, Message: "no Nextcloud server configured for public share"}
	}
	if err := validateServer(server); err != nil {
		return connection{}, err
	}
	return connection{hostname: hostname, server: server}, nil
}

func (a *application) newRowsListCommand() *cobra.Command {
	target := rowTargetOptions{}
	var limit, offset int
	command := &cobra.Command{
		Use: "list [table-id]", Short: "List rows", Args: cobra.ArbitraryArgs,
		RunE: func(command *cobra.Command, args []string) error {
			rows, _, err := a.rowScope(args, target, false)
			if err != nil {
				return err
			}
			result, err := rows.List(command.Context(), tables.ListRowsOptions{Limit: limit, Offset: offset})
			if err != nil {
				return err
			}
			return a.write(result)
		},
	}
	a.addShareFlags(command, &target)
	command.Flags().IntVar(&limit, "limit", 0, "Maximum rows to return")
	command.Flags().IntVar(&offset, "offset", 0, "Rows to skip")
	return command
}

func (a *application) newRowsCreateCommand() *cobra.Command {
	target := rowTargetOptions{}
	var values []string
	command := &cobra.Command{
		Use: "create [table-id]", Short: "Create a row", Args: cobra.ArbitraryArgs,
		RunE: func(command *cobra.Command, args []string) error {
			parsed, err := parseRowValues(values)
			if err != nil {
				return err
			}
			rows, _, err := a.rowScope(args, target, false)
			if err != nil {
				return err
			}
			result, err := rows.Create(command.Context(), tables.CreateRowOptions{Values: parsed})
			if err != nil {
				return err
			}
			return a.write(result)
		},
	}
	a.addShareFlags(command, &target)
	command.Flags().StringArrayVar(&values, "value", nil, "Row value as COLUMN_ID=VALUE")
	return command
}

func (a *application) newRowsUpdateCommand() *cobra.Command {
	target := rowTargetOptions{}
	var values []string
	command := &cobra.Command{
		Use: "update [table-id] <row-id>", Short: "Update a row", Args: cobra.ArbitraryArgs,
		RunE: func(command *cobra.Command, args []string) error {
			parsed, err := parseRowValues(values)
			if err != nil {
				return err
			}
			rows, rowID, err := a.rowScope(args, target, true)
			if err != nil {
				return err
			}
			result, err := rows.Update(command.Context(), rowID, tables.UpdateRowOptions{Values: parsed})
			if err != nil {
				return err
			}
			return a.write(result)
		},
	}
	a.addShareFlags(command, &target)
	command.Flags().StringArrayVar(&values, "value", nil, "Row value as COLUMN_ID=VALUE")
	return command
}

func (a *application) newRowsDeleteCommand() *cobra.Command {
	target := rowTargetOptions{}
	command := &cobra.Command{
		Use: "delete [table-id] <row-id>", Short: "Delete a row", Args: cobra.ArbitraryArgs,
		RunE: func(command *cobra.Command, args []string) error {
			rows, rowID, err := a.rowScope(args, target, true)
			if err != nil {
				return err
			}
			return rows.Delete(command.Context(), rowID)
		},
	}
	a.addShareFlags(command, &target)
	return command
}
