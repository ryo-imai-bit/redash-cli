package cmd

import (
	"github.com/spf13/cobra"
)

func newDestinationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "destination",
		Aliases: []string{"dest"},
		Short:   "Manage notification destinations",
		Long:    "List notification destinations for alerts.",
	}

	cmd.AddCommand(newDestinationListCmd())

	return cmd
}

func newDestinationListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all notification destinations",
		RunE: func(cmd *cobra.Command, args []string) error {
			destinations, err := GetClient().ListDestinations(GetContext())
			if err != nil {
				return err
			}
			return PrintResult(destinations)
		},
	}
}
