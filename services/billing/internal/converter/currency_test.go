package converter

import (
	"testing"

	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
	pb "github.com/Eastwesser/event-horizon/services/billing/proto"
)

func TestCurrencyConverters(t *testing.T) {
	if ProtoCurrencyToRepo(pb.CurrencyType_TICKETS) != repository.Tickets {
		t.Fatal("tickets")
	}
	if RepoCurrencyToProto(repository.Lamps) != pb.CurrencyType_LAMPS {
		t.Fatal("lamps")
	}
}
