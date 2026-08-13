//go:build integration

package repository_test

import "testing"

// Integration tests (testcontainers) for Billing/Shop/Inventory live behind this build tag.
// Run: go test -tags=integration ./internal/repository/...
// Requires Docker. Placeholder keeps the FINAL_PRIORITY checklist path explicit.
func TestIntegrationPlaceholder(t *testing.T) {
	t.Skip("wire testcontainers postgres+nats here when Docker is available in CI")
}
