package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/authors/internal/model"
	"github.com/Eastwesser/event-horizon/services/authors/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/authors/proto"
)

type GRPCHandler struct {
	pb.UnimplementedAuthorsServiceServer
	svc *service.AuthorsService
}

func NewGRPCHandler(svc *service.AuthorsService) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

func toPB(a *model.Author) *pb.Author {
	if a == nil {
		return nil
	}
	return &pb.Author{
		Id: a.ID, UserId: a.UserID, DisplayName: a.DisplayName, Bio: a.Bio,
		AvatarUrl: a.AvatarURL, Active: a.Active,
		CreatedAtUnix: a.CreatedAt.Unix(), UpdatedAtUnix: a.UpdatedAt.Unix(),
	}
}

func (h *GRPCHandler) UpsertProfile(ctx context.Context, req *pb.UpsertProfileRequest) (*pb.UpsertProfileResponse, error) {
	a, err := h.svc.UpsertProfile(ctx, req.GetUserId(), req.GetDisplayName(), req.GetBio(), req.GetAvatarUrl())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.UpsertProfileResponse{Author: toPB(a)}, nil
}

func (h *GRPCHandler) GetAuthor(ctx context.Context, req *pb.GetAuthorRequest) (*pb.GetAuthorResponse, error) {
	a, err := h.svc.GetAuthor(ctx, req.GetUserId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.GetAuthorResponse{Author: toPB(a)}, nil
}

func (h *GRPCHandler) ListAuthors(ctx context.Context, req *pb.ListAuthorsRequest) (*pb.ListAuthorsResponse, error) {
	list, total, err := h.svc.ListAuthors(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*pb.Author, 0, len(list))
	for _, a := range list {
		out = append(out, toPB(a))
	}
	return &pb.ListAuthorsResponse{Authors: out, Total: total}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, model.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, model.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
