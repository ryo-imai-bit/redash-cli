package cmd

import (
	"fmt"
	"strconv"

	"github.com/ryo-imai-bit/redash-cli/internal/client"
	"github.com/spf13/cobra"
)

func newDashboardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Manage dashboards",
		Long:  "List, get, create, update, and manage Redash dashboards.",
	}

	cmd.AddCommand(newDashboardListCmd())
	cmd.AddCommand(newDashboardGetCmd())
	cmd.AddCommand(newDashboardCreateCmd())
	cmd.AddCommand(newDashboardUpdateCmd())
	cmd.AddCommand(newDashboardArchiveCmd())
	cmd.AddCommand(newDashboardForkCmd())
	cmd.AddCommand(newDashboardShareCmd())
	cmd.AddCommand(newDashboardUnshareCmd())
	cmd.AddCommand(newDashboardPublicCmd())
	cmd.AddCommand(newDashboardMyCmd())
	cmd.AddCommand(newDashboardTagsCmd())
	cmd.AddCommand(newDashboardFavoritesCmd())
	cmd.AddCommand(newDashboardFavoriteAddCmd())
	cmd.AddCommand(newDashboardFavoriteRemoveCmd())

	return cmd
}

func newDashboardListCmd() *cobra.Command {
	var page, pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List dashboards",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := GetClient().ListDashboards(GetContext(), page, pageSize)
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

func newDashboardGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id-or-slug>",
		Short: "Get a dashboard by ID or slug",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var dashboard *client.Dashboard
			var err error

			// Try to parse as ID first
			if id, parseErr := strconv.Atoi(args[0]); parseErr == nil {
				dashboard, err = GetClient().GetDashboard(GetContext(), id)
			} else {
				// Treat as slug
				dashboard, err = GetClient().GetDashboardBySlug(GetContext(), args[0])
			}

			if err != nil {
				return err
			}
			return PrintResult(dashboard)
		},
	}
}

func newDashboardCreateCmd() *cobra.Command {
	var name string
	var tags []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &client.CreateDashboardRequest{
				Name: name,
				Tags: tags,
			}

			result, err := GetClient().CreateDashboard(GetContext(), req)
			if err != nil {
				return err
			}
			return PrintResult(result)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "dashboard name (required)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "tags (comma-separated)")

	cmd.MarkFlagRequired("name")

	return cmd
}

func newDashboardUpdateCmd() *cobra.Command {
	var name string
	var tags []string
	var filtersEnabled bool

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a dashboard",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid dashboard ID: %s", args[0])
			}

			req := &client.UpdateDashboardRequest{}
			if cmd.Flags().Changed("name") {
				req.Name = name
			}
			if cmd.Flags().Changed("tags") {
				req.Tags = tags
			}
			if cmd.Flags().Changed("filters-enabled") {
				req.DashboardFiltersEnabled = &filtersEnabled
			}

			result, err := GetClient().UpdateDashboard(GetContext(), id, req)
			if err != nil {
				return err
			}
			return PrintResult(result)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "dashboard name")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "tags (comma-separated)")
	cmd.Flags().BoolVar(&filtersEnabled, "filters-enabled", false, "enable dashboard filters")

	return cmd
}

func newDashboardArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a dashboard",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid dashboard ID: %s", args[0])
			}

			if err := GetClient().ArchiveDashboard(GetContext(), id); err != nil {
				return err
			}
			fmt.Printf("Dashboard %d archived successfully\n", id)
			return nil
		},
	}
}

func newDashboardForkCmd() *cobra.Command {
	var legacy bool
	var name string

	cmd := &cobra.Command{
		Use:   "fork <id-or-slug>",
		Short: "Fork a dashboard",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var dashboard *client.Dashboard
			var err error

			// Parse ID or slug
			id, parseErr := strconv.Atoi(args[0])
			isSlug := parseErr != nil

			if legacy || isSlug || name != "" {
				// Legacy mode: create dashboard + copy widgets
				var original *client.Dashboard
				if isSlug {
					original, err = GetClient().GetDashboardBySlug(GetContext(), args[0])
				} else {
					original, err = GetClient().GetDashboard(GetContext(), id)
				}
				if err != nil {
					return err
				}
				dashboard, err = GetClient().ForkDashboardLegacy(GetContext(), original, name)
			} else {
				// Use native fork API
				dashboard, err = GetClient().ForkDashboard(GetContext(), id)
			}

			if err != nil {
				return err
			}
			return PrintResult(dashboard)
		},
	}

	cmd.Flags().BoolVar(&legacy, "legacy", false, "use legacy fork method (for older Redash versions)")
	cmd.Flags().StringVar(&name, "name", "", "name for the new dashboard (implies --legacy)")

	return cmd
}

func newDashboardShareCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "share <id>",
		Short: "Share a dashboard (create public link)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid dashboard ID: %s", args[0])
			}

			resp, err := GetClient().ShareDashboard(GetContext(), id)
			if err != nil {
				return err
			}
			return PrintResult(resp)
		},
	}
}

func newDashboardUnshareCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unshare <id>",
		Short: "Unshare a dashboard (revoke public link)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid dashboard ID: %s", args[0])
			}

			if err := GetClient().UnshareDashboard(GetContext(), id); err != nil {
				return err
			}
			fmt.Printf("Dashboard %d unshared successfully\n", id)
			return nil
		},
	}
}

func newDashboardPublicCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "public <token>",
		Short: "Get a public dashboard by token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dashboard, err := GetClient().GetPublicDashboard(GetContext(), args[0])
			if err != nil {
				return err
			}
			return PrintResult(dashboard)
		},
	}
}

func newDashboardMyCmd() *cobra.Command {
	var page, pageSize int

	cmd := &cobra.Command{
		Use:   "my",
		Short: "List my dashboards",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := GetClient().GetMyDashboards(GetContext(), page, pageSize)
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

func newDashboardTagsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tags",
		Short: "List dashboard tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			tags, err := GetClient().GetDashboardTags(GetContext())
			if err != nil {
				return err
			}
			return PrintResult(tags)
		},
	}
}

func newDashboardFavoritesCmd() *cobra.Command {
	var page, pageSize int

	cmd := &cobra.Command{
		Use:   "favorites",
		Short: "List favorite dashboards",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := GetClient().GetFavoriteDashboards(GetContext(), page, pageSize)
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

func newDashboardFavoriteAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "favorite-add <id>",
		Short: "Add a dashboard to favorites",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid dashboard ID: %s", args[0])
			}

			if err := GetClient().AddDashboardFavorite(GetContext(), id); err != nil {
				return err
			}
			fmt.Printf("Dashboard %d added to favorites\n", id)
			return nil
		},
	}
}

func newDashboardFavoriteRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "favorite-remove <id>",
		Short: "Remove a dashboard from favorites",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid dashboard ID: %s", args[0])
			}

			if err := GetClient().RemoveDashboardFavorite(GetContext(), id); err != nil {
				return err
			}
			fmt.Printf("Dashboard %d removed from favorites\n", id)
			return nil
		},
	}
}
