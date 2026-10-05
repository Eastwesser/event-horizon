package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/notification/internal/model"
	"github.com/Eastwesser/event-horizon/services/notification/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/notification/proto"
)

type GRPCHandler struct {
	pb.UnimplementedNotificationServiceServer
	inbox *service.InboxService
}

func NewGRPCHandler(inbox *service.InboxService) *GRPCHandler {
	return &GRPCHandler{inbox: inbox}
}

func (h *GRPCHandler) ListNotifications(ctx context.Context, req *pb.ListNotificationsRequest) (*pb.ListNotificationsResponse, error) {
	list, total, unread, err := h.inbox.List(ctx, req.GetUserId(), int(req.GetLimit()), int(req.GetOffset()), req.GetUnreadOnly())
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*pb.Notification, 0, len(list))
	for _, n := range list {
		var readUnix int64
		if n.ReadAt != nil {
			readUnix = n.ReadAt.Unix()
		}
		out = append(out, &pb.Notification{
			Id:             n.ID,
			UserId:         n.UserID,
			Title:          n.Title,
			Body:           n.Body,
			Link:           n.Link,
			CreatedAtUnix:  n.CreatedAt.Unix(),
			ReadAtUnix:     readUnix,
		})
	}
	return &pb.ListNotificationsResponse{
		Notifications: out,
		Total:         total,
		UnreadCount:   unread,
	}, nil
}

func (h *GRPCHandler) MarkRead(ctx context.Context, req *pb.MarkReadRequest) (*pb.MarkReadResponse, error) {
	n, err := h.inbox.MarkRead(ctx, req.GetUserId(), req.GetNotificationId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.MarkReadResponse{Marked: n}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, model.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, model.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
