package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/history/internal/model"
	"github.com/Eastwesser/event-horizon/services/history/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/history/proto"
)

type GRPCHandler struct {
	pb.UnimplementedHistoryServiceServer
	svc *service.HistoryService
}

func NewGRPCHandler(svc *service.HistoryService) *GRPCHandler { return &GRPCHandler{svc: svc} }

func (h *GRPCHandler) RecordEvent(ctx context.Context, req *pb.RecordEventRequest) (*pb.RecordEventResponse, error) {
	id, err := h.svc.RecordEvent(ctx, req.GetUserId(), req.GetEventType(), req.GetPayloadJson())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.RecordEventResponse{EventId: id}, nil
}

func (h *GRPCHandler) ListEvents(ctx context.Context, req *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	list, total, err := h.svc.ListEvents(ctx, req.GetUserId(), req.GetEventType(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*pb.HistoryEvent, 0, len(list))
	for _, e := range list {
		out = append(out, &pb.HistoryEvent{
			Id: e.ID, UserId: e.UserID, EventType: e.EventType,
			PayloadJson: e.Payload, CreatedAtUnix: e.CreatedAt.Unix(),
		})
	}
	return &pb.ListEventsResponse{Events: out, Total: total}, nil
}

func mapErr(err error) error {
	if errors.Is(err, model.ErrInvalidInput) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Errorf(codes.Internal, "%v", err)
}
