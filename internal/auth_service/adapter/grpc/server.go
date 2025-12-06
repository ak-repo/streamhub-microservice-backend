package grpc

import (
	"context"

	"github.com/ak-repo/stream-hub/gen/authpb"
	"github.com/ak-repo/stream-hub/internal/auth_service/domain"
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

func mapUser(u *domain.User) *authpb.AuthUser {
	return &authpb.AuthUser{
		Id:            u.ID,
		Email:         u.Email,
		Username:      u.Username,
		Role:          u.Role,
		EmailVerified: u.EmailVerified,
		IsBanned:      u.IsBanned,
		UploadBlocked: u.UploadBlocked,
		AvatarUrl:     u.Avatar_url,
		CreatedAt:     helper.TimeToString(u.CreatedAt),
	}
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

	resp := mapUser(user)
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

	return &authpb.FindUserResponse{User: mapUser(user)}, nil
}

// FindById finds a user by ID
func (s *AuthServer) FindById(ctx context.Context, req *authpb.FindByIdRequest) (*authpb.FindUserResponse, error) {
	user, err := s.service.FindUser(ctx, req.Id, "id")
	if err != nil {
		return nil, err
	}

	return &authpb.FindUserResponse{User: mapUser(user)}, nil
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

func (s *AuthServer) PasswordReset(ctx context.Context, req *authpb.PasswordResetRequest) (*authpb.PasswordResetResponse, error) {
	if err := s.service.PasswordReset(ctx, req.Email); err != nil {
		return nil, err
	}

	return &authpb.PasswordResetResponse{Success: true}, nil
}

func (s *AuthServer) VerifyPasswordReset(ctx context.Context, req *authpb.PasswordResetVerifyRequest) (*authpb.PasswordResetVerifyResponse, error) {
	if err := s.service.VerifyPasswordReset(ctx, req.Token, req.NewPassword, req.Email); err != nil {
		return nil, err
	}

	return &authpb.PasswordResetVerifyResponse{Success: true}, nil
}

func (s *AuthServer) UpdateProfile(ctx context.Context, req *authpb.UpdateProfileRequest) (*authpb.UpdateProfileResponse, error) {

	user, err := s.service.UpdateProfile(ctx, req.Id, req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	resp := mapUser(user)

	return &authpb.UpdateProfileResponse{User: resp}, nil
}

func (s *AuthServer) ChangePassword(ctx context.Context, req *authpb.ChangePasswordRequest) (*authpb.ChangePasswordResponse, error) {

	if err := s.service.ChangePassword(ctx, req.Id, req.Password, req.NewPassword); err != nil {
		return nil, err
	}

	return &authpb.ChangePasswordResponse{Success: true}, nil

}

func (s *AuthServer) SearchUsers(ctx context.Context, req *authpb.SearchUsersRequest) (*authpb.SearchUsersResponse, error) {

	users, err := s.service.FindAllUsers(ctx, req.Query)

	if err != nil {
		return nil, err
	}

	var resp []*authpb.AuthUser
	for _, u := range users {
		resp = append(resp, mapUser(u))
	}

	return &authpb.SearchUsersResponse{Users: resp, Total: int32(len(resp))}, nil
}

func (s *AuthServer) UploadAvatar(ctx context.Context, req *authpb.UploadAvatarRequest) (*authpb.UploadAvatarResponse, error) {
	url, err := s.service.UploadAvatar(ctx, req.UserId, req.File, req.Filename, req.ContentType)
	if err != nil {
		return nil, err
	}

	return &authpb.UploadAvatarResponse{Url: url}, nil
}
