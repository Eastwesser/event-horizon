package rag

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearchFindsOutbox(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "4.architecture_patterns")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `# Integration patterns — Event Horizon

## Outbox (transactional messaging)

Shop writes purchase and outbox in one Postgres transaction, then a worker publishes shop.purchased to NATS.
`
	if err := os.WriteFile(filepath.Join(sub, "03_ARCH_INTEGRATION_PATTERNS.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	idx, err := BuildFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	hits := idx.Search("how does Outbox work in Shop", 3)
	if len(hits) == 0 {
		t.Fatal("expected at least one hit")
	}
	if hits[0].Path == "" || hits[0].Score <= 0 {
		t.Fatalf("bad hit: %+v", hits[0])
	}
}
