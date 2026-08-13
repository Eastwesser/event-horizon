package service

import (
	"context"
	"errors"
	"testing"

	paymentPb "github.com/Eastwesser/event-horizon/services/payment/proto"
	"google.golang.org/grpc"
)

type stubPayment struct {
	allowed bool
	reason  string
}

func (s stubPayment) CreateCheckout(context.Context, *paymentPb.CreateCheckoutRequest, ...grpc.CallOption) (*paymentPb.CreateCheckoutResponse, error) {
	return nil, errors.New("unused")
}
func (s stubPayment) ConfirmPayment(context.Context, *paymentPb.ConfirmPaymentRequest, ...grpc.CallOption) (*paymentPb.ConfirmPaymentResponse, error) {
	return nil, errors.New("unused")
}
func (s stubPayment) GetSubscription(context.Context, *paymentPb.GetSubscriptionRequest, ...grpc.CallOption) (*paymentPb.GetSubscriptionResponse, error) {
	return nil, errors.New("unused")
}
func (s stubPayment) CanPurchaseMerch(context.Context, *paymentPb.CanPurchaseMerchRequest, ...grpc.CallOption) (*paymentPb.CanPurchaseMerchResponse, error) {
	return &paymentPb.CanPurchaseMerchResponse{Allowed: s.allowed, Reason: s.reason}, nil
}

func TestPurchaseItem_MerchBlockedWithoutSubscription(t *testing.T) {
	svc := &shopService{payment: stubPayment{allowed: false, reason: "no subscription"}}
	gate, err := svc.payment.CanPurchaseMerch(context.Background(), &paymentPb.CanPurchaseMerchRequest{UserId: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	if gate.GetAllowed() {
		t.Fatal("expected blocked")
	}
	if gate.GetReason() != "no subscription" {
		t.Fatalf("reason=%q", gate.GetReason())
	}
}

func TestPurchaseItem_MerchAllowedWithSubscription(t *testing.T) {
	svc := &shopService{payment: stubPayment{allowed: true}}
	gate, err := svc.payment.CanPurchaseMerch(context.Background(), &paymentPb.CanPurchaseMerchRequest{UserId: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	if !gate.GetAllowed() {
		t.Fatal("expected allowed")
	}
}
