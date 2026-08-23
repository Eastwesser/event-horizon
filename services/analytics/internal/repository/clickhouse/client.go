package clickhouse

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// identifierOK restricts DB/table names used as SQL identifiers (never user input).
var identifierOK = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Client talks to ClickHouse over HTTP using named query parameters
// ({name:Type} + param_name=…) — never string-concatenate user values into SQL.
type Client struct {
	base string
	db   string
	http *http.Client
}

func New(baseURL, db string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if db == "" {
		db = "default"
	}
	return &Client{
		base: baseURL,
		db:   db,
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) DB() string { return c.db }

func (c *Client) Ping(ctx context.Context) error {
	// Use system DB so cold starts work before EnsureSchema creates c.db.
	_, err := c.execDB(ctx, "default", "SELECT 1", nil)
	return err
}

func (c *Client) EnsureSchema(ctx context.Context) error {
	if !identifierOK.MatchString(c.db) {
		return fmt.Errorf("invalid clickhouse database name %q", c.db)
	}
	if _, err := c.execDB(ctx, "default", fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s`, c.db), nil); err != nil {
		return err
	}
	ddl := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.analytics_events (
			event_time DateTime,
			event_date Date DEFAULT toDate(event_time),
			user_id String,
			event_type String,
			payload String
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(event_date)
		ORDER BY (event_date, event_type, user_id, event_time)`, c.db)
	_, err := c.Exec(ctx, ddl, nil)
	return err
}

func (c *Client) InsertEvent(ctx context.Context, userID, eventType, payload string, at time.Time) error {
	if at.IsZero() {
		at = time.Now().UTC()
	}
	if payload == "" {
		payload = "{}"
	}
	if !identifierOK.MatchString(c.db) {
		return fmt.Errorf("invalid clickhouse database name %q", c.db)
	}
	sql := fmt.Sprintf(`INSERT INTO %s.analytics_events (event_time, user_id, event_type, payload)
		VALUES ({event_time:String}, {user_id:String}, {event_type:String}, {payload:String})`, c.db)
	params := map[string]string{
		"event_time": at.UTC().Format("2006-01-02 15:04:05"),
		"user_id":    userID,
		"event_type": eventType,
		"payload":    payload,
	}
	_, err := c.Exec(ctx, sql, params)
	return err
}

func (c *Client) QueryTSV(ctx context.Context, sql string, params map[string]string) (string, error) {
	return c.Exec(ctx, sql+" FORMAT TabSeparated", params)
}

// Exec runs SQL with named ClickHouse HTTP parameters (param_<name>).
func (c *Client) Exec(ctx context.Context, sql string, params map[string]string) (string, error) {
	return c.execDB(ctx, c.db, sql, params)
}

func (c *Client) execDB(ctx context.Context, database, sql string, params map[string]string) (string, error) {
	u, err := url.Parse(c.base)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("database", database)
	for k, v := range params {
		q.Set("param_"+k, v)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(sql))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("clickhouse %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return string(body), nil
}

// BuildParamURL is exported for unit tests: proves values land in query params, not SQL body.
func BuildParamURL(base, database, sql string, params map[string]string) (string, string, error) {
	u, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil {
		return "", "", err
	}
	q := u.Query()
	q.Set("database", database)
	for k, v := range params {
		q.Set("param_"+k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), sql, nil
}

// IntParam formats an int for ClickHouse Int32/UInt32 named params.
func IntParam(n int) string { return strconv.Itoa(n) }
