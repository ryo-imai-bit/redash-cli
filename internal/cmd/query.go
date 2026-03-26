package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ryo-imai-bit/redash-cli/internal/client"
	"github.com/spf13/cobra"
)

func newQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Manage queries",
		Long:  "List, get, create, update, execute, and manage Redash queries.",
	}

	cmd.AddCommand(newQueryListCmd())
	cmd.AddCommand(newQueryGetCmd())
	cmd.AddCommand(newQueryCreateCmd())
	cmd.AddCommand(newQueryUpdateCmd())
	cmd.AddCommand(newQueryArchiveCmd())
	cmd.AddCommand(newQueryForkCmd())
	cmd.AddCommand(newQueryExecuteCmd())
	cmd.AddCommand(newQueryAdhocCmd())
	cmd.AddCommand(newQueryCsvCmd())
	cmd.AddCommand(newQueryMyCmd())
	cmd.AddCommand(newQueryRecentCmd())
	cmd.AddCommand(newQueryTagsCmd())
	cmd.AddCommand(newQueryFavoritesCmd())
	cmd.AddCommand(newQueryFavoriteAddCmd())
	cmd.AddCommand(newQueryFavoriteRemoveCmd())

	return cmd
}

func newQueryListCmd() *cobra.Command {
	var page, pageSize int
	var search string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List queries",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := GetClient().ListQueries(GetContext(), page, pageSize, search)
			if err != nil {
				return err
			}
			return PrintResult(result.Results)
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 25, "items per page")
	cmd.Flags().StringVarP(&search, "search", "q", "", "search query")

	return cmd
}

func newQueryGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a query by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid query ID: %s", args[0])
			}

			query, err := GetClient().GetQuery(GetContext(), id)
			if err != nil {
				return err
			}
			return PrintResult(query)
		},
	}
}

func newQueryCreateCmd() *cobra.Command {
	var name, query, description string
	var dataSourceID int
	var tags []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new query",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &client.CreateQueryRequest{
				Name:         name,
				DataSourceID: dataSourceID,
				Query:        query,
				Description:  description,
				Tags:         tags,
			}

			result, err := GetClient().CreateQuery(GetContext(), req)
			if err != nil {
				return err
			}
			return PrintResult(result)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "query name (required)")
	cmd.Flags().IntVar(&dataSourceID, "data-source-id", 0, "data source ID (required)")
	cmd.Flags().StringVar(&query, "query", "", "SQL query (required)")
	cmd.Flags().StringVar(&description, "description", "", "query description")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "tags (comma-separated)")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("data-source-id")
	cmd.MarkFlagRequired("query")

	return cmd
}

func newQueryUpdateCmd() *cobra.Command {
	var name, query, description string
	var dataSourceID int
	var tags []string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a query",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid query ID: %s", args[0])
			}

			req := &client.UpdateQueryRequest{}
			if cmd.Flags().Changed("name") {
				req.Name = name
			}
			if cmd.Flags().Changed("data-source-id") {
				req.DataSourceID = dataSourceID
			}
			if cmd.Flags().Changed("query") {
				req.Query = query
			}
			if cmd.Flags().Changed("description") {
				req.Description = description
			}
			if cmd.Flags().Changed("tags") {
				req.Tags = tags
			}

			result, err := GetClient().UpdateQuery(GetContext(), id, req)
			if err != nil {
				return err
			}
			return PrintResult(result)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "query name")
	cmd.Flags().IntVar(&dataSourceID, "data-source-id", 0, "data source ID")
	cmd.Flags().StringVar(&query, "query", "", "SQL query")
	cmd.Flags().StringVar(&description, "description", "", "query description")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "tags (comma-separated)")

	return cmd
}

func newQueryArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a query",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid query ID: %s", args[0])
			}

			if err := GetClient().ArchiveQuery(GetContext(), id); err != nil {
				return err
			}
			fmt.Printf("Query %d archived successfully\n", id)
			return nil
		},
	}
}

func newQueryForkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fork <id>",
		Short: "Fork a query",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid query ID: %s", args[0])
			}

			query, err := GetClient().ForkQuery(GetContext(), id)
			if err != nil {
				return err
			}
			return PrintResult(query)
		},
	}
}

func newQueryExecuteCmd() *cobra.Command {
	var paramsJSON string

	cmd := &cobra.Command{
		Use:   "execute <id>",
		Short: "Execute a saved query",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid query ID: %s", args[0])
			}

			var params map[string]any
			if paramsJSON != "" {
				if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
					return fmt.Errorf("invalid params JSON: %w", err)
				}
			}

			result, err := GetClient().ExecuteQuery(GetContext(), id, params)
			if err != nil {
				return err
			}
			return PrintResult(result)
		},
	}

	cmd.Flags().StringVar(&paramsJSON, "params", "", "query parameters as JSON")

	return cmd
}

func newQueryAdhocCmd() *cobra.Command {
	var query string
	var dataSourceID int

	cmd := &cobra.Command{
		Use:   "adhoc",
		Short: "Execute an ad-hoc query",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := GetClient().ExecuteAdhocQuery(GetContext(), query, dataSourceID)
			if err != nil {
				return err
			}
			return PrintResult(result)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "SQL query (required)")
	cmd.Flags().IntVar(&dataSourceID, "data-source-id", 0, "data source ID (required)")

	cmd.MarkFlagRequired("query")
	cmd.MarkFlagRequired("data-source-id")

	return cmd
}

func newQueryCsvCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "csv <id>",
		Short: "Get query results as CSV",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid query ID: %s", args[0])
			}

			csv, err := GetClient().GetQueryResultsCSV(GetContext(), id)
			if err != nil {
				return err
			}
			fmt.Print(csv)
			return nil
		},
	}
}

func newQueryMyCmd() *cobra.Command {
	var page, pageSize int

	cmd := &cobra.Command{
		Use:   "my",
		Short: "List my queries",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := GetClient().GetMyQueries(GetContext(), page, pageSize)
			if err != nil {
				return err
			}
			return PrintResult(result.Results)
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 25, "items per page")

	return cmd
}

func newQueryRecentCmd() *cobra.Command {
	var page, pageSize int

	cmd := &cobra.Command{
		Use:   "recent",
		Short: "List recent queries",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := GetClient().GetRecentQueries(GetContext(), page, pageSize)
			if err != nil {
				return err
			}
			return PrintResult(result.Results)
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 25, "items per page")

	return cmd
}

func newQueryTagsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tags",
		Short: "List query tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			tags, err := GetClient().GetQueryTags(GetContext())
			if err != nil {
				return err
			}
			return PrintResult(tags)
		},
	}
}

func newQueryFavoritesCmd() *cobra.Command {
	var page, pageSize int

	cmd := &cobra.Command{
		Use:   "favorites",
		Short: "List favorite queries",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := GetClient().GetFavoriteQueries(GetContext(), page, pageSize)
			if err != nil {
				return err
			}
			return PrintResult(result.Results)
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "page number")
	cmd.Flags().IntVar(&pageSize, "page-size", 25, "items per page")

	return cmd
}

func newQueryFavoriteAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "favorite-add <id>",
		Short: "Add a query to favorites",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid query ID: %s", args[0])
			}

			if err := GetClient().AddQueryFavorite(GetContext(), id); err != nil {
				return err
			}
			fmt.Printf("Query %d added to favorites\n", id)
			return nil
		},
	}
}

func newQueryFavoriteRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "favorite-remove <id>",
		Short: "Remove a query from favorites",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid query ID: %s", args[0])
			}

			if err := GetClient().RemoveQueryFavorite(GetContext(), id); err != nil {
				return err
			}
			fmt.Printf("Query %d removed from favorites\n", id)
			return nil
		},
	}
}
