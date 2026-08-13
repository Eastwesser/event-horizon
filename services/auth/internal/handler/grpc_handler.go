package handler

import (
	"context"
	"net/mail"
	"strings"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/auth/internal/converter"
	"github.com/Eastwesser/event-horizon/services/auth/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/auth/proto"
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
	if err := validateCredentials(req.Email, req.Password); err != nil {
		return nil, err
	}
	if role := strings.TrimSpace(req.Role); role != "" && role != "user" && role != "author" {
		return nil, status.Error(codes.InvalidArgument, "role must be empty, user, or author")
	}

	userID, role, err := h.authService.Register(ctx, req.Email, req.Password, req.Role)
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
		Role:    role,
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := validateCredentials(req.Email, req.Password); err != nil {
		return nil, err
	}

	pair, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &pb.LoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    pair.ExpiresIn,
		UserId:       pair.UserID,
		Email:        req.Email,
		Role:         pair.Role,
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token required")
	}
	pair, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &pb.RefreshTokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    pair.ExpiresIn,
		UserId:       pair.UserID,
		Role:         pair.Role,
	}, nil
}

func (h *AuthHandler) Whoami(ctx context.Context, req *pb.WhoamiRequest) (*pb.WhoamiResponse, error) {
	if req.GetAccessToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token required")
	}
	user, err := h.authService.Whoami(ctx, req.AccessToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &pb.WhoamiResponse{
		UserId:   user.ID,
		Email:    user.Email,
		Role:     user.Role,
		Nickname: user.Nickname,
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token required")
	}

	userID, email, role, err := h.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:     true,
		UserId:    userID,
		Email:     email,
		Role:      role,
		ExpiresAt: 0, // можно добавить из claims
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token required")
	}
	if err := h.authService.Logout(ctx, req.Token); err != nil {
		return &pb.LogoutResponse{Success: false}, nil
	}
	return &pb.LogoutResponse{Success: true}, nil
}

func (h *AuthHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}

	user, err := h.authService.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	bestScores, totalScore, err := h.authService.GetUserScores(ctx, req.UserId)
	if err != nil {
		bestScores = make(map[string]int32)
		totalScore = 0
	}

	return converter.UserToProtoWithScores(user, bestScores, totalScore), nil
}

func (h *AuthHandler) UpdateNickname(ctx context.Context, req *pb.UpdateNicknameRequest) (*pb.UpdateNicknameResponse, error) {
	err := h.authService.UpdateNickname(ctx, req.UserId, req.Nickname)
	if err != nil {
		return &pb.UpdateNicknameResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.UpdateNicknameResponse{Success: true, Message: "nickname updated"}, nil
}

// UpdateRole меняет роль пользователя. Проверка "запрос делает admin" — на стороне Gateway.
func (h *AuthHandler) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleResponse, error) {
	if req.UserId == "" || req.Role == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and role required")
	}
	if err := h.authService.UpdateRole(ctx, req.UserId, req.Role); err != nil {
		return &pb.UpdateRoleResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.UpdateRoleResponse{Success: true, Message: "role updated"}, nil
}

// validateCredentials mirrors intended protoc-gen-validate rules for Login/Register
// (email format, password min_len=8, max_len=128) until .pb.validate.go is generated.
func validateCredentials(email, password string) error {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return status.Error(codes.InvalidArgument, "email and password required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return status.Error(codes.InvalidArgument, "invalid email")
	}
	n := utf8.RuneCountInString(password)
	if n < 8 || n > 128 {
		return status.Error(codes.InvalidArgument, "password must be 8-128 characters")
	}
	return nil
}
