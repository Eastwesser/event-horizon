// Package migrator applies Goose-style SQL migrations without the goose dependency.
// It tracks versions in goose_db_version so `make migrate-*` (goose CLI) stays compatible.
package migrator

import (
	"database/sql"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var versionRe = regexp.MustCompile(`^(\d+)_`)

// Up applies all pending -- +goose Up sections from migrationsFS (*.sql).
func Up(db *sql.DB, migrationsFS fs.FS) error {
	if err := ensureVersionTable(db); err != nil {
		return err
	}
	applied, err := appliedVersions(db)
	if err != nil {
		return err
	}

	entries, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	type file struct {
		name    string
		version int64
	}
	var files []file
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		m := versionRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		v, _ := strconv.ParseInt(m[1], 10, 64)
		files = append(files, file{name: e.Name(), version: v})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].version < files[j].version })

	for _, f := range files {
		if applied[f.version] {
			continue
		}
		body, err := fs.ReadFile(migrationsFS, path.Clean(f.name))
		if err != nil {
			return err
		}
		upSQL := extractUp(string(body))
		if strings.TrimSpace(upSQL) == "" {
			continue
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(upSQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s: %w", f.name, err)
		}
		if _, err := tx.Exec(
			`INSERT INTO goose_db_version (version_id, is_applied) VALUES ($1, TRUE)`,
			f.version,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", f.name, err)
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func ensureVersionTable(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS goose_db_version (
    id SERIAL PRIMARY KEY,
    version_id BIGINT NOT NULL,
    is_applied BOOLEAN NOT NULL,
    tstamp TIMESTAMP DEFAULT NOW()
)`)
	return err
}

func appliedVersions(db *sql.DB) (map[int64]bool, error) {
	rows, err := db.Query(`SELECT version_id FROM goose_db_version WHERE is_applied = TRUE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int64]bool)
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

func extractUp(sqlText string) string {
	const upMark = "-- +goose Up"
	const downMark = "-- +goose Down"
	i := strings.Index(sqlText, upMark)
	if i < 0 {
		// no markers — treat whole file as up
		return sqlText
	}
	rest := sqlText[i+len(upMark):]
	if j := strings.Index(rest, downMark); j >= 0 {
		rest = rest[:j]
	}
	return rest
}
