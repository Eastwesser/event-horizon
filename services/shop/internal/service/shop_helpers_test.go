package service

import (
	"regexp"
	"testing"
)

func TestNewEventID_Format(t *testing.T) {
	re := regexp.MustCompile(`^[0-9a-f]{32}$`)
	for i := 0; i < 5; i++ {
		id := newEventID()
		if !re.MatchString(id) {
			t.Fatalf("bad id %q", id)
		}
	}
}
