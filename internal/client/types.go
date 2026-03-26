package client

import "time"

// PaginatedResult represents a paginated API response
type PaginatedResult[T any] struct {
	Count    int `json:"count"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Results  []T `json:"results"`
}

// Query represents a Redash query
type Query struct {
	ID                int             `json:"id"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	Query             string          `json:"query"`
	DataSourceID      int             `json:"data_source_id"`
	LatestResultID    *int            `json:"latest_query_data_id"`
	IsArchived        bool            `json:"is_archived"`
	IsDraft           bool            `json:"is_draft"`
	IsSafe            bool            `json:"is_safe"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	Runtime           float64         `json:"runtime"`
	RetrievedAt       *time.Time      `json:"retrieved_at"`
	Schedule          *Schedule       `json:"schedule"`
	Options           QueryOptions    `json:"options"`
	Tags              []string        `json:"tags"`
	Visualizations    []Visualization `json:"visualizations"`
	User              *User           `json:"user"`
	LastModifiedBy    *User           `json:"last_modified_by"`
	CanEdit           bool            `json:"can_edit"`
	IsFavorite        bool            `json:"is_favorite"`
	APIKey            string          `json:"api_key"`
	Version           int             `json:"version"`
	PermissionVisible bool            `json:"permission_visible"`
}

// CreateQueryRequest represents a request to create a query
type CreateQueryRequest struct {
	Name         string        `json:"name"`
	DataSourceID int           `json:"data_source_id"`
	Query        string        `json:"query"`
	Description  string        `json:"description,omitempty"`
	Options      *QueryOptions `json:"options,omitempty"`
	Schedule     *Schedule     `json:"schedule,omitempty"`
	Tags         []string      `json:"tags,omitempty"`
}

// UpdateQueryRequest represents a request to update a query
type UpdateQueryRequest struct {
	Name         string        `json:"name,omitempty"`
	DataSourceID int           `json:"data_source_id,omitempty"`
	Query        string        `json:"query,omitempty"`
	Description  string        `json:"description,omitempty"`
	Options      *QueryOptions `json:"options,omitempty"`
	Schedule     *Schedule     `json:"schedule,omitempty"`
	Tags         []string      `json:"tags,omitempty"`
	IsArchived   *bool         `json:"is_archived,omitempty"`
	IsDraft      *bool         `json:"is_draft,omitempty"`
}

// QueryOptions represents query options
type QueryOptions struct {
	Parameters []QueryParameter `json:"parameters,omitempty"`
}

// QueryParameter represents a query parameter
type QueryParameter struct {
	Name         string `json:"name"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	Value        any    `json:"value"`
	EnumOptions  string `json:"enumOptions,omitempty"`
	QueryID      int    `json:"queryId,omitempty"`
	GlobalParam  bool   `json:"global,omitempty"`
	ParentQueryID int   `json:"parentQueryId,omitempty"`
}

// Schedule represents a query schedule
type Schedule struct {
	Interval  int     `json:"interval"`
	Time      string  `json:"time,omitempty"`
	DayOfWeek string  `json:"day_of_week,omitempty"`
	Until     *string `json:"until,omitempty"`
}

// QueryResult represents query execution results
type QueryResult struct {
	ID           int         `json:"id"`
	QueryID      int         `json:"query_id"`
	DataSourceID int         `json:"data_source_id"`
	QueryHash    string      `json:"query_hash"`
	Query        string      `json:"query"`
	Data         *ResultData `json:"data"`
	Runtime      float64     `json:"runtime"`
	RetrievedAt  time.Time   `json:"retrieved_at"`
}

// ResultData represents the data portion of query results
type ResultData struct {
	Columns []Column         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
}

// Column represents a column in query results
type Column struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	FriendlyName string `json:"friendly_name"`
}

// Dashboard represents a Redash dashboard
type Dashboard struct {
	ID                      int       `json:"id"`
	Slug                    string    `json:"slug"`
	Name                    string    `json:"name"`
	Tags                    []string  `json:"tags"`
	IsArchived              bool      `json:"is_archived"`
	IsDraft                 bool      `json:"is_draft"`
	IsFavorite              bool      `json:"is_favorite"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	Version                 int       `json:"version"`
	DashboardFiltersEnabled bool      `json:"dashboard_filters_enabled"`
	Widgets                 []Widget  `json:"widgets"`
	User                    *User     `json:"user"`
	CanEdit                 bool      `json:"can_edit"`
	PublicURL               string    `json:"public_url,omitempty"`
	APIKey                  string    `json:"api_key,omitempty"`
}

// CreateDashboardRequest represents a request to create a dashboard
type CreateDashboardRequest struct {
	Name string   `json:"name"`
	Tags []string `json:"tags,omitempty"`
}

// UpdateDashboardRequest represents a request to update a dashboard
type UpdateDashboardRequest struct {
	Name                    string   `json:"name,omitempty"`
	Tags                    []string `json:"tags,omitempty"`
	IsArchived              *bool    `json:"is_archived,omitempty"`
	IsDraft                 *bool    `json:"is_draft,omitempty"`
	DashboardFiltersEnabled *bool    `json:"dashboard_filters_enabled,omitempty"`
}

// Widget represents a dashboard widget
type Widget struct {
	ID              int                    `json:"id"`
	DashboardID     int                    `json:"dashboard_id"`
	VisualizationID *int                   `json:"visualization_id"`
	Visualization   *Visualization         `json:"visualization"`
	Text            string                 `json:"text"`
	Width           int                    `json:"width"`
	Options         map[string]any         `json:"options"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// CreateWidgetRequest represents a request to create a widget
type CreateWidgetRequest struct {
	DashboardID     int            `json:"dashboard_id"`
	VisualizationID *int           `json:"visualization_id,omitempty"`
	Text            string         `json:"text,omitempty"`
	Width           int            `json:"width"`
	Options         map[string]any `json:"options,omitempty"`
}

// UpdateWidgetRequest represents a request to update a widget
type UpdateWidgetRequest struct {
	Text    string         `json:"text,omitempty"`
	Width   int            `json:"width,omitempty"`
	Options map[string]any `json:"options,omitempty"`
}

// Visualization represents a query visualization
type Visualization struct {
	ID          int            `json:"id"`
	QueryID     int            `json:"query_id"`
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Options     map[string]any `json:"options"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// CreateVisualizationRequest represents a request to create a visualization
type CreateVisualizationRequest struct {
	QueryID     int            `json:"query_id"`
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Options     map[string]any `json:"options"`
}

// UpdateVisualizationRequest represents a request to update a visualization
type UpdateVisualizationRequest struct {
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Options     map[string]any `json:"options,omitempty"`
}

// Alert represents a Redash alert
type Alert struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	QueryID     int          `json:"query_id"`
	Query       *Query       `json:"query"`
	Options     AlertOptions `json:"options"`
	State       string       `json:"state"`
	Rearm       *int         `json:"rearm"`
	LastTriggeredAt *time.Time `json:"last_triggered_at"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	User        *User        `json:"user"`
}

// AlertOptions represents alert options
type AlertOptions struct {
	Column    string `json:"column"`
	Op        string `json:"op"`
	Value     any    `json:"value"`
	Muted     bool   `json:"muted"`
}

// CreateAlertRequest represents a request to create an alert
type CreateAlertRequest struct {
	Name    string       `json:"name"`
	QueryID int          `json:"query_id"`
	Options AlertOptions `json:"options"`
	Rearm   *int         `json:"rearm,omitempty"`
}

// UpdateAlertRequest represents a request to update an alert
type UpdateAlertRequest struct {
	Name    string        `json:"name,omitempty"`
	Options *AlertOptions `json:"options,omitempty"`
	Rearm   *int          `json:"rearm,omitempty"`
}

// AlertSubscription represents an alert subscription
type AlertSubscription struct {
	ID            int          `json:"id"`
	AlertID       int          `json:"alert_id"`
	DestinationID *int         `json:"destination_id"`
	Destination   *Destination `json:"destination"`
	User          *User        `json:"user"`
}

// CreateAlertSubscriptionRequest represents a request to create an alert subscription
type CreateAlertSubscriptionRequest struct {
	DestinationID *int `json:"destination_id,omitempty"`
}

// Destination represents a notification destination
type Destination struct {
	ID      int            `json:"id"`
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	Options map[string]any `json:"options"`
}

// QuerySnippet represents a query snippet
type QuerySnippet struct {
	ID          int       `json:"id"`
	Trigger     string    `json:"trigger"`
	Description string    `json:"description"`
	Snippet     string    `json:"snippet"`
	User        *User     `json:"user"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateQuerySnippetRequest represents a request to create a query snippet
type CreateQuerySnippetRequest struct {
	Trigger     string `json:"trigger"`
	Description string `json:"description,omitempty"`
	Snippet     string `json:"snippet"`
}

// UpdateQuerySnippetRequest represents a request to update a query snippet
type UpdateQuerySnippetRequest struct {
	Trigger     string `json:"trigger,omitempty"`
	Description string `json:"description,omitempty"`
	Snippet     string `json:"snippet,omitempty"`
}

// DataSource represents a Redash data source
type DataSource struct {
	ID                  int            `json:"id"`
	Name                string         `json:"name"`
	Type                string         `json:"type"`
	Syntax              string         `json:"syntax"`
	Paused              int            `json:"paused"`
	PauseReason         string         `json:"pause_reason"`
	SupportsAutoLimit   bool           `json:"supports_auto_limit"`
	Options             map[string]any `json:"options"`
	ScheduledQueueName  string         `json:"scheduled_queue_name"`
	QueueName           string         `json:"queue_name"`
	ViewOnly            bool           `json:"view_only"`
	Groups              map[string]bool `json:"groups"`
}

// DataSourceSchema represents a data source schema
type DataSourceSchema struct {
	Schema []SchemaTable `json:"schema"`
}

// SchemaTable represents a table in the schema
type SchemaTable struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
}

// User represents a Redash user
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Job represents an async job
type Job struct {
	ID            string `json:"id"`
	Status        int    `json:"status"`
	Error         string `json:"error"`
	QueryResultID int    `json:"query_result_id"`
}

// JobStatus constants
const (
	JobStatusPending   = 1
	JobStatusStarted   = 2
	JobStatusSuccess   = 3
	JobStatusFailure   = 4
	JobStatusCancelled = 5
)

// Tag represents a tag with count
type Tag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ExecuteResponse represents the response from executing a query
type ExecuteResponse struct {
	Job         *Job         `json:"job,omitempty"`
	QueryResult *QueryResult `json:"query_result,omitempty"`
}

// ShareResponse represents the response from sharing a dashboard
type ShareResponse struct {
	PublicURL string `json:"public_url"`
	APIKey    string `json:"api_key"`
}
