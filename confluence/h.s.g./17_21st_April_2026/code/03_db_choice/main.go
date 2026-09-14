// DB choice cheat-sheet — сценарий → инструмент (для устного ответа).
package main

import "fmt"

func choose(scenario string) string {
	m := map[string]string{
		"oltp orders":     "PostgreSQL",
		"time series iot": "TimeScaleDB / VictoriaMetrics",
		"logs analytics":  "ClickHouse (columnar OLAP)",
		"wide write kv":   "Cassandra / Scylla",
		"friends graph":   "Neo4j",
		"full text":       "Elasticsearch",
		"hot cache":       "Redis / Dragonfly",
	}
	if v, ok := m[scenario]; ok {
		return v
	}
	return "уточни read/write, объём, latency"
}

func main() {
	for _, s := range []string{"oltp orders", "logs analytics", "friends graph"} {
		fmt.Printf("%s → %s\n", s, choose(s))
	}
}
