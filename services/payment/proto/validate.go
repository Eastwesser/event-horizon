package payment

import "fmt"

func (r *CreateCheckoutRequest) Validate() error {
	if r.GetUserId() == "" {
		return fmt.Errorf("user_id is required")
	}
	switch r.GetPlan() {
	case "present", "future":
	default:
		return fmt.Errorf("plan must be present or future")
	}
	return nil
}

func (r *ConfirmPaymentRequest) Validate() error {
	if r.GetPaymentId() == "" {
		return fmt.Errorf("payment_id is required")
	}
	return nil
}

func (r *GetSubscriptionRequest) Validate() error {
	if r.GetUserId() == "" {
		return fmt.Errorf("user_id is required")
	}
	return nil
}

func (r *CanPurchaseMerchRequest) Validate() error {
	if r.GetUserId() == "" {
		return fmt.Errorf("user_id is required")
	}
	return nil
}
