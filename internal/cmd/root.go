package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ryo-imai-bit/redash-cli/internal/client"
	"github.com/ryo-imai-bit/redash-cli/internal/config"
	"github.com/ryo-imai-bit/redash-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	cfgFile   string
	profile   string
	outputFmt string
	verbose   bool
	noColor   bool
	urlFlag   string
	apiKeyFlag string
	timeout   int

	// Global config and client
	cfg       *config.Config
	apiClient *client.Client
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "redash-cli",
	Short: "Command line interface for Redash",
	Long: `redash-cli is a portable command-line interface for managing Redash
queries, dashboards, visualizations, alerts, and more.

Set REDASH_URL and REDASH_API_KEY environment variables or use a config file.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for help and version commands
		if cmd.Name() == "help" || cmd.Name() == "version" || cmd.Name() == "completion" {
			return nil
		}

		var err error
		cfg, err = config.Load(cfgFile, profile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override with flags
		if urlFlag != "" {
			cfg.URL = urlFlag
		}
		if apiKeyFlag != "" {
			cfg.APIKey = apiKeyFlag
		}
		if outputFmt != "" {
			cfg.Output = outputFmt
		}
		cfg.Verbose = verbose
		cfg.NoColor = noColor

		// Validate required config
		if err := cfg.Validate(); err != nil {
			return err
		}

		// Create API client
		opts := []client.Option{
			client.WithTimeout(cfg.Timeout),
			client.WithMaxResults(cfg.MaxResults),
		}
		if len(cfg.ExtraHeaders) > 0 {
			opts = append(opts, client.WithExtraHeaders(cfg.ExtraHeaders))
		}
		if cfg.SocksProxy != "" {
			opts = append(opts, client.WithSocksProxy(cfg.SocksProxy))
		}

		apiClient = client.New(cfg.URL, cfg.APIKey, opts...)
		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "config profile name")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "", "output format (json, table)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
	rootCmd.PersistentFlags().StringVar(&urlFlag, "url", "", "Redash URL (overrides config)")
	rootCmd.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "API key (overrides config)")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 30, "request timeout in seconds")

	// Add subcommands
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newQueryCmd())
	rootCmd.AddCommand(newDashboardCmd())
	rootCmd.AddCommand(newVisualizationCmd())
	rootCmd.AddCommand(newWidgetCmd())
	rootCmd.AddCommand(newAlertCmd())
	rootCmd.AddCommand(newSnippetCmd())
	rootCmd.AddCommand(newDatasourceCmd())
	rootCmd.AddCommand(newDestinationCmd())
}

// GetClient returns the API client
func GetClient() *client.Client {
	return apiClient
}

// GetConfig returns the config
func GetConfig() *config.Config {
	return cfg
}

// GetContext returns a context for API calls
func GetContext() context.Context {
	return context.Background()
}

// GetFormatter returns the output formatter based on config
func GetFormatter() output.Formatter {
	format := output.FormatTable
	if cfg != nil && cfg.Output == "json" {
		format = output.FormatJSON
	}
	return output.NewFormatter(format)
}

// PrintResult prints the result using the configured formatter
func PrintResult(data any) error {
	formatter := GetFormatter()
	return formatter.Format(os.Stdout, data)
}

// version command
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("redash-cli version dev")
		},
	}
}
