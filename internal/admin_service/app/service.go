package app

import (
	"context"

	"github.com/ak-repo/stream-hub/gen/channelpb"
	"github.com/ak-repo/stream-hub/internal/admin_service/domain"
	"github.com/ak-repo/stream-hub/internal/admin_service/port"
	"github.com/ak-repo/stream-hub/internal/gateway/clients"
	"github.com/ak-repo/stream-hub/pkg/errors"
)

type adminService struct {
	repo    port.AdminRepository
	clients *clients.Clients
}

// Corrected constructor
func NewAdminService(repo port.AdminRepository, clients *clients.Clients) port.AdminService {
	return &adminService{repo: repo, clients: clients}
}

// =========================
// -------- User Governance --------
func (s *adminService) ListUsers(ctx context.Context, filter string) ([]*domain.User, error) {
	var users []*domain.User
	var err error

	switch filter {
	case "active":
		users, err = s.repo.ListActiveUsers(ctx)
	case "banned":
		users, err = s.repo.ListBannedUsers(ctx)
	default:
		users, err = s.repo.ListUsers(ctx)
	}

	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "no users found", err)
	}
	return users, nil
}

func (s *adminService) BanUser(ctx context.Context, userID, reason string) error {
	if err := s.repo.BanUser(ctx, userID, reason); err != nil {
		return errors.New(errors.CodeInternal, "failed to ban user", err)
	}
	return nil
}

func (s *adminService) UnbanUser(ctx context.Context, userID, reason string) error {
	if err := s.repo.UnbanUser(ctx, userID, reason); err != nil {
		return errors.New(errors.CodeInternal, "failed to unban user", err)
	}
	return nil
}

func (s *adminService) UpdateRole(ctx context.Context, userID, role string) error {
	if err := s.repo.UpdateRole(ctx, userID, role); err != nil {
		return errors.New(errors.CodeInternal, "failed to update user role", err)
	}
	return nil
}

// -------- Channel Management --------
func (s *adminService) ListChannels(ctx context.Context) ([]*domain.ChannelWithMembers, error) {
	channels, err := s.repo.ListChannels(ctx)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "no channels found", err)
	}
	return channels, nil
}

func (s *adminService) GetChannel(ctx context.Context, channelID string) (*domain.Channel, error) {
	chann, err := s.clients.Channel.GetChannel(ctx, &channelpb.GetChannelRequest{ChannelId: channelID})
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "channel not found", err)
	}
	return &domain.Channel{
		ID:   chann.Channel.ChannelId,
		Name: chann.Channel.Name,
	}, nil

}

func (s *adminService) FreezeChannel(ctx context.Context, channelID, reason string) error {
	if err := s.repo.FreezeChannel(ctx, channelID, reason); err != nil {
		return errors.New(errors.CodeInternal, "failed to freeze channel", err)
	}
	return nil
}

func (s *adminService) UnfreezeChannel(ctx context.Context, channelID string) error {
	if err := s.repo.UnfreezeChannel(ctx, channelID); err != nil {
		return errors.New(errors.CodeInternal, "failed to unfreeze channel", err)
	}
	return nil
}

func (s *adminService) DeleteChannel(ctx context.Context, channelID string) error {
	if err := s.repo.DeleteChannel(ctx, channelID); err != nil {
		return errors.New(errors.CodeInternal, "failed to delete channel", err)
	}
	return nil
}

// -------- File Management --------
func (s *adminService) ListAllFiles(ctx context.Context, adminID string) ([]*domain.File, error) {
	files, err := s.repo.ListAllFiles(ctx)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "no files found", err)
	}
	return files, nil
}

func (s *adminService) DeleteFile(ctx context.Context, adminID, fileID string) error {
	if err := s.repo.DeleteFile(ctx, fileID); err != nil {
		return errors.New(errors.CodeInternal, "failed to delete file", err)
	}
	return nil
}

func (s *adminService) BlockUserUpload(ctx context.Context, adminID, userID string, block bool) error {
	if err := s.repo.SetUserUploadBlocked(ctx, userID, block); err != nil {
		return errors.New(errors.CodeInternal, "failed to update user upload status", err)
	}
	return nil
}
