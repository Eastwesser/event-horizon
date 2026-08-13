package converter

import (
	"github.com/Eastwesser/event-horizon/services/billing/internal/model"
	"github.com/Eastwesser/event-horizon/services/billing/internal/repository"
	pb "github.com/Eastwesser/event-horizon/services/billing/proto"
)

func ProtoCurrencyToRepo(c pb.CurrencyType) repository.CurrencyType {
	switch c {
	case pb.CurrencyType_LAMPS:
		return repository.Lamps
	case pb.CurrencyType_TICKETS:
		return repository.Tickets
	default:
		return repository.Lamps
	}
}

func RepoCurrencyToProto(c repository.CurrencyType) pb.CurrencyType {
	switch c {
	case repository.Lamps:
		return pb.CurrencyType_LAMPS
	case repository.Tickets:
		return pb.CurrencyType_TICKETS
	default:
		return pb.CurrencyType_LAMPS
	}
}

func BalanceToProto(userID string, currency repository.CurrencyType, amount int32, updatedAt int64) *pb.GetBalanceResponse {
	return &pb.GetBalanceResponse{
		UserId:    userID,
		Currency:  RepoCurrencyToProto(currency),
		Balance:   amount,
		UpdatedAt: updatedAt,
	}
}

func DomainBalanceToEntry(b model.Balance) *pb.BalanceEntry {
	var c repository.CurrencyType
	switch b.Currency {
	case string(repository.Tickets):
		c = repository.Tickets
	default:
		c = repository.Lamps
	}
	return &pb.BalanceEntry{
		Currency:  RepoCurrencyToProto(c),
		Balance:   b.Amount,
		UpdatedAt: b.UpdatedAt,
	}
}
