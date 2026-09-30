package cli

import (
	"github.com/spf13/cobra"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

func (a *application) newColumnsCommand() *cobra.Command {
	command := &cobra.Command{Use: "columns", Short: "Work with table columns"}
	command.AddCommand(a.newColumnsListCommand())
	command.AddCommand(a.newColumnsCreateCommand())
	command.AddCommand(a.newColumnsUpdateCommand())
	command.AddCommand(a.newColumnsDeleteCommand())
	return command
}

func (a *application) newColumnsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <table-id>",
		Short: "List columns",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			tableID, err := parseID("table ID", args[0])
			if err != nil {
				return err
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			result, err := client.Table(tableID).Columns().List(command.Context(), tables.ListColumnsOptions{})
			if err != nil {
				return err
			}
			return a.write(result)
		},
	}
}

func (a *application) newColumnsCreateCommand() *cobra.Command {
	var options tables.CreateColumnOptions
	command := &cobra.Command{
		Use:   "create <table-id>",
		Short: "Create a column",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			tableID, err := parseID("table ID", args[0])
			if err != nil {
				return err
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			result, err := client.Table(tableID).Columns().Create(command.Context(), options)
			if err != nil {
				return err
			}
			return a.write(result)
		},
	}
	flags := command.Flags()
	flags.StringVar(&options.Title, "title", "", "Column title")
	flags.StringVar(&options.TechnicalName, "technical-name", "", "Column technical name")
	flags.StringVar(&options.Type, "type", "", "Column type")
	flags.StringVar(&options.Subtype, "subtype", "", "Column subtype")
	flags.BoolVar(&options.Mandatory, "mandatory", false, "Require a value")
	flags.StringVar(&options.Description, "description", "", "Column description")
	_ = command.MarkFlagRequired("title")
	_ = command.MarkFlagRequired("type")
	return command
}

func (a *application) newColumnsUpdateCommand() *cobra.Command {
	var title, technicalName, subtype, description string
	var mandatory bool
	command := &cobra.Command{
		Use:   "update <table-id> <column-id>",
		Short: "Update a column",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			tableID, err := parseID("table ID", args[0])
			if err != nil {
				return err
			}
			columnID, err := parseID("column ID", args[1])
			if err != nil {
				return err
			}
			options := tables.UpdateColumnOptions{}
			flags := command.Flags()
			if flags.Changed("title") {
				options.Title = &title
			}
			if flags.Changed("technical-name") {
				options.TechnicalName = &technicalName
			}
			if flags.Changed("subtype") {
				options.Subtype = &subtype
			}
			if flags.Changed("mandatory") {
				options.Mandatory = &mandatory
			}
			if flags.Changed("description") {
				options.Description = &description
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			result, err := client.Table(tableID).Columns().Update(command.Context(), columnID, options)
			if err != nil {
				return err
			}
			return a.write(result)
		},
	}
	flags := command.Flags()
	flags.StringVar(&title, "title", "", "Column title")
	flags.StringVar(&technicalName, "technical-name", "", "Column technical name")
	flags.StringVar(&subtype, "subtype", "", "Column subtype")
	flags.BoolVar(&mandatory, "mandatory", false, "Require a value")
	flags.StringVar(&description, "description", "", "Column description")
	return command
}

func (a *application) newColumnsDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <table-id> <column-id>",
		Short: "Delete a column",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			tableID, err := parseID("table ID", args[0])
			if err != nil {
				return err
			}
			columnID, err := parseID("column ID", args[1])
			if err != nil {
				return err
			}
			client, err := a.client()
			if err != nil {
				return err
			}
			_, err = client.Table(tableID).Columns().Delete(command.Context(), columnID)
			return err
		},
	}
}
