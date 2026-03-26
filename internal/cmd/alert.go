package cmd

import (
	"fmt"
	"strconv"

	"github.com/ryo-imai-bit/redash-cli/internal/client"
	"github.com/spf13/cobra"
)

func newAlertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alert",
		Short: "Manage alerts",
		Long:  "List, get, create, update, delete, and manage Redash alerts.",
	}

	cmd.AddCommand(newAlertListCmd())
	cmd.AddCommand(newAlertGetCmd())
	cmd.AddCommand(newAlertCreateCmd())
	cmd.AddCommand(newAlertUpdateCmd())
	cmd.AddCommand(newAlertDeleteCmd())
	cmd.AddCommand(newAlertMuteCmd())
	cmd.AddCommand(newAlertSubscriptionCmd())

	return cmd
}

func newAlertListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all alerts",
		RunE: func(cmd *cobra.Command, args []string) error {
			alerts, err := GetClient().ListAlerts(GetContext())
			if err != nil {
				return err
			}
			return PrintResult(alerts)
		},
	}
}

func newAlertGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get an alert by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid alert ID: %s", args[0])
			}

			alert, err := GetClient().GetAlert(GetContext(), id)
			if err != nil {
				return err
			}
			return PrintResult(alert)
		},
	}
}

func newAlertCreateCmd() *cobra.Command {
	var name, column, op string
	var queryID int
	var value float64
	var rearm int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new alert",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &client.CreateAlertRequest{
				Name:    name,
				QueryID: queryID,
				Options: client.AlertOptions{
					Column: column,
					Op:     op,
					Value:  value,
				},
			}
			if cmd.Flags().Changed("rearm") {
				req.Rearm = &rearm
			}

			result, err := GetClient().CreateAlert(GetContext(), req)
			if err != nil {
				return err
			}
			return PrintResult(result)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "alert name (required)")
	cmd.Flags().IntVar(&queryID, "query-id", 0, "query ID (required)")
	cmd.Flags().StringVar(&column, "column", "", "column to monitor (required)")
	cmd.Flags().StringVar(&op, "op", "", "operator (>, >=, <, <=, ==, !=) (required)")
	cmd.Flags().Float64Var(&value, "value", 0, "threshold value (required)")
	cmd.Flags().IntVar(&rearm, "rearm", 0, "seconds to wait before re-triggering")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("query-id")
	cmd.MarkFlagRequired("column")
	cmd.MarkFlagRequired("op")

	return cmd
}

func newAlertUpdateCmd() *cobra.Command {
	var name, column, op string
	var value float64
	var rearm int

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an alert",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid alert ID: %s", args[0])
			}

			req := &client.UpdateAlertRequest{}
			if cmd.Flags().Changed("name") {
				req.Name = name
			}
			if cmd.Flags().Changed("column") || cmd.Flags().Changed("op") || cmd.Flags().Changed("value") {
				req.Options = &client.AlertOptions{
					Column: column,
					Op:     op,
					Value:  value,
				}
			}
			if cmd.Flags().Changed("rearm") {
				req.Rearm = &rearm
			}

			result, err := GetClient().UpdateAlert(GetContext(), id, req)
			if err != nil {
				return err
			}
			return PrintResult(result)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "alert name")
	cmd.Flags().StringVar(&column, "column", "", "column to monitor")
	cmd.Flags().StringVar(&op, "op", "", "operator (>, >=, <, <=, ==, !=)")
	cmd.Flags().Float64Var(&value, "value", 0, "threshold value")
	cmd.Flags().IntVar(&rearm, "rearm", 0, "seconds to wait before re-triggering")

	return cmd
}

func newAlertDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an alert",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid alert ID: %s", args[0])
			}

			if err := GetClient().DeleteAlert(GetContext(), id); err != nil {
				return err
			}
			fmt.Printf("Alert %d deleted successfully\n", id)
			return nil
		},
	}
}

func newAlertMuteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mute <id>",
		Short: "Mute an alert",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid alert ID: %s", args[0])
			}

			if err := GetClient().MuteAlert(GetContext(), id); err != nil {
				return err
			}
			fmt.Printf("Alert %d muted successfully\n", id)
			return nil
		},
	}
}

func newAlertSubscriptionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscription",
		Short: "Manage alert subscriptions",
	}

	cmd.AddCommand(newAlertSubscriptionListCmd())
	cmd.AddCommand(newAlertSubscriptionAddCmd())
	cmd.AddCommand(newAlertSubscriptionRemoveCmd())

	return cmd
}

func newAlertSubscriptionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <alert-id>",
		Short: "List subscriptions for an alert",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alertID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid alert ID: %s", args[0])
			}

			subs, err := GetClient().GetAlertSubscriptions(GetContext(), alertID)
			if err != nil {
				return err
			}
			return PrintResult(subs)
		},
	}
}

func newAlertSubscriptionAddCmd() *cobra.Command {
	var destinationID int

	cmd := &cobra.Command{
		Use:   "add <alert-id>",
		Short: "Add a subscription to an alert",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alertID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid alert ID: %s", args[0])
			}

			req := &client.CreateAlertSubscriptionRequest{}
			if cmd.Flags().Changed("destination-id") {
				req.DestinationID = &destinationID
			}

			sub, err := GetClient().AddAlertSubscription(GetContext(), alertID, req)
			if err != nil {
				return err
			}
			return PrintResult(sub)
		},
	}

	cmd.Flags().IntVar(&destinationID, "destination-id", 0, "destination ID (optional, defaults to email)")

	return cmd
}

func newAlertSubscriptionRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <alert-id> <subscription-id>",
		Short: "Remove a subscription from an alert",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			alertID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid alert ID: %s", args[0])
			}
			subID, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid subscription ID: %s", args[1])
			}

			if err := GetClient().RemoveAlertSubscription(GetContext(), alertID, subID); err != nil {
				return err
			}
			fmt.Printf("Subscription %d removed from alert %d\n", subID, alertID)
			return nil
		},
	}
}
