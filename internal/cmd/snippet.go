package cmd

import (
	"fmt"
	"strconv"

	"github.com/ryo-imai-bit/redash-cli/internal/client"
	"github.com/spf13/cobra"
)

func newSnippetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snippet",
		Short: "Manage query snippets",
		Long:  "List, get, create, update, and delete query snippets.",
	}

	cmd.AddCommand(newSnippetListCmd())
	cmd.AddCommand(newSnippetGetCmd())
	cmd.AddCommand(newSnippetCreateCmd())
	cmd.AddCommand(newSnippetUpdateCmd())
	cmd.AddCommand(newSnippetDeleteCmd())

	return cmd
}

func newSnippetListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all query snippets",
		RunE: func(cmd *cobra.Command, args []string) error {
			snippets, err := GetClient().ListQuerySnippets(GetContext())
			if err != nil {
				return err
			}
			return PrintResult(snippets)
		},
	}
}

func newSnippetGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a query snippet by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid snippet ID: %s", args[0])
			}

			snippet, err := GetClient().GetQuerySnippet(GetContext(), id)
			if err != nil {
				return err
			}
			return PrintResult(snippet)
		},
	}
}

func newSnippetCreateCmd() *cobra.Command {
	var trigger, snippet, description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new query snippet",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &client.CreateQuerySnippetRequest{
				Trigger:     trigger,
				Snippet:     snippet,
				Description: description,
			}

			result, err := GetClient().CreateQuerySnippet(GetContext(), req)
			if err != nil {
				return err
			}
			return PrintResult(result)
		},
	}

	cmd.Flags().StringVar(&trigger, "trigger", "", "snippet trigger (required)")
	cmd.Flags().StringVar(&snippet, "snippet", "", "snippet content (required)")
	cmd.Flags().StringVar(&description, "description", "", "snippet description")

	cmd.MarkFlagRequired("trigger")
	cmd.MarkFlagRequired("snippet")

	return cmd
}

func newSnippetUpdateCmd() *cobra.Command {
	var trigger, snippet, description string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a query snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid snippet ID: %s", args[0])
			}

			req := &client.UpdateQuerySnippetRequest{}
			if cmd.Flags().Changed("trigger") {
				req.Trigger = trigger
			}
			if cmd.Flags().Changed("snippet") {
				req.Snippet = snippet
			}
			if cmd.Flags().Changed("description") {
				req.Description = description
			}

			result, err := GetClient().UpdateQuerySnippet(GetContext(), id, req)
			if err != nil {
				return err
			}
			return PrintResult(result)
		},
	}

	cmd.Flags().StringVar(&trigger, "trigger", "", "snippet trigger")
	cmd.Flags().StringVar(&snippet, "snippet", "", "snippet content")
	cmd.Flags().StringVar(&description, "description", "", "snippet description")

	return cmd
}

func newSnippetDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a query snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid snippet ID: %s", args[0])
			}

			if err := GetClient().DeleteQuerySnippet(GetContext(), id); err != nil {
				return err
			}
			fmt.Printf("Snippet %d deleted successfully\n", id)
			return nil
		},
	}
}
