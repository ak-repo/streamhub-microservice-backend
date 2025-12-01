package grpc

import (
	"context"

	"github.com/ak-repo/stream-hub/gen/authpb"
	"github.com/ak-repo/stream-hub/internal/auth_service/port"
)

type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	service port.AuthService
}

func NewAuthServer(s port.AuthService) *AuthServer {
	return &AuthServer{service: s}
}

func (s *AuthServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	err := s.service.Register(ctx, req.Email, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &authpb.RegisterResponse{
		Success: true,
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	user, err := s.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	resp := &authpb.AuthUser{
		Id:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		EmailVerified: user.EmailVerified,
		Role:          user.Role,
	}
	return &authpb.LoginResponse{
		User: resp,
	}, nil
}

// FindByEmail finds a user by email
func (s *AuthServer) FindByEmail(ctx context.Context, req *authpb.FindByEmailRequest) (*authpb.FindUserResponse, error) {
	user, err := s.service.FindUser(ctx, req.Email, "email")
	if err != nil {
		return nil, err
	}

	return &authpb.FindUserResponse{User: &authpb.AuthUser{
		Id:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
		IsBanned:      user.IsBanned,
		CreatedAt:     user.CreatedAt.String(),
	}}, nil
}

// FindById finds a user by ID
func (s *AuthServer) FindById(ctx context.Context, req *authpb.FindByIdRequest) (*authpb.FindUserResponse, error) {
	user, err := s.service.FindUser(ctx, req.Id, "id")
	if err != nil {
		return nil, err
	}

	return &authpb.FindUserResponse{User: &authpb.AuthUser{
		Id:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
		IsBanned:      user.IsBanned,
		CreatedAt:     user.CreatedAt.String(),
	}}, nil
}

// Send Magic Link
func (s *AuthServer) SendMagicLink(ctx context.Context, req *authpb.SendMagicLinkRequest) (*authpb.SendMagicLinkResponse, error) {

	magicLink, exp, err := s.service.SendMagicLink(req.Email)
	if err != nil {
		return nil, err
	}

	return &authpb.SendMagicLinkResponse{
		MagicLink: magicLink,
		ExpiresAt: exp,
	}, nil
}

// Verify Magic Link
func (s *AuthServer) VerifyMagicLink(ctx context.Context, req *authpb.VerifyMagicLinkRequest) (*authpb.VerifyMagicLinkResponse, error) {

	err := s.service.VerifyMagicLink(ctx, req.Token, req.Email)
	if err != nil {
		return nil, err
	}

	return &authpb.VerifyMagicLinkResponse{Success: true}, nil
}
