package billing

import (
	"fmt"
	"strings"
)

func (m *GetBalanceRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	if m.Currency == CurrencyType_CURRENCY_UNSPECIFIED {
		return fmt.Errorf("currency is required")
	}
	return nil
}

func (m *GetAllBalancesRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	return nil
}

func (m *AddCurrencyRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	if m.Currency == CurrencyType_CURRENCY_UNSPECIFIED {
		return fmt.Errorf("currency is required")
	}
	if m.Amount <= 0 {
		return fmt.Errorf("amount must be > 0")
	}
	return nil
}

func (m *SpendCurrencyRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	if m.Currency == CurrencyType_CURRENCY_UNSPECIFIED {
		return fmt.Errorf("currency is required")
	}
	if m.Amount <= 0 {
		return fmt.Errorf("amount must be > 0")
	}
	return nil
}

func (m *GetTransactionHistoryRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	if m.Limit < 0 || m.Limit > 100 {
		return fmt.Errorf("limit must be 0-100")
	}
	if m.Offset < 0 {
		return fmt.Errorf("offset must be >= 0")
	}
	return nil
}
