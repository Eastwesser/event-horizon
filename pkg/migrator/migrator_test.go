package migrator_test

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Eastwesser/event-horizon/pkg/migrator"
)

func TestExtractViaEmptyApplyPath(t *testing.T) {
	// Ensure package is importable and FS typing works.
	var migrations fs.FS = fstest.MapFS{
		"20260101000000_init.sql": &fstest.MapFile{Data: []byte("-- +goose Up\nSELECT 1;\n-- +goose Down\nSELECT 0;\n")},
	}
	entries, err := fs.ReadDir(migrations, ".")
	if err != nil || len(entries) != 1 {
		t.Fatal(err, entries)
	}
	_ = migrator.Up
	if !strings.Contains(entries[0].Name(), "init") {
		t.Fatal(entries[0].Name())
	}
}
