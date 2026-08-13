package model

import "testing"

func TestErrVersionConflict(t *testing.T) {
	if ErrVersionConflict == nil || ErrVersionConflict.Error() == "" {
		t.Fatal("ErrVersionConflict must be defined")
	}
}
