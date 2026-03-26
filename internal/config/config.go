package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	URL          string            `mapstructure:"url"`
	APIKey       string            `mapstructure:"api_key"`
	Timeout      time.Duration     `mapstructure:"timeout"`
	MaxResults   int               `mapstructure:"max_results"`
	ExtraHeaders map[string]string `mapstructure:"extra_headers"`
	SocksProxy   string            `mapstructure:"socks_proxy"`
	Output       string            `mapstructure:"output"`
	Verbose      bool              `mapstructure:"verbose"`
	NoColor      bool              `mapstructure:"no_color"`
}

// Load loads configuration from file, environment variables, and flags
func Load(configFile, profile string) (*Config, error) {
	v := viper.New()

	// Config file settings
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("$HOME/.config/redash-cli")
		v.AddConfigPath("$HOME/.redash-cli")
		v.AddConfigPath(".")
	}

	// Environment variable settings
	v.SetEnvPrefix("REDASH")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	// Default values
	v.SetDefault("timeout", 30000)
	v.SetDefault("max_results", 1000)
	v.SetDefault("output", "table")
	v.SetDefault("verbose", false)
	v.SetDefault("no_color", false)

	// Try to read config file (ignore if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only return error if it's not a "file not found" error
			if configFile != "" {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		}
	}

	cfg := &Config{
		Output:       v.GetString("output"),
		Verbose:      v.GetBool("verbose"),
		NoColor:      v.GetBool("no_color"),
		ExtraHeaders: make(map[string]string),
	}

	// Load top-level config values first
	cfg.URL = v.GetString("url")
	cfg.APIKey = v.GetString("api_key")
	cfg.Timeout = time.Duration(v.GetInt("timeout")) * time.Millisecond
	cfg.MaxResults = v.GetInt("max_results")
	cfg.SocksProxy = v.GetString("socks_proxy")

	// Load extra headers from top-level config
	headers := v.GetStringMapString("extra_headers")
	for k, val := range headers {
		cfg.ExtraHeaders[k] = val
	}

	// If profile is specified, override with profile settings
	if profile == "" {
		profile = v.GetString("default_profile")
		if profile == "" {
			profile = os.Getenv("REDASH_PROFILE")
		}
	}

	if profile != "" {
		sub := v.Sub("profiles." + profile)
		if sub != nil {
			if u := sub.GetString("url"); u != "" {
				cfg.URL = u
			}
			if k := sub.GetString("api_key"); k != "" {
				cfg.APIKey = k
			}
			if t := sub.GetInt("timeout"); t > 0 {
				cfg.Timeout = time.Duration(t) * time.Millisecond
			}
			if m := sub.GetInt("max_results"); m > 0 {
				cfg.MaxResults = m
			}
			if s := sub.GetString("socks_proxy"); s != "" {
				cfg.SocksProxy = s
			}

			// Load extra headers from profile (merge with top-level)
			profileHeaders := sub.GetStringMapString("extra_headers")
			for k, val := range profileHeaders {
				cfg.ExtraHeaders[k] = val
			}
		}
	}

	// Environment variables override profile settings
	if url := os.Getenv("REDASH_URL"); url != "" {
		cfg.URL = url
	}
	if apiKey := os.Getenv("REDASH_API_KEY"); apiKey != "" {
		cfg.APIKey = apiKey
	}
	if timeout := os.Getenv("REDASH_TIMEOUT"); timeout != "" {
		if t, err := strconv.Atoi(timeout); err == nil {
			cfg.Timeout = time.Duration(t) * time.Millisecond
		}
	}
	if maxResults := os.Getenv("REDASH_MAX_RESULTS"); maxResults != "" {
		if m, err := strconv.Atoi(maxResults); err == nil {
			cfg.MaxResults = m
		}
	}
	if socksProxy := os.Getenv("REDASH_SOCKS_PROXY"); socksProxy != "" {
		cfg.SocksProxy = socksProxy
	}

	// Parse extra headers from environment
	if headers := os.Getenv("REDASH_EXTRA_HEADERS"); headers != "" {
		if err := parseExtraHeaders(headers, cfg.ExtraHeaders); err != nil {
			return nil, fmt.Errorf("failed to parse REDASH_EXTRA_HEADERS: %w", err)
		}
	}

	// Set defaults if not set
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxResults == 0 {
		cfg.MaxResults = 1000
	}

	return cfg, nil
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("REDASH_URL is required")
	}
	if c.APIKey == "" {
		return fmt.Errorf("REDASH_API_KEY is required")
	}
	return nil
}

// parseExtraHeaders parses extra headers from string (JSON or key=value format)
func parseExtraHeaders(s string, headers map[string]string) error {
	s = strings.TrimSpace(s)

	// Try JSON format first
	if strings.HasPrefix(s, "{") {
		return json.Unmarshal([]byte(s), &headers)
	}

	// Parse key=value;key=value format
	pairs := strings.Split(s, ";")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid header format: %s (expected key=value)", pair)
		}
		headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return nil
}

