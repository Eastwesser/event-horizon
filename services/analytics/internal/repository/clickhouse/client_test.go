package clickhouse

import (
	"net/url"
	"strings"
	"testing"
)

func TestBuildParamURL_InjectionStaysInParamsNotSQL(t *testing.T) {
	evil := `'; DROP TABLE analytics_events; --`
	sql := `INSERT INTO eventhorizon.analytics_events (event_time, user_id, event_type, payload)
		VALUES ({event_time:String}, {user_id:String}, {event_type:String}, {payload:String})`
	params := map[string]string{
		"event_time": "2026-08-13 12:00:00",
		"user_id":    evil,
		"event_type": "score.updated",
		"payload":    `{"x":1}`,
	}
	reqURL, body, err := BuildParamURL("http://localhost:8123", "eventhorizon", sql, params)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(body, evil) {
		t.Fatalf("evil payload must not appear in SQL body, got: %s", body)
	}
	if !strings.Contains(body, "{user_id:String}") {
		t.Fatalf("expected named placeholder in SQL, got: %s", body)
	}
	u, err := url.Parse(reqURL)
	if err != nil {
		t.Fatal(err)
	}
	got := u.Query().Get("param_user_id")
	if got != evil {
		t.Fatalf("param_user_id=%q want %q", got, evil)
	}
	if strings.Contains(u.Path, "DROP") {
		t.Fatal("DROP must not appear in URL path")
	}
}

func TestInvalidDatabaseNameRejected(t *testing.T) {
	c := New("http://localhost:8123", "eh; DROP DATABASE x")
	if err := c.EnsureSchema(t.Context()); err == nil {
		t.Fatal("expected invalid database name error")
	}
}
