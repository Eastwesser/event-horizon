package model

import "errors"

var (
	ErrItemNotFound           = errors.New("item not found")
	ErrInsufficientFunds      = errors.New("insufficient funds")
	ErrAlreadyOwned           = errors.New("item already owned")
	ErrItemUnavailable        = errors.New("item unavailable")
	ErrSubscriptionRequired   = errors.New("subscription_required")
	ErrPurchaseNotFound       = errors.New("purchase not found")
	ErrRefundWindowExpired    = errors.New("Срок возврата истёк (7 дней с покупки).")
	ErrAlreadyFulfilled       = errors.New("Товар уже отправлен — возврат недоступен.")
)
