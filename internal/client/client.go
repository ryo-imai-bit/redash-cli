package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/net/proxy"
)

// Client is the Redash API client
type Client struct {
	baseURL      string
	apiKey       string
	httpClient   *http.Client
	timeout      time.Duration
	maxResults   int
	extraHeaders map[string]string
}

// Option is a function that configures the client
type Option func(*Client)

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
		c.httpClient.Timeout = timeout
	}
}

// WithMaxResults sets the maximum results to return
func WithMaxResults(max int) Option {
	return func(c *Client) {
		c.maxResults = max
	}
}

// WithExtraHeaders sets additional HTTP headers
func WithExtraHeaders(headers map[string]string) Option {
	return func(c *Client) {
		for k, v := range headers {
			c.extraHeaders[k] = v
		}
	}
}

// WithSocksProxy configures a SOCKS proxy
func WithSocksProxy(proxyURL string) Option {
	return func(c *Client) {
		if proxyURL == "" {
			return
		}
		u, err := url.Parse(proxyURL)
		if err != nil {
			return
		}
		dialer, err := proxy.FromURL(u, proxy.Direct)
		if err != nil {
			return
		}
		c.httpClient.Transport = &http.Transport{
			Dial: dialer.Dial,
		}
	}
}

// New creates a new Redash API client
func New(baseURL, apiKey string, opts ...Option) *Client {
	c := &Client{
		baseURL:      baseURL,
		apiKey:       apiKey,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		timeout:      30 * time.Second,
		maxResults:   1000,
		extraHeaders: make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// do performs an HTTP request
func (c *Client) do(ctx context.Context, method, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Key "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// doRaw performs an HTTP request and returns the raw response body
func (c *Client) doRaw(ctx context.Context, method, path string) ([]byte, error) {
	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Key "+c.apiKey)
	for k, v := range c.extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// pollJobResult polls for async query results
func (c *Client) pollJobResult(ctx context.Context, jobID string) (*QueryResult, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.After(60 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("query execution timed out")
		case <-ticker.C:
			var jobResp struct {
				Job Job `json:"job"`
			}
			if err := c.do(ctx, "GET", "/api/jobs/"+jobID, nil, &jobResp); err != nil {
				return nil, fmt.Errorf("failed to get job status: %w", err)
			}

			switch jobResp.Job.Status {
			case JobStatusSuccess:
				var result QueryResult
				if err := c.do(ctx, "GET", fmt.Sprintf("/api/query_results/%d", jobResp.Job.QueryResultID), nil, &result); err != nil {
					return nil, fmt.Errorf("failed to get query result: %w", err)
				}
				return &result, nil
			case JobStatusFailure:
				return nil, fmt.Errorf("query execution failed: %s", jobResp.Job.Error)
			case JobStatusCancelled:
				return nil, fmt.Errorf("query execution was cancelled")
			}
		}
	}
}

// ========== Query API ==========

// ListQueries returns a paginated list of queries
func (c *Client) ListQueries(ctx context.Context, page, pageSize int, search string) (*PaginatedResult[Query], error) {
	path := fmt.Sprintf("/api/queries?page=%d&page_size=%d", page, pageSize)
	if search != "" {
		path += "&q=" + url.QueryEscape(search)
	}

	var result PaginatedResult[Query]
	if err := c.do(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to list queries: %w", err)
	}
	return &result, nil
}

// GetQuery returns a specific query
func (c *Client) GetQuery(ctx context.Context, id int) (*Query, error) {
	var query Query
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/queries/%d", id), nil, &query); err != nil {
		return nil, fmt.Errorf("failed to get query: %w", err)
	}
	return &query, nil
}

// CreateQuery creates a new query
func (c *Client) CreateQuery(ctx context.Context, req *CreateQueryRequest) (*Query, error) {
	var query Query
	if err := c.do(ctx, "POST", "/api/queries", req, &query); err != nil {
		return nil, fmt.Errorf("failed to create query: %w", err)
	}
	return &query, nil
}

// UpdateQuery updates an existing query
func (c *Client) UpdateQuery(ctx context.Context, id int, req *UpdateQueryRequest) (*Query, error) {
	var query Query
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/queries/%d", id), req, &query); err != nil {
		return nil, fmt.Errorf("failed to update query: %w", err)
	}
	return &query, nil
}

// ArchiveQuery archives a query
func (c *Client) ArchiveQuery(ctx context.Context, id int) error {
	if err := c.do(ctx, "DELETE", fmt.Sprintf("/api/queries/%d", id), nil, nil); err != nil {
		return fmt.Errorf("failed to archive query: %w", err)
	}
	return nil
}

// ForkQuery forks a query
func (c *Client) ForkQuery(ctx context.Context, id int) (*Query, error) {
	var query Query
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/queries/%d/fork", id), nil, &query); err != nil {
		return nil, fmt.Errorf("failed to fork query: %w", err)
	}
	return &query, nil
}

// ExecuteQuery executes a saved query
func (c *Client) ExecuteQuery(ctx context.Context, id int, params map[string]any) (*QueryResult, error) {
	body := map[string]any{"parameters": params}

	var resp ExecuteResponse
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/queries/%d/results", id), body, &resp); err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	if resp.Job != nil {
		return c.pollJobResult(ctx, resp.Job.ID)
	}

	return resp.QueryResult, nil
}

// ExecuteAdhocQuery executes an ad-hoc query
func (c *Client) ExecuteAdhocQuery(ctx context.Context, query string, dataSourceID int) (*QueryResult, error) {
	body := map[string]any{
		"query":          query,
		"data_source_id": dataSourceID,
		"max_age":        0,
	}

	var resp ExecuteResponse
	if err := c.do(ctx, "POST", "/api/query_results", body, &resp); err != nil {
		return nil, fmt.Errorf("failed to execute adhoc query: %w", err)
	}

	if resp.Job != nil {
		return c.pollJobResult(ctx, resp.Job.ID)
	}

	return resp.QueryResult, nil
}

// GetQueryResultsCSV returns query results as CSV
func (c *Client) GetQueryResultsCSV(ctx context.Context, id int) (string, error) {
	data, err := c.doRaw(ctx, "GET", fmt.Sprintf("/api/queries/%d/results.csv", id))
	if err != nil {
		return "", fmt.Errorf("failed to get CSV results: %w", err)
	}
	return string(data), nil
}

// GetMyQueries returns queries owned by the current user
func (c *Client) GetMyQueries(ctx context.Context, page, pageSize int) (*PaginatedResult[Query], error) {
	path := fmt.Sprintf("/api/queries/my?page=%d&page_size=%d", page, pageSize)
	var result PaginatedResult[Query]
	if err := c.do(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get my queries: %w", err)
	}
	return &result, nil
}

// GetRecentQueries returns recently accessed queries
func (c *Client) GetRecentQueries(ctx context.Context, page, pageSize int) (*PaginatedResult[Query], error) {
	path := fmt.Sprintf("/api/queries/recent?page=%d&page_size=%d", page, pageSize)
	var result PaginatedResult[Query]
	if err := c.do(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get recent queries: %w", err)
	}
	return &result, nil
}

// GetQueryTags returns all query tags
func (c *Client) GetQueryTags(ctx context.Context) ([]Tag, error) {
	var resp struct {
		Tags []Tag `json:"tags"`
	}
	if err := c.do(ctx, "GET", "/api/queries/tags", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get query tags: %w", err)
	}
	return resp.Tags, nil
}

// GetFavoriteQueries returns favorite queries
func (c *Client) GetFavoriteQueries(ctx context.Context, page, pageSize int) (*PaginatedResult[Query], error) {
	path := fmt.Sprintf("/api/queries/favorites?page=%d&page_size=%d", page, pageSize)
	var result PaginatedResult[Query]
	if err := c.do(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get favorite queries: %w", err)
	}
	return &result, nil
}

// AddQueryFavorite adds a query to favorites
func (c *Client) AddQueryFavorite(ctx context.Context, id int) error {
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/queries/%d/favorite", id), nil, nil); err != nil {
		return fmt.Errorf("failed to add query to favorites: %w", err)
	}
	return nil
}

// RemoveQueryFavorite removes a query from favorites
func (c *Client) RemoveQueryFavorite(ctx context.Context, id int) error {
	if err := c.do(ctx, "DELETE", fmt.Sprintf("/api/queries/%d/favorite", id), nil, nil); err != nil {
		return fmt.Errorf("failed to remove query from favorites: %w", err)
	}
	return nil
}

// ========== Dashboard API ==========

// ListDashboards returns a paginated list of dashboards
func (c *Client) ListDashboards(ctx context.Context, page, pageSize int) (*PaginatedResult[Dashboard], error) {
	path := fmt.Sprintf("/api/dashboards?page=%d&page_size=%d", page, pageSize)
	var result PaginatedResult[Dashboard]
	if err := c.do(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to list dashboards: %w", err)
	}
	return &result, nil
}

// GetDashboard returns a specific dashboard
func (c *Client) GetDashboard(ctx context.Context, id int) (*Dashboard, error) {
	var dashboard Dashboard
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/dashboards/%d", id), nil, &dashboard); err != nil {
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}
	return &dashboard, nil
}

// CreateDashboard creates a new dashboard
func (c *Client) CreateDashboard(ctx context.Context, req *CreateDashboardRequest) (*Dashboard, error) {
	var dashboard Dashboard
	if err := c.do(ctx, "POST", "/api/dashboards", req, &dashboard); err != nil {
		return nil, fmt.Errorf("failed to create dashboard: %w", err)
	}
	return &dashboard, nil
}

// UpdateDashboard updates an existing dashboard
func (c *Client) UpdateDashboard(ctx context.Context, id int, req *UpdateDashboardRequest) (*Dashboard, error) {
	var dashboard Dashboard
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/dashboards/%d", id), req, &dashboard); err != nil {
		return nil, fmt.Errorf("failed to update dashboard: %w", err)
	}
	return &dashboard, nil
}

// ArchiveDashboard archives a dashboard
func (c *Client) ArchiveDashboard(ctx context.Context, id int) error {
	if err := c.do(ctx, "DELETE", fmt.Sprintf("/api/dashboards/%d", id), nil, nil); err != nil {
		return fmt.Errorf("failed to archive dashboard: %w", err)
	}
	return nil
}

// ForkDashboard forks a dashboard
func (c *Client) ForkDashboard(ctx context.Context, id int) (*Dashboard, error) {
	var dashboard Dashboard
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/dashboards/%d/fork", id), nil, &dashboard); err != nil {
		return nil, fmt.Errorf("failed to fork dashboard: %w", err)
	}
	return &dashboard, nil
}

// ShareDashboard creates a public link for a dashboard
func (c *Client) ShareDashboard(ctx context.Context, id int) (*ShareResponse, error) {
	var resp ShareResponse
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/dashboards/%d/share", id), nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to share dashboard: %w", err)
	}
	return &resp, nil
}

// UnshareDashboard removes the public link for a dashboard
func (c *Client) UnshareDashboard(ctx context.Context, id int) error {
	if err := c.do(ctx, "DELETE", fmt.Sprintf("/api/dashboards/%d/share", id), nil, nil); err != nil {
		return fmt.Errorf("failed to unshare dashboard: %w", err)
	}
	return nil
}

// GetPublicDashboard returns a public dashboard by token
func (c *Client) GetPublicDashboard(ctx context.Context, token string) (*Dashboard, error) {
	var dashboard Dashboard
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/dashboards/public/%s", token), nil, &dashboard); err != nil {
		return nil, fmt.Errorf("failed to get public dashboard: %w", err)
	}
	return &dashboard, nil
}

// GetMyDashboards returns dashboards owned by the current user
func (c *Client) GetMyDashboards(ctx context.Context, page, pageSize int) (*PaginatedResult[Dashboard], error) {
	path := fmt.Sprintf("/api/dashboards/my?page=%d&page_size=%d", page, pageSize)
	var result PaginatedResult[Dashboard]
	if err := c.do(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get my dashboards: %w", err)
	}
	return &result, nil
}

// GetDashboardTags returns all dashboard tags
func (c *Client) GetDashboardTags(ctx context.Context) ([]Tag, error) {
	var resp struct {
		Tags []Tag `json:"tags"`
	}
	if err := c.do(ctx, "GET", "/api/dashboards/tags", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get dashboard tags: %w", err)
	}
	return resp.Tags, nil
}

// GetFavoriteDashboards returns favorite dashboards
func (c *Client) GetFavoriteDashboards(ctx context.Context, page, pageSize int) (*PaginatedResult[Dashboard], error) {
	path := fmt.Sprintf("/api/dashboards/favorites?page=%d&page_size=%d", page, pageSize)
	var result PaginatedResult[Dashboard]
	if err := c.do(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get favorite dashboards: %w", err)
	}
	return &result, nil
}

// AddDashboardFavorite adds a dashboard to favorites
func (c *Client) AddDashboardFavorite(ctx context.Context, id int) error {
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/dashboards/%d/favorite", id), nil, nil); err != nil {
		return fmt.Errorf("failed to add dashboard to favorites: %w", err)
	}
	return nil
}

// RemoveDashboardFavorite removes a dashboard from favorites
func (c *Client) RemoveDashboardFavorite(ctx context.Context, id int) error {
	if err := c.do(ctx, "DELETE", fmt.Sprintf("/api/dashboards/%d/favorite", id), nil, nil); err != nil {
		return fmt.Errorf("failed to remove dashboard from favorites: %w", err)
	}
	return nil
}

// ========== Visualization API ==========

// GetVisualization returns a specific visualization
func (c *Client) GetVisualization(ctx context.Context, id int) (*Visualization, error) {
	var viz Visualization
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/visualizations/%d", id), nil, &viz); err != nil {
		return nil, fmt.Errorf("failed to get visualization: %w", err)
	}
	return &viz, nil
}

// CreateVisualization creates a new visualization
func (c *Client) CreateVisualization(ctx context.Context, req *CreateVisualizationRequest) (*Visualization, error) {
	var viz Visualization
	if err := c.do(ctx, "POST", "/api/visualizations", req, &viz); err != nil {
		return nil, fmt.Errorf("failed to create visualization: %w", err)
	}
	return &viz, nil
}

// UpdateVisualization updates an existing visualization
func (c *Client) UpdateVisualization(ctx context.Context, id int, req *UpdateVisualizationRequest) (*Visualization, error) {
	var viz Visualization
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/visualizations/%d", id), req, &viz); err != nil {
		return nil, fmt.Errorf("failed to update visualization: %w", err)
	}
	return &viz, nil
}

// DeleteVisualization deletes a visualization
func (c *Client) DeleteVisualization(ctx context.Context, id int) error {
	if err := c.do(ctx, "DELETE", fmt.Sprintf("/api/visualizations/%d", id), nil, nil); err != nil {
		return fmt.Errorf("failed to delete visualization: %w", err)
	}
	return nil
}

// ========== Widget API ==========

// ListWidgets returns all widgets
func (c *Client) ListWidgets(ctx context.Context) ([]Widget, error) {
	var widgets []Widget
	if err := c.do(ctx, "GET", "/api/widgets", nil, &widgets); err != nil {
		return nil, fmt.Errorf("failed to list widgets: %w", err)
	}
	return widgets, nil
}

// GetWidget returns a specific widget
func (c *Client) GetWidget(ctx context.Context, id int) (*Widget, error) {
	var widget Widget
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/widgets/%d", id), nil, &widget); err != nil {
		return nil, fmt.Errorf("failed to get widget: %w", err)
	}
	return &widget, nil
}

// CreateWidget creates a new widget
func (c *Client) CreateWidget(ctx context.Context, req *CreateWidgetRequest) (*Widget, error) {
	var widget Widget
	if err := c.do(ctx, "POST", "/api/widgets", req, &widget); err != nil {
		return nil, fmt.Errorf("failed to create widget: %w", err)
	}
	return &widget, nil
}

// UpdateWidget updates an existing widget
func (c *Client) UpdateWidget(ctx context.Context, id int, req *UpdateWidgetRequest) (*Widget, error) {
	var widget Widget
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/widgets/%d", id), req, &widget); err != nil {
		return nil, fmt.Errorf("failed to update widget: %w", err)
	}
	return &widget, nil
}

// DeleteWidget deletes a widget
func (c *Client) DeleteWidget(ctx context.Context, id int) error {
	if err := c.do(ctx, "DELETE", fmt.Sprintf("/api/widgets/%d", id), nil, nil); err != nil {
		return fmt.Errorf("failed to delete widget: %w", err)
	}
	return nil
}

// ========== Alert API ==========

// ListAlerts returns all alerts
func (c *Client) ListAlerts(ctx context.Context) ([]Alert, error) {
	var alerts []Alert
	if err := c.do(ctx, "GET", "/api/alerts", nil, &alerts); err != nil {
		return nil, fmt.Errorf("failed to list alerts: %w", err)
	}
	return alerts, nil
}

// GetAlert returns a specific alert
func (c *Client) GetAlert(ctx context.Context, id int) (*Alert, error) {
	var alert Alert
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/alerts/%d", id), nil, &alert); err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}
	return &alert, nil
}

// CreateAlert creates a new alert
func (c *Client) CreateAlert(ctx context.Context, req *CreateAlertRequest) (*Alert, error) {
	var alert Alert
	if err := c.do(ctx, "POST", "/api/alerts", req, &alert); err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}
	return &alert, nil
}

// UpdateAlert updates an existing alert
func (c *Client) UpdateAlert(ctx context.Context, id int, req *UpdateAlertRequest) (*Alert, error) {
	var alert Alert
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/alerts/%d", id), req, &alert); err != nil {
		return nil, fmt.Errorf("failed to update alert: %w", err)
	}
	return &alert, nil
}

// DeleteAlert deletes an alert
func (c *Client) DeleteAlert(ctx context.Context, id int) error {
	if err := c.do(ctx, "DELETE", fmt.Sprintf("/api/alerts/%d", id), nil, nil); err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}
	return nil
}

// MuteAlert mutes an alert
func (c *Client) MuteAlert(ctx context.Context, id int) error {
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/alerts/%d/mute", id), nil, nil); err != nil {
		return fmt.Errorf("failed to mute alert: %w", err)
	}
	return nil
}

// GetAlertSubscriptions returns subscriptions for an alert
func (c *Client) GetAlertSubscriptions(ctx context.Context, alertID int) ([]AlertSubscription, error) {
	var subs []AlertSubscription
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/alerts/%d/subscriptions", alertID), nil, &subs); err != nil {
		return nil, fmt.Errorf("failed to get alert subscriptions: %w", err)
	}
	return subs, nil
}

// AddAlertSubscription adds a subscription to an alert
func (c *Client) AddAlertSubscription(ctx context.Context, alertID int, req *CreateAlertSubscriptionRequest) (*AlertSubscription, error) {
	var sub AlertSubscription
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/alerts/%d/subscriptions", alertID), req, &sub); err != nil {
		return nil, fmt.Errorf("failed to add alert subscription: %w", err)
	}
	return &sub, nil
}

// RemoveAlertSubscription removes a subscription from an alert
func (c *Client) RemoveAlertSubscription(ctx context.Context, alertID, subscriptionID int) error {
	path := fmt.Sprintf("/api/alerts/%d/subscriptions/%d", alertID, subscriptionID)
	if err := c.do(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to remove alert subscription: %w", err)
	}
	return nil
}

// ========== Query Snippet API ==========

// ListQuerySnippets returns all query snippets
func (c *Client) ListQuerySnippets(ctx context.Context) ([]QuerySnippet, error) {
	var snippets []QuerySnippet
	if err := c.do(ctx, "GET", "/api/query_snippets", nil, &snippets); err != nil {
		return nil, fmt.Errorf("failed to list query snippets: %w", err)
	}
	return snippets, nil
}

// GetQuerySnippet returns a specific query snippet
func (c *Client) GetQuerySnippet(ctx context.Context, id int) (*QuerySnippet, error) {
	var snippet QuerySnippet
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/query_snippets/%d", id), nil, &snippet); err != nil {
		return nil, fmt.Errorf("failed to get query snippet: %w", err)
	}
	return &snippet, nil
}

// CreateQuerySnippet creates a new query snippet
func (c *Client) CreateQuerySnippet(ctx context.Context, req *CreateQuerySnippetRequest) (*QuerySnippet, error) {
	var snippet QuerySnippet
	if err := c.do(ctx, "POST", "/api/query_snippets", req, &snippet); err != nil {
		return nil, fmt.Errorf("failed to create query snippet: %w", err)
	}
	return &snippet, nil
}

// UpdateQuerySnippet updates an existing query snippet
func (c *Client) UpdateQuerySnippet(ctx context.Context, id int, req *UpdateQuerySnippetRequest) (*QuerySnippet, error) {
	var snippet QuerySnippet
	if err := c.do(ctx, "POST", fmt.Sprintf("/api/query_snippets/%d", id), req, &snippet); err != nil {
		return nil, fmt.Errorf("failed to update query snippet: %w", err)
	}
	return &snippet, nil
}

// DeleteQuerySnippet deletes a query snippet
func (c *Client) DeleteQuerySnippet(ctx context.Context, id int) error {
	if err := c.do(ctx, "DELETE", fmt.Sprintf("/api/query_snippets/%d", id), nil, nil); err != nil {
		return fmt.Errorf("failed to delete query snippet: %w", err)
	}
	return nil
}

// ========== DataSource API ==========

// ListDataSources returns all data sources
func (c *Client) ListDataSources(ctx context.Context) ([]DataSource, error) {
	var dataSources []DataSource
	if err := c.do(ctx, "GET", "/api/data_sources", nil, &dataSources); err != nil {
		return nil, fmt.Errorf("failed to list data sources: %w", err)
	}
	return dataSources, nil
}

// GetDataSourceSchema returns the schema for a data source
func (c *Client) GetDataSourceSchema(ctx context.Context, id int) (*DataSourceSchema, error) {
	var schema DataSourceSchema
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/data_sources/%d/schema", id), nil, &schema); err != nil {
		return nil, fmt.Errorf("failed to get data source schema: %w", err)
	}
	return &schema, nil
}

// ========== Destination API ==========

// ListDestinations returns all notification destinations
func (c *Client) ListDestinations(ctx context.Context) ([]Destination, error) {
	var destinations []Destination
	if err := c.do(ctx, "GET", "/api/destinations", nil, &destinations); err != nil {
		return nil, fmt.Errorf("failed to list destinations: %w", err)
	}
	return destinations, nil
}

// Helper to convert int to string for path params
func itoa(i int) string {
	return strconv.Itoa(i)
}
