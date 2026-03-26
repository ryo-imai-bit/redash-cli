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

	// If profile is specified, load from profiles section
	if profile == "" {
		profile = v.GetString("default_profile")
		if profile == "" {
			profile = os.Getenv("REDASH_PROFILE")
		}
	}

	if profile != "" {
		sub := v.Sub("profiles." + profile)
		if sub != nil {
			cfg.URL = sub.GetString("url")
			cfg.APIKey = sub.GetString("api_key")
			cfg.Timeout = time.Duration(sub.GetInt("timeout")) * time.Millisecond
			cfg.MaxResults = sub.GetInt("max_results")
			cfg.SocksProxy = sub.GetString("socks_proxy")

			// Load extra headers from profile
			headers := sub.GetStringMapString("extra_headers")
			for k, v := range headers {
				cfg.ExtraHeaders[k] = v
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

	// Parse REDASH_COOKIE_* environment variables
	cookies := parseCookieEnvVars()
	if cookies != "" {
		if existing, ok := cfg.ExtraHeaders["Cookie"]; ok && existing != "" {
			cfg.ExtraHeaders["Cookie"] = existing + "; " + cookies
		} else {
			cfg.ExtraHeaders["Cookie"] = cookies
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

// parseCookieEnvVars parses REDASH_COOKIE_* environment variables into a cookie string
// Example: REDASH_COOKIE_SESSION=abc -> session=abc
// Example: REDASH_COOKIE__OAUTH2_PROXY=xyz -> _oauth2_proxy=xyz (double underscore for leading underscore)
func parseCookieEnvVars() string {
	const prefix = "REDASH_COOKIE_"
	var cookies []string

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, prefix) {
			continue
		}

		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 || parts[1] == "" {
			continue
		}

		// Extract cookie name from env var name
		// REDASH_COOKIE_SESSION -> session
		// REDASH_COOKIE__OAUTH2_PROXY -> _oauth2_proxy (leading double underscore becomes single)
		cookieName := strings.TrimPrefix(parts[0], prefix)
		cookieName = strings.ToLower(cookieName)

		// Handle leading underscore: __NAME -> _name
		if strings.HasPrefix(cookieName, "_") {
			cookieName = "_" + strings.TrimPrefix(cookieName, "_")
		}

		cookies = append(cookies, cookieName+"="+parts[1])
	}

	return strings.Join(cookies, "; ")
}
