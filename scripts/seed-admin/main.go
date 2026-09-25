package main

// Dev-only idempotent admin god account seed.
// Usage:
//   make seed-admin
//   # or: go run ./scripts/seed-admin
//
// Credentials come from env / scripts/.env.seed.admin (gitignored) — never the frontend.

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"'`)
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
}

func mustOpen(dsn string) *sql.DB {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		fatalf("open db: %v", err)
	}
	db.SetConnMaxLifetime(time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		fatalf("ping db: %v\nDSN host tip: is docker compose up?", err)
	}
	return db
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "seed-admin: "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	if getenv("ALLOW_DEV_SEED", "") != "1" && getenv("APP_ENV", "dev") == "production" {
		fatalf("refusing to run: set ALLOW_DEV_SEED=1 and never use APP_ENV=production")
	}

	loadDotEnv("scripts/.env.seed.admin")

	email := getenv("SEED_ADMIN_EMAIL", "admin@eventhorizon.local")
	password := getenv("SEED_ADMIN_PASSWORD", "")
	if password == "" {
		fatalf("SEED_ADMIN_PASSWORD is empty — copy scripts/.env.seed.admin.example → scripts/.env.seed.admin")
	}
	lamps := getenv("SEED_ADMIN_LAMPS", "1000000")
	tickets := getenv("SEED_ADMIN_TICKETS", "1000000")

	authDSN := getenv(
		"SEED_AUTH_DSN",
		"postgres://eventhorizon:eventhorizon@localhost:5460/eventhorizon?sslmode=disable",
	)
	billingDSN := getenv(
		"SEED_BILLING_DSN",
		"postgres://eventhorizon:eventhorizon@localhost:5462/eventhorizon_billing?sslmode=disable",
	)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		fatalf("bcrypt: %v", err)
	}

	auth := mustOpen(authDSN)
	defer auth.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var userID string
	err = auth.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, 'admin')
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			role = 'admin',
			updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`, email, string(hash)).Scan(&userID)
	if err != nil {
		fatalf("upsert auth user: %v", err)
	}

	billing := mustOpen(billingDSN)
	defer billing.Close()

	for _, row := range []struct {
		currency string
		balance  string
	}{
		{"lamps", lamps},
		{"tickets", tickets},
	} {
		_, err := billing.ExecContext(ctx, `
			INSERT INTO user_currencies (user_id, currency_type, balance, updated_at)
			VALUES ($1::uuid, $2, $3::int, CURRENT_TIMESTAMP)
			ON CONFLICT (user_id, currency_type) DO UPDATE SET
				balance = EXCLUDED.balance,
				updated_at = CURRENT_TIMESTAMP
		`, userID, row.currency, row.balance)
		if err != nil {
			fatalf("upsert %s balance: %v", row.currency, err)
		}
	}

	fmt.Println("seed-admin: ok")
	fmt.Printf("  user_id:  %s\n", userID)
	fmt.Printf("  email:    %s\n", email)
	fmt.Printf("  role:     admin\n")
	fmt.Printf("  lamps:    %s\n", lamps)
	fmt.Printf("  tickets:  %s\n", tickets)
	fmt.Println("  password: (from SEED_ADMIN_PASSWORD / scripts/.env.seed.admin — not printed)")
	fmt.Println("Re-run: make seed-admin   (idempotent upsert)")
}
