# redash-cli

A portable command-line interface for Redash written in Go.
Inspired by https://github.com/suthio/redash-mcp

## Features

- Manage queries, dashboards, visualizations, widgets, alerts, and more
- Cross-platform single binary (Linux, macOS, Windows)
- JSON and table output formats
- Configuration via environment variables or config file
- SOCKS proxy support

## Installation

### From source

```bash
go install github.com/ryo-imai-bit/redash-cli/cmd/redash-cli@latest
```

### Build from source

```bash
git clone https://github.com/ryo-imai-bit/redash-cli.git
cd redash-cli
make build
```

## Configuration

### Environment Variables

```bash
export REDASH_URL="https://your-redash-instance.com"
export REDASH_API_KEY="your-api-key"

# Optional
export REDASH_TIMEOUT=30000              # Timeout in milliseconds
export REDASH_MAX_RESULTS=1000           # Maximum results to return
export REDASH_SOCKS_PROXY=socks5://localhost:1080
export REDASH_EXTRA_HEADERS='{"Header-Name": "value"}'
```

### Config File

Create `~/.config/redash-cli/config.yaml`:

```yaml
default_profile: production

profiles:
  production:
    url: https://redash.example.com
    api_key: your-api-key
    timeout: 30000
  staging:
    url: https://redash-staging.example.com
    api_key: your-staging-api-key
```

## Usage

```bash
# List queries
redash-cli query list

# Get a specific query
redash-cli query get 123

# Execute a query
redash-cli query execute 123

# Execute with parameters
redash-cli query execute 123 --params '{"date": "2024-01-01"}'

# Get results as CSV
redash-cli query csv 123

# List dashboards
redash-cli dashboard list

# List data sources
redash-cli datasource list

# Get data source schema
redash-cli datasource schema 1

# Output as JSON
redash-cli query list -o json
```

## Commands

### Query

```
redash-cli query list          # List queries
redash-cli query get <id>      # Get a query
redash-cli query create        # Create a query
redash-cli query update <id>   # Update a query
redash-cli query archive <id>  # Archive a query
redash-cli query fork <id>     # Fork a query
redash-cli query execute <id>  # Execute a query
redash-cli query adhoc         # Execute ad-hoc query
redash-cli query csv <id>      # Get results as CSV
redash-cli query my            # List my queries
redash-cli query recent        # List recent queries
redash-cli query favorites     # List favorite queries
redash-cli query tags          # List query tags
```

### Dashboard

```
redash-cli dashboard list              # List dashboards
redash-cli dashboard get <id>          # Get a dashboard
redash-cli dashboard create            # Create a dashboard
redash-cli dashboard update <id>       # Update a dashboard
redash-cli dashboard archive <id>      # Archive a dashboard
redash-cli dashboard fork <id>         # Fork a dashboard
redash-cli dashboard share <id>        # Share a dashboard
redash-cli dashboard unshare <id>      # Unshare a dashboard
redash-cli dashboard my                # List my dashboards
redash-cli dashboard favorites         # List favorite dashboards
```

### Visualization

```
redash-cli visualization get <id>      # Get a visualization
redash-cli visualization create        # Create a visualization
redash-cli visualization update <id>   # Update a visualization
redash-cli visualization delete <id>   # Delete a visualization
```

### Widget

```
redash-cli widget list          # List widgets
redash-cli widget get <id>      # Get a widget
redash-cli widget create        # Create a widget
redash-cli widget update <id>   # Update a widget
redash-cli widget delete <id>   # Delete a widget
```

### Alert

```
redash-cli alert list                        # List alerts
redash-cli alert get <id>                    # Get an alert
redash-cli alert create                      # Create an alert
redash-cli alert update <id>                 # Update an alert
redash-cli alert delete <id>                 # Delete an alert
redash-cli alert mute <id>                   # Mute an alert
redash-cli alert subscription list <id>      # List subscriptions
redash-cli alert subscription add <id>       # Add subscription
redash-cli alert subscription remove <a> <s> # Remove subscription
```

### Query Snippet

```
redash-cli snippet list          # List snippets
redash-cli snippet get <id>      # Get a snippet
redash-cli snippet create        # Create a snippet
redash-cli snippet update <id>   # Update a snippet
redash-cli snippet delete <id>   # Delete a snippet
```

### Data Source

```
redash-cli datasource list       # List data sources
redash-cli datasource schema <id># Get data source schema
```

### Destination

```
redash-cli destination list      # List notification destinations
```

## Global Flags

```
-o, --output string    Output format (json, table) [default: table]
-c, --config string    Config file path
-p, --profile string   Config profile name
    --url string       Redash URL (overrides config)
    --api-key string   API key (overrides config)
    --timeout int      Request timeout in seconds [default: 30]
-v, --verbose          Verbose output
    --no-color         Disable color output
```

## License

MIT
