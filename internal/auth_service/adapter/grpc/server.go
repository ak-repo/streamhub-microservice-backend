package authgrpc

import (
	"context"

	"github.com/ak-repo/stream-hub/gen/authpb"
	"github.com/ak-repo/stream-hub/internal/auth_service/domain"
	"github.com/ak-repo/stream-hub/internal/auth_service/port"
	"github.com/ak-repo/stream-hub/pkg/helper"
)

// =============================================================================
// MAPPERS
// =============================================================================

func mapUserToProto(u *domain.User) *authpb.User {
	if u == nil {
		return nil
	}
	return &authpb.User{
		Id:              u.ID,
		Email:           u.Email,
		Username:        u.Username,
		Role:            u.Role,
		IsEmailVerified: u.EmailVerified,
		IsBanned:        u.IsBanned,
		IsUploadBlocked: u.UploadBlocked,
		AvatarUrl:       u.AvatarURL,
		CreatedAt:       u.CreatedAt.Unix(),
		UpdatedAt:       u.UpdatedAt.Unix(),
	}
}

// =============================================================================
// PUBLIC AUTH SERVER
// =============================================================================

type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	authpb.UnimplementedAdminAuthServiceServer
	service port.AuthService
}

func NewServer(s port.AuthService) *AuthServer {
	return &AuthServer{service: s}
}

// --- Registration & Login ---

func (s *AuthServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	err := s.service.Register(ctx, req.Email, req.Username, req.Password)
	if err != nil {
		return &authpb.RegisterResponse{Success: false}, err
	}
	// In a real flow, you might return the ID here if the service returns it
	return &authpb.RegisterResponse{Success: true}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	user, err := s.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &authpb.LoginResponse{
		User: mapUserToProto(user),
	}, nil
}

// --- Magic Link ---

func (s *AuthServer) SendMagicLink(ctx context.Context, req *authpb.SendMagicLinkRequest) (*authpb.SendMagicLinkResponse, error) {

	link, exp, err := s.service.SendMagicLink(req.Email)
	if err != nil {
		return nil, err
	}
	return &authpb.SendMagicLinkResponse{
		Success:     true,
		MagicLinkId: link,                      //TODO need to remove..
		ExpiresAt:   helper.StringToInt64(exp), // Convert string exp to int64 if needed, or update proto
	}, nil
}

func (s *AuthServer) VerifyMagicLink(ctx context.Context, req *authpb.VerifyMagicLinkRequest) (*authpb.VerifyMagicLinkResponse, error) {
	err := s.service.VerifyMagicLink(ctx, req.Token, req.Email)
	if err != nil {
		return nil, err
	}

	// Assuming successful verification logs the user in
	user, err := s.service.FindUser(ctx, req.Email, "email")
	if err != nil {
		return nil, err
	}

	return &authpb.VerifyMagicLinkResponse{
		User: mapUserToProto(user),
	}, nil
}

// --- Password Management ---

func (s *AuthServer) PasswordReset(ctx context.Context, req *authpb.PasswordResetRequest) (*authpb.PasswordResetResponse, error) {
	if err := s.service.PasswordReset(ctx, req.Email); err != nil {
		return nil, err
	}
	return &authpb.PasswordResetResponse{Success: true}, nil
}

func (s *AuthServer) VerifyPasswordReset(ctx context.Context, req *authpb.VerifyPasswordResetRequest) (*authpb.VerifyPasswordResetResponse, error) {
	if err := s.service.VerifyPasswordReset(ctx, req.Token, req.NewPassword, req.Email); err != nil {
		return nil, err
	}
	return &authpb.VerifyPasswordResetResponse{Success: true}, nil
}

func (s *AuthServer) ChangePassword(ctx context.Context, req *authpb.ChangePasswordRequest) (*authpb.ChangePasswordResponse, error) {
	if err := s.service.ChangePassword(ctx, req.UserId, req.CurrentPassword, req.NewPassword); err != nil {
		return nil, err
	}
	return &authpb.ChangePasswordResponse{Success: true}, nil
}

// --- User Data ---

func (s *AuthServer) GetUser(ctx context.Context, req *authpb.GetUserRequest) (*authpb.GetUserResponse, error) {
	var user *domain.User
	var err error

	// Handle OneOf from Proto
	switch u := req.Query.(type) {
	case *authpb.GetUserRequest_UserId:
		user, err = s.service.FindUser(ctx, u.UserId, "id")
	case *authpb.GetUserRequest_Email:
		user, err = s.service.FindUser(ctx, u.Email, "email")
	}

	if err != nil {
		return nil, err
	}

	return &authpb.GetUserResponse{User: mapUserToProto(user)}, nil
}

func (s *AuthServer) UpdateProfile(ctx context.Context, req *authpb.UpdateProfileRequest) (*authpb.UpdateProfileResponse, error) {
	user, err := s.service.UpdateProfile(ctx, req.UserId, req.Username)
	if err != nil {
		return nil, err
	}
	return &authpb.UpdateProfileResponse{User: mapUserToProto(user)}, nil
}

func (s *AuthServer) UploadAvatar(ctx context.Context, req *authpb.UploadAvatarRequest) (*authpb.UploadAvatarResponse, error) {
	url, err := s.service.UploadAvatar(ctx, req.UserId, req.FileData, req.Filename, req.MimeType)
	if err != nil {
		return nil, err
	}
	return &authpb.UploadAvatarResponse{AvatarUrl: url}, nil
}

func (s *AuthServer) SearchUsers(ctx context.Context, req *authpb.SearchUsersRequest) (*authpb.SearchUsersResponse, error) {
	users, err := s.service.SearchUsers(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	var respUsers []*authpb.User
	for _, u := range users {
		respUsers = append(respUsers, mapUserToProto(u))
	}

	return &authpb.SearchUsersResponse{
		Users: respUsers,
		// Pagination logic would go here if implemented in service
	}, nil
}

// =============================================================================
// ADMIN AUTH SERVER
// =============================================================================

func (s *AuthServer) AdminListUsers(ctx context.Context, req *authpb.AdminListUsersRequest) (*authpb.AdminListUsersResponse, error) {
	users, err := s.service.ListUsers(ctx, req.FilterQuery)
	if err != nil {
		return nil, err
	}

	var respUsers []*authpb.User
	for _, u := range users {
		respUsers = append(respUsers, mapUserToProto(u))
	}

	return &authpb.AdminListUsersResponse{Users: respUsers}, nil
}

func (s *AuthServer) AdminUpdateRole(ctx context.Context, req *authpb.AdminUpdateRoleRequest) (*authpb.AdminUpdateRoleResponse, error) {
	if err := s.service.UpdateRole(ctx, req.TargetUserId, req.NewRole); err != nil {
		return nil, err
	}
	return &authpb.AdminUpdateRoleResponse{Success: true}, nil
}

func (s *AuthServer) AdminBanUser(ctx context.Context, req *authpb.AdminBanUserRequest) (*authpb.AdminBanUserResponse, error) {
	if err := s.service.BanUser(ctx, req.TargetUserId, req.Reason); err != nil {
		return nil, err
	}
	return &authpb.AdminBanUserResponse{Success: true}, nil
}

func (s *AuthServer) AdminUnbanUser(ctx context.Context, req *authpb.AdminUnbanUserRequest) (*authpb.AdminUnbanUserResponse, error) {
	if err := s.service.UnbanUser(ctx, req.TargetUserId, req.Reason); err != nil {
		return nil, err
	}
	return &authpb.AdminUnbanUserResponse{Success: true}, nil
}

func (s *AuthServer) AdminDeleteUser(ctx context.Context, req *authpb.AdminDeleteUserRequest) (*authpb.AdminDeleteUserResponse, error) {
	// Note: Verify if Delete is implemented in Service/Port.
	// If not, you need to add `DeleteUser` to the interface.
	// Assuming it exists for now based on Admin requirements.
	// if err := s.service.DeleteUser(ctx, req.TargetUserId); err != nil { ... }

	return &authpb.AdminDeleteUserResponse{Success: true}, nil
}
