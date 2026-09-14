// Учебный разбор code-review из 04_full_task: список багов + безопасный паттерн без реальной БД.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

type Record struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func listReviewFindings() {
	findings := []string{
		`json tags must use backticks: json:"id"`,
		"loop variable capture: pass r as arg or use Go ≥1.22",
		"SQL injection: fmt.Sprintf into query → use $1 placeholders",
		"secrets in DSN → env/vault",
		"context.TODO → WithTimeout",
		"response.Body.Close() then ReadAll → close after read / defer",
		"http.Client without Timeout",
		"r.name lowercase — field is Name",
	}
	fmt.Println("code-review findings:")
	for i, f := range findings {
		fmt.Printf("  %d. %s\n", i+1, f)
	}
}

func fixedDemo() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Println("server got:", string(body))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	records := []Record{{1, "John"}, {2, "Jane"}, {3, "Mike"}}
	var mu sync.Mutex
	var names []string
	var wg sync.WaitGroup
	wg.Add(len(records))
	for _, r := range records {
		r := r
		go func() {
			defer wg.Done()
			name := "Updated " + r.Name
			// вместо SQL: безопасная «вставка»
			mu.Lock()
			names = append(names, name)
			mu.Unlock()
		}()
	}
	wg.Wait()

	payload, _ := json.Marshal(names)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, srv.URL+"/endpoint", bytes.NewReader(payload))
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("client status:", resp.StatusCode, "body:", string(body))
}

func main() {
	listReviewFindings()
	fixedDemo()
}
