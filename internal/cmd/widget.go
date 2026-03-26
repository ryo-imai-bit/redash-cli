package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ryo-imai-bit/redash-cli/internal/client"
	"github.com/spf13/cobra"
)

func newWidgetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "widget",
		Short: "Manage widgets",
		Long:  "List, get, create, update, and delete dashboard widgets.",
	}

	cmd.AddCommand(newWidgetListCmd())
	cmd.AddCommand(newWidgetGetCmd())
	cmd.AddCommand(newWidgetCreateCmd())
	cmd.AddCommand(newWidgetUpdateCmd())
	cmd.AddCommand(newWidgetDeleteCmd())

	return cmd
}

func newWidgetListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all widgets",
		RunE: func(cmd *cobra.Command, args []string) error {
			widgets, err := GetClient().ListWidgets(GetContext())
			if err != nil {
				return err
			}
			return PrintResult(widgets)
		},
	}
}

func newWidgetGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a widget by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid widget ID: %s", args[0])
			}

			widget, err := GetClient().GetWidget(GetContext(), id)
			if err != nil {
				return err
			}
			return PrintResult(widget)
		},
	}
}

func newWidgetCreateCmd() *cobra.Command {
	var dashboardID, visualizationID, width int
	var text, optionsJSON string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new widget",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &client.CreateWidgetRequest{
				DashboardID: dashboardID,
				Width:       width,
			}

			if cmd.Flags().Changed("visualization-id") {
				req.VisualizationID = &visualizationID
			}
			if cmd.Flags().Changed("text") {
				req.Text = text
			}
			if cmd.Flags().Changed("options") {
				var options map[string]any
				if err := json.Unmarshal([]byte(optionsJSON), &options); err != nil {
					return fmt.Errorf("invalid options JSON: %w", err)
				}
				req.Options = options
			}

			result, err := GetClient().CreateWidget(GetContext(), req)
			if err != nil {
				return err
			}
			return PrintResult(result)
		},
	}

	cmd.Flags().IntVar(&dashboardID, "dashboard-id", 0, "dashboard ID (required)")
	cmd.Flags().IntVar(&visualizationID, "visualization-id", 0, "visualization ID")
	cmd.Flags().IntVar(&width, "width", 1, "widget width (required)")
	cmd.Flags().StringVar(&text, "text", "", "widget text (for text widgets)")
	cmd.Flags().StringVar(&optionsJSON, "options", "", "widget options as JSON")

	cmd.MarkFlagRequired("dashboard-id")
	cmd.MarkFlagRequired("width")

	return cmd
}

func newWidgetUpdateCmd() *cobra.Command {
	var width int
	var text, optionsJSON string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a widget",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid widget ID: %s", args[0])
			}

			req := &client.UpdateWidgetRequest{}
			if cmd.Flags().Changed("width") {
				req.Width = width
			}
			if cmd.Flags().Changed("text") {
				req.Text = text
			}
			if cmd.Flags().Changed("options") {
				var options map[string]any
				if err := json.Unmarshal([]byte(optionsJSON), &options); err != nil {
					return fmt.Errorf("invalid options JSON: %w", err)
				}
				req.Options = options
			}

			result, err := GetClient().UpdateWidget(GetContext(), id, req)
			if err != nil {
				return err
			}
			return PrintResult(result)
		},
	}

	cmd.Flags().IntVar(&width, "width", 0, "widget width")
	cmd.Flags().StringVar(&text, "text", "", "widget text")
	cmd.Flags().StringVar(&optionsJSON, "options", "", "widget options as JSON")

	return cmd
}

func newWidgetDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a widget",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid widget ID: %s", args[0])
			}

			if err := GetClient().DeleteWidget(GetContext(), id); err != nil {
				return err
			}
			fmt.Printf("Widget %d deleted successfully\n", id)
			return nil
		},
	}
}
