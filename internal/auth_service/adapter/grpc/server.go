package grpc

import (
	"context"

	"github.com/ak-repo/stream-hub/gen/authpb"
	"github.com/ak-repo/stream-hub/internal/auth_service/port"
	"github.com/ak-repo/stream-hub/pkg/helper"
)

type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	service port.AuthService
}

func NewAuthServer(s port.AuthService) *AuthServer {
	return &AuthServer{service: s}
}

func (s *AuthServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	user, err := s.service.Register(ctx, req.Email, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	resp := &authpb.AuthUser{
		Id:       user.ID,
		Email:    user.Email,
		Username: user.Username,
	}

	return &authpb.RegisterResponse{
		User: resp,
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	user, err := s.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	resp := &authpb.AuthUser{
		Id:       user.ID,
		Email:    user.Email,
		Username: user.Username,
	}
	return &authpb.LoginResponse{
		User: resp,
	}, nil
}

// FindByEmail finds a user by email
func (s *AuthServer) FindByEmail(ctx context.Context, req *authpb.FindByEmailRequest) (*authpb.FindUserResponse, error) {

	return &authpb.FindUserResponse{User: &authpb.AuthUser{}}, nil
}

// FindById finds a user by ID
func (s *AuthServer) FindById(ctx context.Context, req *authpb.FindByIdRequest) (*authpb.FindUserResponse, error) {

	return &authpb.FindUserResponse{User: &authpb.AuthUser{}}, nil
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

	user, err := s.service.VerifyMagicLink(ctx, req.Token, req.Email)
	if err != nil {
		return nil, err
	}

	resp := &authpb.AuthUser{
		Id:            user.ID,
		Email:         user.Email,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
		CreatedAt:     helper.TimeToString(user.CreatedAt),
	}

	return &authpb.VerifyMagicLinkResponse{User: resp}, nil
}
