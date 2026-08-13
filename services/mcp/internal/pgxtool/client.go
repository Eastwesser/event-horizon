package pgxtool

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, dsn string) (*Client, error) {
	if dsn == "" {
		return nil, fmt.Errorf("MCP_POSTGRES_DSN / DATABASE_URL not set")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 4
	cfg.MinConns = 0
	cfg.MaxConnLifetime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Client{pool: pool}, nil
}

func (c *Client) Close() {
	if c != nil && c.pool != nil {
		c.pool.Close()
	}
}

// ValidateSelect rejects anything that is not a single SELECT/WITH query.
func ValidateSelect(sql string) error {
	s := strings.TrimSpace(sql)
	if s == "" {
		return fmt.Errorf("empty query")
	}
	if strings.Contains(s, ";") {
		return fmt.Errorf("multiple statements / trailing semicolon not allowed")
	}
	lower := strings.ToLower(s)
	// strip leading comments
	for strings.HasPrefix(lower, "--") || strings.HasPrefix(lower, "/*") {
		return fmt.Errorf("comments not allowed in MCP queries")
	}
	if !(strings.HasPrefix(lower, "select") || strings.HasPrefix(lower, "with")) {
		return fmt.Errorf("only SELECT / WITH (read-only) queries are allowed")
	}
	forbidden := []string{
		" insert ", " update ", " delete ", " drop ", " alter ", " create ",
		" truncate ", " grant ", " revoke ", " copy ", " call ", " do ",
		" into ", // blocks SELECT INTO
	}
	padded := " " + lower + " "
	for _, f := range forbidden {
		if strings.Contains(padded, f) {
			return fmt.Errorf("forbidden keyword in query: %s", strings.TrimSpace(f))
		}
	}
	return nil
}

func (c *Client) Query(ctx context.Context, sql string, limit int) (string, error) {
	if c == nil || c.pool == nil {
		return "", fmt.Errorf("postgres not connected")
	}
	if err := ValidateSelect(sql); err != nil {
		return "", err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	// wrap to enforce row cap without rewriting user SQL semantics too hard
	wrapped := fmt.Sprintf("SELECT * FROM (%s) AS _mcp_q LIMIT %d", sql, limit)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := c.pool.Query(ctx, wrapped)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	cols := make([]string, len(fields))
	for i, f := range fields {
		cols[i] = string(f.Name)
	}

	var out strings.Builder
	out.WriteString(strings.Join(cols, "\t"))
	out.WriteByte('\n')
	n := 0
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return "", err
		}
		parts := make([]string, len(vals))
		for i, v := range vals {
			parts[i] = fmt.Sprint(v)
		}
		out.WriteString(strings.Join(parts, "\t"))
		out.WriteByte('\n')
		n++
	}
	out.WriteString(fmt.Sprintf("\n(%d rows)\n", n))
	return out.String(), rows.Err()
}
