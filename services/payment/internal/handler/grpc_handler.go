package handler

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/payment/internal/model"
	"github.com/Eastwesser/event-horizon/services/payment/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/payment/proto"
)

type GRPCHandler struct {
	pb.UnimplementedPaymentServiceServer
	svc *service.PaymentService
}

func NewGRPCHandler(svc *service.PaymentService) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

func (h *GRPCHandler) CreateCheckout(ctx context.Context, req *pb.CreateCheckoutRequest) (*pb.CreateCheckoutResponse, error) {
	p, err := h.svc.CreateCheckout(ctx, req.GetUserId(), req.GetPlan())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.CreateCheckoutResponse{
		PaymentId:   p.ID,
		CheckoutUrl: p.CheckoutURL,
		AmountRub:   int32(p.AmountRub),
		Plan:        p.Plan,
		Status:      p.Status,
	}, nil
}

func (h *GRPCHandler) ConfirmPayment(ctx context.Context, req *pb.ConfirmPaymentRequest) (*pb.ConfirmPaymentResponse, error) {
	sub, err := h.svc.ConfirmPayment(ctx, req.GetPaymentId(), req.GetProviderRef(), req.GetWebhookSecret())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.ConfirmPaymentResponse{
		Success:        true,
		Message:        "subscription activated",
		SubscriptionId: sub.ID,
		ExpiresAtUnix:  sub.ExpiresAt.Unix(),
	}, nil
}

func (h *GRPCHandler) GetSubscription(ctx context.Context, req *pb.GetSubscriptionRequest) (*pb.GetSubscriptionResponse, error) {
	sub, err := h.svc.GetSubscription(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &pb.GetSubscriptionResponse{Active: false, Status: "none"}, nil
		}
		return nil, mapErr(err)
	}
	return &pb.GetSubscriptionResponse{
		Active:        sub.IsActive(time.Now().UTC()),
		Plan:          sub.Plan,
		Status:        sub.Status,
		ExpiresAtUnix: sub.ExpiresAt.Unix(),
		AmountRub:     int32(sub.AmountRub),
	}, nil
}

func (h *GRPCHandler) CanPurchaseMerch(ctx context.Context, req *pb.CanPurchaseMerchRequest) (*pb.CanPurchaseMerchResponse, error) {
	ok, reason := h.svc.CanPurchaseMerch(ctx, req.GetUserId())
	return &pb.CanPurchaseMerchResponse{Allowed: ok, Reason: reason}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, model.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, model.ErrInvalidInput), errors.Is(err, model.ErrInvalidStatus):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, model.ErrUnauthorized):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, model.ErrAlreadyPaid):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
