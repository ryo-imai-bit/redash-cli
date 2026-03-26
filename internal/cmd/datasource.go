package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newDatasourceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "datasource",
		Aliases: []string{"ds"},
		Short:   "Manage data sources",
		Long:    "List data sources and get their schemas.",
	}

	cmd.AddCommand(newDatasourceListCmd())
	cmd.AddCommand(newDatasourceSchemaCmd())

	return cmd
}

func newDatasourceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all data sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			dataSources, err := GetClient().ListDataSources(GetContext())
			if err != nil {
				return err
			}
			return PrintResult(dataSources)
		},
	}
}

func newDatasourceSchemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "schema <id>",
		Short: "Get the schema of a data source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid data source ID: %s", args[0])
			}

			schema, err := GetClient().GetDataSourceSchema(GetContext(), id)
			if err != nil {
				return err
			}
			return PrintResult(schema)
		},
	}
}
