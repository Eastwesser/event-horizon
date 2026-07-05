package handler

import (
	"context"
	
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	
	pb "github.com/Eastwesser/event-horizon/services/auth/proto"
	"github.com/Eastwesser/event-horizon/services/auth/internal/service"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
    if req.Email == "" || req.Password == "" {
        return nil, status.Error(codes.InvalidArgument, "email and password required")
    }
    
    userID, err := h.authService.Register(ctx, req.Email, req.Password)
    if err != nil {
        return &pb.RegisterResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }
    
    return &pb.RegisterResponse{
        UserId:  userID,
        Email:   req.Email,
        Success: true,
        Message: "user registered successfully",
    }, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password required")
	}
	
	token, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	
	return &pb.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   86400, // 24 часа в секундах
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token required")
	}
	
	userID, email, err := h.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}
	
	return &pb.ValidateTokenResponse{
		Valid:   true,
		UserId:  userID,
		Email:   email,
		ExpiresAt: 0, // можно добавить из claims
	}, nil
}

func (h *AuthHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id required")
    }

    user, err := h.authService.GetUser(ctx, req.UserId)
    if err != nil {
        return nil, status.Error(codes.NotFound, err.Error())
    }

    // Получаем рекорды
    bestScores, totalScore, err := h.authService.GetUserScores(ctx, req.UserId)
    if err != nil {
        // Если ошибка — просто логируем, но не прерываем запрос
        // Можно вернуть пустые рекорды
        bestScores = make(map[string]int32)
        totalScore = 0
    }

    return &pb.GetUserResponse{
        UserId:      user.ID,
        Email:       user.Email,
        Nickname:    user.Nickname,
        BestScores:  bestScores,
        TotalScore:  totalScore,
    }, nil
}

func (h *AuthHandler) UpdateNickname(ctx context.Context, req *pb.UpdateNicknameRequest) (*pb.UpdateNicknameResponse, error) {
    err := h.authService.UpdateNickname(ctx, req.UserId, req.Nickname)
    if err != nil {
        return &pb.UpdateNicknameResponse{Success: false, Message: err.Error()}, nil
    }
    return &pb.UpdateNicknameResponse{Success: true, Message: "nickname updated"}, nil
}
