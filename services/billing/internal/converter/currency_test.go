package converter

import (
	"testing"

	"github.com/Eastwesser/event-horizon/services/billing/internal/model"
	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
	pb "github.com/Eastwesser/event-horizon/services/billing/proto"
)

func TestCurrencyConverters(t *testing.T) {
	if ProtoCurrencyToRepo(pb.CurrencyType_TICKETS) != repository.Tickets {
		t.Fatal("tickets")
	}
	if ProtoCurrencyToRepo(pb.CurrencyType_LAMPS) != repository.Lamps {
		t.Fatal("lamps")
	}
	if ProtoCurrencyToRepo(pb.CurrencyType(99)) != repository.Lamps {
		t.Fatal("default proto")
	}
	if RepoCurrencyToProto(repository.Lamps) != pb.CurrencyType_LAMPS {
		t.Fatal("lamps repo")
	}
	if RepoCurrencyToProto(repository.Tickets) != pb.CurrencyType_TICKETS {
		t.Fatal("tickets repo")
	}
	if RepoCurrencyToProto(repository.CurrencyType("x")) != pb.CurrencyType_LAMPS {
		t.Fatal("default repo")
	}
}

func TestBalanceToProto(t *testing.T) {
	p := BalanceToProto("u1", repository.Tickets, 42, 100)
	if p.UserId != "u1" || p.Balance != 42 || p.Currency != pb.CurrencyType_TICKETS {
		t.Fatalf("%+v", p)
	}
}

func TestDomainBalanceToEntry(t *testing.T) {
	e := DomainBalanceToEntry(model.Balance{Currency: "tickets", Amount: 7, UpdatedAt: 1})
	if e.Balance != 7 || e.Currency != pb.CurrencyType_TICKETS {
		t.Fatalf("%+v", e)
	}
	e2 := DomainBalanceToEntry(model.Balance{Currency: "lamps", Amount: 3, UpdatedAt: 2})
	if e2.Currency != pb.CurrencyType_LAMPS {
		t.Fatalf("%+v", e2)
	}
}

