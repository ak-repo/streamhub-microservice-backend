package admingrpc

import (
	"context"
	"log"

	"github.com/ak-repo/stream-hub/gen/adminpb"
	"github.com/ak-repo/stream-hub/internal/admin_service/domain"
	"github.com/ak-repo/stream-hub/internal/admin_service/port"
	"github.com/ak-repo/stream-hub/pkg/helper"
)

type AdminServer struct {
	adminpb.UnimplementedAdminServiceServer
	service port.AdminService
}

func NewAdminServer(service port.AdminService) *AdminServer {
	return &AdminServer{service: service}
}

//
// ---------------------------------------------------------
// Helpers: Model → Protobuf Converters
// ---------------------------------------------------------
//

func mapUser(u *domain.User) *adminpb.User {
	return &adminpb.User{
		Id:            u.ID,
		Email:         u.Email,
		Username:      u.Username,
		Role:          u.Role,
		EmailVerified: u.EmailVerified,
		IsBanned:      u.IsBanned,
		CreatedAt:     helper.TimeToString(u.CreatedAt),
	}
}

func mapMember(m *domain.ChannelMember) *adminpb.MemberInfo {
	return &adminpb.MemberInfo{
		Id:       m.ID,
		UserId:   m.UserID,
		Username: m.Username,
		JoinedAt: helper.TimeToString(m.JoinedAt),
	}
}

func mapMembers(members []*domain.ChannelMember) []*adminpb.MemberInfo {
	resp := make([]*adminpb.MemberInfo, 0, len(members))
	for _, m := range members {
		resp = append(resp, mapMember(m))
	}
	return resp
}

func mapChannel(c *domain.Channel, members []*domain.ChannelMember) *adminpb.Channel {
	return &adminpb.Channel{
		Id:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatedBy:   c.CreatedBy,
		IsFrozen:    c.IsFrozen,
		OwnerName:   c.OwnerName,
		CreatedAt:   helper.TimeToString(c.CreatedAt),
		Members:     mapMembers(members),
	}
}

func mapFile(f *domain.File) *adminpb.File {
	return &adminpb.File{
		Id:          f.ID,
		OwnerId:     f.OwnerID,
		ChannelId:   f.ChannelID,
		Filename:    f.Filename,
		Size:        f.Size,
		MimeType:    f.MimeType,
		StoragePath: f.StoragePath,
		IsPublic:    f.IsPublic,
		CreatedAt:   helper.TimeToString(f.CreatedAt),
		OwnerName:   f.OwnerName,
		ChannelName: f.ChannelName,
	}
}

//
// ---------------------------------------------------------
// USER MANAGEMENT
// ---------------------------------------------------------
//

func (s *AdminServer) ListUsers(ctx context.Context, req *adminpb.ListUsersRequest) (*adminpb.ListUsersResponse, error) {
	users, err := s.service.ListUsers(ctx, req.FilterBy)
	if err != nil {
		return nil, err
	}

	resp := make([]*adminpb.User, 0, len(users))
	for _, u := range users {
		resp = append(resp, mapUser(u))
	}

	return &adminpb.ListUsersResponse{Users: resp}, nil
}

func (s *AdminServer) BanUser(ctx context.Context, req *adminpb.BanUserRequest) (*adminpb.BanUserResponse, error) {
	if err := s.service.BanUser(ctx, req.UserId, req.Reason); err != nil {
		return nil, err
	}
	return &adminpb.BanUserResponse{Success: true}, nil
}

func (s *AdminServer) UnbanUser(ctx context.Context, req *adminpb.UnbanUserRequest) (*adminpb.UnbanUserResponse, error) {
	if err := s.service.UnbanUser(ctx, req.UserId, req.Reason); err != nil {
		return nil, err
	}
	return &adminpb.UnbanUserResponse{Success: true}, nil
}

func (s *AdminServer) UpdateRole(ctx context.Context, req *adminpb.UpdateRoleRequest) (*adminpb.UpdateRoleResponse, error) {
	if err := s.service.UpdateRole(ctx, req.UserId, req.Role); err != nil {
		return nil, err
	}
	return &adminpb.UpdateRoleResponse{Success: true}, nil
}

//
// ---------------------------------------------------------
// CHANNEL MANAGEMENT
// ---------------------------------------------------------
//

func (s *AdminServer) ListChannels(ctx context.Context, req *adminpb.ListChannelsRequest) (*adminpb.ListChannelsResponse, error) {
	channels, err := s.service.ListChannels(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]*adminpb.Channel, 0, len(channels))

	for _, ch := range channels {
		resp = append(resp, mapChannel(ch.Channel, ch.Members))
	}
	return &adminpb.ListChannelsResponse{Channels: resp}, nil
}

func (s *AdminServer) FreezeChannel(ctx context.Context, req *adminpb.FreezeChannelRequest) (*adminpb.FreezeChannelResponse, error) {

	log.Println("id: ", req.ChannelId)
	if err := s.service.FreezeChannel(ctx, req.ChannelId, req.Reason); err != nil {
		return nil, err
	}
	return &adminpb.FreezeChannelResponse{Success: true}, nil
}

func (s *AdminServer) UnfreezeChannel(ctx context.Context, req *adminpb.UnfreezeChannelRequest) (*adminpb.UnfreezeChannelResponse, error) {
	if err := s.service.UnfreezeChannel(ctx, req.ChannelId); err != nil {
		return nil, err
	}
	return &adminpb.UnfreezeChannelResponse{Success: true}, nil
}

func (s *AdminServer) DeleteChannel(ctx context.Context, req *adminpb.DeleteChannelRequest) (*adminpb.DeleteChannelResponse, error) {
	if err := s.service.DeleteChannel(ctx, req.AdminId, req.ChannelId); err != nil {
		return nil, err
	}
	return &adminpb.DeleteChannelResponse{Success: true}, nil
}

//
// ---------------------------------------------------------
// FILE MANAGEMENT
// ---------------------------------------------------------
//

func (s *AdminServer) AdminListAllFiles(ctx context.Context, req *adminpb.AdminListAllFilesRequest) (*adminpb.AdminListAllFilesResponse, error) {
	files, err := s.service.ListAllFiles(ctx, req.AdminId)
	if err != nil {
		return nil, err
	}

	resp := make([]*adminpb.File, 0, len(files))
	for _, f := range files {
		resp = append(resp, mapFile(f))
	}

	return &adminpb.AdminListAllFilesResponse{Files: resp}, nil
}

func (s *AdminServer) AdminDeleteFile(ctx context.Context, req *adminpb.AdminDeleteFileRequest) (*adminpb.AdminDeleteFileResponse, error) {
	if err := s.service.DeleteFile(ctx, req.AdminId, req.FileId); err != nil {
		return nil, err
	}
	return &adminpb.AdminDeleteFileResponse{Success: true}, nil
}

func (s *AdminServer) AdminBlockUserUpload(ctx context.Context, req *adminpb.AdminBlockUserUploadRequest) (*adminpb.AdminBlockUserUploadResponse, error) {
	if err := s.service.BlockUserUpload(ctx, req.AdminId, req.UserId, req.Block); err != nil {
		return nil, err
	}
	return &adminpb.AdminBlockUserUploadResponse{Success: true}, nil
}

func (s *AdminServer) IsAdmin(ctx context.Context, req *adminpb.IsAdminRequest) (*adminpb.IsAdminResponse, error) {

	return &adminpb.IsAdminResponse{Success: true}, nil

}
