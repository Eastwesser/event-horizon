// Исправленный учебный вариант currency CLI из code-review (01_analyze_task).
// Без реальной БД: in-memory store. Комментарии = что сказать на ревью.
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type bankJob struct {
	name, from, to, path string
	needsAuth            bool
}

type store struct {
	mu    sync.Mutex
	rates []string
}

func (s *store) insert(bank, from, to string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// parameterized insert в проде: db.ExecContext(ctx, `INSERT ... VALUES ($1,$2,$3,$4)`, ...)
	s.rates = append(s.rates, fmt.Sprintf("%s %s→%s = %.2f", bank, from, to, value))
}

func fetchRate(ctx context.Context, client *http.Client, job bankJob, base string, token string) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+job.path, nil)
	if err != nil {
		return 0, err
	}
	if job.needsAuth {
		req.Header.Set("Authorization", "Bearer "+token) // токен из env, не в коде
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	raw := string(body)
	if job.name == "Bank 1" {
		raw = strings.ReplaceAll(raw, ",", ".") // локаль банка: запятая как decimal
	}
	return strconv.ParseFloat(strings.TrimSpace(raw), 64)
}

func runUpdate(ctx context.Context, st *store) error {
	// httptest имитирует банки; в проде — разные URL из конфига
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/b1":
			fmt.Fprint(w, "92,50")
		case "/b2":
			fmt.Fprint(w, "93.10")
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	jobs := []bankJob{
		{"Bank 1", "RUB", "USD", "/b1", false},
		{"Bank 2", "RUB", "USD", "/b2", true},
	}
	client := &http.Client{Timeout: 5 * time.Second} // не http.DefaultClient
	token := os.Getenv("BANK_TOKEN")
	if token == "" {
		token = "demo"
	}

	for _, job := range jobs {
		val, err := fetchRate(ctx, client, job, srv.URL, token)
		if err != nil {
			return fmt.Errorf("%s: %w", job.name, err) // wrap, не panic
		}
		// БАГ в оригинале: curFrom, curFrom — здесь to корректный
		st.insert(job.name, job.from, job.to, val)
	}
	return nil
}

func main() {
	// без аргумента — demo update (на ревью: guard clause + subcommands)
	if len(os.Args) >= 2 && os.Args[1] == "help" {
		fmt.Println("Usage: go run main.go update")
		return
	}
	if len(os.Args) >= 2 && os.Args[1] != "update" {
		fmt.Println("Usage: go run main.go update")
		os.Exit(1)
	}
	st := &store{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := runUpdate(ctx, st); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	for _, line := range st.rates {
		fmt.Println(line)
	}
}
