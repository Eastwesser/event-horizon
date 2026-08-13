package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	paymentPb "github.com/Eastwesser/event-horizon/services/payment/proto"
	"google.golang.org/grpc"
)

type stubPayment struct {
	allowed bool
	reason  string
	err     error
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
	if s.err != nil {
		return nil, s.err
	}
	return &paymentPb.CanPurchaseMerchResponse{Allowed: s.allowed, Reason: s.reason}, nil
}

func TestCheckMerchAllowed_NilClient(t *testing.T) {
	s := &shopService{}
	err := s.checkMerchAllowed(context.Background(), "u1")
	if err == nil || !strings.Contains(err.Error(), "payment service") {
		t.Fatalf("got %v", err)
	}
}

func TestCheckMerchAllowed_Blocked(t *testing.T) {
	s := &shopService{payment: stubPayment{allowed: false, reason: "no subscription"}}
	err := s.checkMerchAllowed(context.Background(), "u1")
	if err == nil || !strings.Contains(err.Error(), "no subscription") {
		t.Fatalf("got %v", err)
	}
}

func TestCheckMerchAllowed_Allowed(t *testing.T) {
	s := &shopService{payment: stubPayment{allowed: true}}
	if err := s.checkMerchAllowed(context.Background(), "u1"); err != nil {
		t.Fatal(err)
	}
}

func TestCheckMerchAllowed_RPCError(t *testing.T) {
	s := &shopService{payment: stubPayment{err: errors.New("down")}}
	err := s.checkMerchAllowed(context.Background(), "u1")
	if err == nil || !strings.Contains(err.Error(), "subscription check failed") {
		t.Fatalf("got %v", err)
	}
}

func TestCheckMerchAllowed_DefaultReason(t *testing.T) {
	s := &shopService{payment: stubPayment{allowed: false, reason: ""}}
	err := s.checkMerchAllowed(context.Background(), "u1")
	if err == nil || !strings.Contains(err.Error(), "active subscription required") {
		t.Fatalf("got %v", err)
	}
}
