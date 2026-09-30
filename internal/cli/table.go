package cli

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

func (a *application) newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List tables",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}
			result, err := client.ListTables(command.Context(), tables.ListTablesOptions{})
			if err != nil {
				return err
			}
			return a.write(result)
		},
	}
}

func (a *application) newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <table-id>",
		Short: "Get a table",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			id, err := parseID("table ID", args[0])
			if err != nil {
				return err
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			result, err := client.GetTable(command.Context(), id)
			if err != nil {
				return err
			}
			return a.write(result)
		},
	}
}

func (a *application) newCreateCommand() *cobra.Command {
	var options tables.CreateTableOptions
	command := &cobra.Command{
		Use:   "create",
		Short: "Create a table",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}
			result, err := client.CreateTable(command.Context(), options)
			if err != nil {
				return err
			}
			return a.write(result)
		},
	}
	flags := command.Flags()
	flags.StringVar(&options.Title, "title", "", "Table title")
	flags.StringVar(&options.Emoji, "emoji", "", "Table emoji")
	flags.StringVar(&options.Description, "description", "", "Table description")
	flags.StringVar(&options.Template, "table-template", "custom", "Table template")
	_ = command.MarkFlagRequired("title")
	return command
}

func (a *application) newUpdateCommand() *cobra.Command {
	var title, emoji, description string
	var archived bool
	command := &cobra.Command{
		Use:   "update <table-id>",
		Short: "Update a table",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			id, err := parseID("table ID", args[0])
			if err != nil {
				return err
			}
			options := tables.UpdateTableOptions{}
			flags := command.Flags()
			if flags.Changed("title") {
				options.Title = &title
			}
			if flags.Changed("emoji") {
				options.Emoji = &emoji
			}
			if flags.Changed("description") {
				options.Description = &description
			}
			if flags.Changed("archived") {
				options.Archived = &archived
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			result, err := client.UpdateTable(command.Context(), id, options)
			if err != nil {
				return err
			}
			return a.write(result)
		},
	}
	flags := command.Flags()
	flags.StringVar(&title, "title", "", "Table title")
	flags.StringVar(&emoji, "emoji", "", "Table emoji")
	flags.StringVar(&description, "description", "", "Table description")
	flags.BoolVar(&archived, "archived", false, "Archive the table")
	return command
}

func (a *application) newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <table-id>",
		Short: "Delete a table",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			id, err := parseID("table ID", args[0])
			if err != nil {
				return err
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			return client.DeleteTable(command.Context(), id)
		},
	}
}

func parseID(name, raw string) (int64, error) {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, &tables.Error{ErrorCode: tables.CodeInvalidArgument, Message: name + " must be a positive integer", Cause: err}
	}
	return value, nil
}
