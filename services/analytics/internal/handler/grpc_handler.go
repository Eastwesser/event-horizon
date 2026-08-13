package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/analytics/internal/model"
	"github.com/Eastwesser/event-horizon/services/analytics/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/analytics/proto"
)

type GRPCHandler struct {
	pb.UnimplementedAnalyticsServiceServer
	svc *service.AnalyticsService
}

func NewGRPCHandler(svc *service.AnalyticsService) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

func (h *GRPCHandler) RecordEvent(ctx context.Context, req *pb.RecordEventRequest) (*pb.RecordEventResponse, error) {
	if err := h.svc.RecordEvent(ctx, req.GetUserId(), req.GetEventType(), req.GetPayloadJson()); err != nil {
		return nil, mapErr(err)
	}
	return &pb.RecordEventResponse{Ok: true}, nil
}

func (h *GRPCHandler) GetDAU(ctx context.Context, req *pb.GetDAURequest) (*pb.GetDAUResponse, error) {
	days, err := h.svc.GetDAU(ctx, int(req.GetDays()))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*pb.DayCount, 0, len(days))
	for _, d := range days {
		out = append(out, &pb.DayCount{Day: d.Day, Count: d.Count})
	}
	return &pb.GetDAUResponse{Days: out}, nil
}

func (h *GRPCHandler) GetMAU(ctx context.Context, req *pb.GetMAURequest) (*pb.GetMAUResponse, error) {
	mau, window, err := h.svc.GetMAU(ctx, int(req.GetDays()))
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.GetMAUResponse{Mau: mau, WindowDays: int32(window)}, nil
}

func (h *GRPCHandler) GetRetention(ctx context.Context, req *pb.GetRetentionRequest) (*pb.GetRetentionResponse, error) {
	ret, err := h.svc.GetRetention(ctx, int(req.GetCohortDaysAgo()), int(req.GetWindowDays()))
	if err != nil {
		return nil, mapErr(err)
	}
	points := make([]*pb.RetentionPoint, 0, len(ret.Points))
	for _, p := range ret.Points {
		points = append(points, &pb.RetentionPoint{DayN: p.DayN, Rate: p.Rate})
	}
	return &pb.GetRetentionResponse{
		CohortDay:  ret.CohortDay,
		CohortSize: ret.CohortSize,
		Points:     points,
	}, nil
}

func mapErr(err error) error {
	if errors.Is(err, model.ErrInvalidInput) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Errorf(codes.Internal, "%v", err)
}
