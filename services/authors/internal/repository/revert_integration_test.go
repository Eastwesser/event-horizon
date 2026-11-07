//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/Eastwesser/event-horizon/services/authors/internal/repository"
	"github.com/google/uuid"
)

func TestRevertApplication_SoftDeletesAuthor(t *testing.T) {
	pool, cleanup := authorsPool(t)
	defer cleanup()

	ctx := context.Background()
	repo := repository.NewPostgresRepo(pool)

	userID := "u-revert-1"       // user_id — text, не uuid, тут всё ок
	authorID := uuid.NewString() // id — uuid, нужен валидный UUID
	appID := uuid.NewString()    // application id — тоже uuid (проверь \d author_applications)

	if _, err := pool.Exec(ctx, `INSERT INTO authors (id, user_id, display_name, active) VALUES ($1,$2,$3,true)`,
		authorID, userID, "Den"); err != nil {
		t.Fatalf("insert author: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO author_applications (id, user_id, status, payload) VALUES ($1,$2,'approved','{}')`,
		appID, userID); err != nil {
		t.Fatalf("insert application: %v", err)
	}

	if _, err := repo.RevertApplication(ctx, appID); err != nil {
		t.Fatalf("revert: %v", err)
	}

	var active bool
	if err := pool.QueryRow(ctx, `SELECT active FROM authors WHERE user_id = $1`, userID).Scan(&active); err != nil {
		t.Fatalf("author row must survive soft-delete: %v", err)
	}
	if active {
		t.Fatal("author must be deactivated, not deleted")
	}
}