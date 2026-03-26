package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ryo-imai-bit/redash-cli/internal/client"
	"github.com/spf13/cobra"
)

func newVisualizationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "visualization",
		Aliases: []string{"viz"},
		Short:   "Manage visualizations",
		Long:    "Get, create, update, and delete Redash visualizations.",
	}

	cmd.AddCommand(newVisualizationGetCmd())
	cmd.AddCommand(newVisualizationCreateCmd())
	cmd.AddCommand(newVisualizationUpdateCmd())
	cmd.AddCommand(newVisualizationDeleteCmd())

	return cmd
}

func newVisualizationGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a visualization by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid visualization ID: %s", args[0])
			}

			viz, err := GetClient().GetVisualization(GetContext(), id)
			if err != nil {
				return err
			}
			return PrintResult(viz)
		},
	}
}

func newVisualizationCreateCmd() *cobra.Command {
	var queryID int
	var vizType, name, description, optionsJSON string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new visualization",
		RunE: func(cmd *cobra.Command, args []string) error {
			var options map[string]any
			if optionsJSON != "" {
				if err := json.Unmarshal([]byte(optionsJSON), &options); err != nil {
					return fmt.Errorf("invalid options JSON: %w", err)
				}
			} else {
				options = make(map[string]any)
			}

			req := &client.CreateVisualizationRequest{
				QueryID:     queryID,
				Type:        vizType,
				Name:        name,
				Description: description,
				Options:     options,
			}

			result, err := GetClient().CreateVisualization(GetContext(), req)
			if err != nil {
				return err
			}
			return PrintResult(result)
		},
	}

	cmd.Flags().IntVar(&queryID, "query-id", 0, "query ID (required)")
	cmd.Flags().StringVar(&vizType, "type", "", "visualization type (required)")
	cmd.Flags().StringVar(&name, "name", "", "visualization name (required)")
	cmd.Flags().StringVar(&description, "description", "", "visualization description")
	cmd.Flags().StringVar(&optionsJSON, "options", "", "visualization options as JSON")

	cmd.MarkFlagRequired("query-id")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("name")

	return cmd
}

func newVisualizationUpdateCmd() *cobra.Command {
	var name, description, optionsJSON string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a visualization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid visualization ID: %s", args[0])
			}

			req := &client.UpdateVisualizationRequest{}
			if cmd.Flags().Changed("name") {
				req.Name = name
			}
			if cmd.Flags().Changed("description") {
				req.Description = description
			}
			if cmd.Flags().Changed("options") {
				var options map[string]any
				if err := json.Unmarshal([]byte(optionsJSON), &options); err != nil {
					return fmt.Errorf("invalid options JSON: %w", err)
				}
				req.Options = options
			}

			result, err := GetClient().UpdateVisualization(GetContext(), id, req)
			if err != nil {
				return err
			}
			return PrintResult(result)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "visualization name")
	cmd.Flags().StringVar(&description, "description", "", "visualization description")
	cmd.Flags().StringVar(&optionsJSON, "options", "", "visualization options as JSON")

	return cmd
}

func newVisualizationDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a visualization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid visualization ID: %s", args[0])
			}

			if err := GetClient().DeleteVisualization(GetContext(), id); err != nil {
				return err
			}
			fmt.Printf("Visualization %d deleted successfully\n", id)
			return nil
		},
	}
}
