package app

import (
	"context"
	"log"
	"time"

	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/gen/authpb"
	"github.com/ak-repo/stream-hub/internal/channel_service/domain"
	"github.com/ak-repo/stream-hub/internal/channel_service/port"
	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/ak-repo/stream-hub/pkg/grpc/clients"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"github.com/google/uuid"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
)

type channelService struct {
	repo    port.ChannelRepository
	pubsub  port.PubSub
	clients *clients.Clients
	cfg     *config.Config
}

func NewChannelService(
	repo port.ChannelRepository,
	pubsub port.PubSub,
	clients *clients.Clients,
	cfg *config.Config,
) port.ChannelService {
	return &channelService{
		cfg:     cfg,
		repo:    repo,
		pubsub:  pubsub,
		clients: clients,
	}
}

// =============================================================================
// MESSAGE HANDLING
// =============================================================================

func (s *channelService) PostMessage(ctx context.Context, senderID, channelID, content string) (*domain.Message, error) {
	// 1. Authorization: Is user a member?
	isMember, err := s.repo.IsUserMember(ctx, channelID, senderID)
	if err != nil {
		return nil, errors.New(errors.CodeInternal, "membership check failed", err)
	}
	if !isMember {
		return nil, errors.New(errors.CodeForbidden, "user is not a member of this channel", nil)
	}

	// 2. Authorization: Is channel frozen?
	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil || ch == nil {
		return nil, errors.New(errors.CodeNotFound, "channel not found", err)
	}
	if ch.IsFrozen {
		return nil, errors.New(errors.CodeForbidden, "channel is frozen", nil)
	}

	// 3. Create Message Object (attachment removed)
	msg := &domain.Message{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		SenderID:  senderID,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}

	// 4. Persist
	if err := s.repo.SaveMessage(ctx, msg); err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to save message", err)
	}

	// 5. Broadcast (Real-time)
	if err := s.pubsub.Publish(ctx, channelID, msg); err != nil {
		logger.Log.Error("redis publish failed", zap.String("channel_id", channelID), zap.Error(err))
	}

	return msg, nil
}

func (s *channelService) GetHistory(ctx context.Context, channelID string, limit, offset int) ([]*domain.Message, error) {
	messages, err := s.repo.ListHistory(ctx, channelID, limit, offset)
	if err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to fetch history", err)
	}
	return messages, nil
}

func (s *channelService) SubscribeToChannel(ctx context.Context, channelID string) (<-chan *domain.Message, error) {
	return s.pubsub.Subscribe(ctx, channelID)
}

// =============================================================================
// CHANNEL MANAGEMENT
// =============================================================================

func (s *channelService) CreateChannel(ctx context.Context, name, description, visibility, creatorID string) (*domain.Channel, error) {
	ch := &domain.Channel{
		ID:           uuid.New().String(),
		Name:         name,
		Description:  description,
		Visibility:   visibility,
		CreatedBy:    creatorID,
		CreatedAt:    time.Now().UTC(),
		ActivePlanID: "6c378084-4d59-405a-8eab-1c3de20fe0f5",
		IsFrozen:     false,
	}

	// 1. Create Channel
	if err := s.repo.CreateChannel(ctx, ch); err != nil {
		return nil, errors.New(errors.CodeInternal, "database error creating channel", err)
	}

	// 2. Add Creator as Admin/Owner
	member := &domain.ChannelMember{
		ID:        uuid.New().String(),
		ChannelID: ch.ID,
		UserID:    creatorID,
		Role:      "admin",
		JoinedAt:  time.Now().UTC(),
	}
	if err := s.repo.AddMember(ctx, member); err != nil {
		logger.Log.Error("failed to add creator as member", zap.Error(err))
		return nil, errors.New(errors.CodeInternal, "database error creating channel", err)
	}

	return ch, nil
}

func (s *channelService) GetChannel(ctx context.Context, channelID string) (*domain.Channel, error) {
	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, errors.New(errors.CodeInternal, "database error", err)
	}
	if ch == nil {
		return nil, errors.New(errors.CodeNotFound, "channel not found", nil)
	}
	return ch, nil
}

func (s *channelService) ListUserChannels(ctx context.Context, userID string) ([]*domain.Channel, error) {
	_, err := s.clients.Auth.GetUser(ctx, &authpb.GetUserRequest{Query: &authpb.GetUserRequest_UserId{UserId: userID}})
	if err != nil {
		return nil, errors.New(errors.CodeUnauthorized, "invalid user", err)
	}

	channels, err := s.repo.ListUserChannels(ctx, userID)
	if err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to list channels", err)
	}
	return channels, nil
}

func (s *channelService) DeleteChannel(ctx context.Context, channelID, requesterID string) error {
	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return errors.New(errors.CodeNotFound, "channel not found", err)
	}

	if ch.CreatedBy != requesterID {
		return errors.New(errors.CodeForbidden, "only owner can delete channel", nil)
	}

	if err := s.repo.DeleteChannel(ctx, channelID); err != nil {
		return errors.New(errors.CodeInternal, "delete failed", err)
	}
	return nil
}

func (s *channelService) SearchChannels(ctx context.Context, filer string, limit, offset int32) ([]*domain.Channel, error) {
	if limit >= 0 {
		limit = 10
	}

	channels, err := s.repo.SearchChannels(ctx, filer, limit, offset)
	if err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to search channels", err)
	}

	log.Println("channels length:", len(channels))
	return channels, nil
}

// =============================================================================
// MEMBER MANAGEMENT
// =============================================================================

func (s *channelService) AddMember(ctx context.Context, channelID, userID string) (*domain.ChannelMember, error) {
	if _, err := s.GetChannel(ctx, channelID); err != nil {
		return nil, err
	}

	m := &domain.ChannelMember{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		UserID:    userID,
		Role:      "member",
		JoinedAt:  time.Now().UTC(),
	}

	if err := s.repo.AddMember(ctx, m); err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to add member", err)
	}

	log.Println("insode add memeber")

	s.NotifyAdminUserJoined(ctx, channelID, userID)
	return m, nil
}

func (s *channelService) RemoveMember(ctx context.Context, channelID, userID, requesterID string) error {

	member, err := s.repo.IsUserMember(ctx, channelID, userID)
	if err != nil || !member {
		return errors.New(errors.CodeInternal, "failed to identify member", err)
	}

	if err := s.repo.RemoveMember(ctx, channelID, userID); err != nil {
		return errors.New(errors.CodeInternal, "failed to remove member", err)
	}
	return nil
}

func (s *channelService) ListMembers(ctx context.Context, channelID string) ([]*domain.ChannelMember, error) {

	members, err := s.repo.ListChannelMembers(ctx, channelID)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "failed to get members", err)
	}
	return members, nil
}

func (s *channelService) CheckMembership(ctx context.Context, channelID, userID string) (bool, error) {
	return s.repo.IsUserMember(ctx, channelID, userID)
}

// =============================================================================
// REQUESTS (INVITES & JOINS)
// =============================================================================

func (s *channelService) SendInvite(ctx context.Context, targetUserID, channelID, senderID string) error {
	if exists := s.repo.CheckExistingRequest(ctx, targetUserID, channelID, "join"); exists {
		return errors.New(errors.CodeForbidden, "request already exists", nil)
	}

	isMember, _ := s.repo.IsUserMember(ctx, channelID, senderID)
	if !isMember {
		return errors.New(errors.CodeForbidden, "must be member to invite", nil)
	}
	isMember, _ = s.repo.IsUserMember(ctx, channelID, targetUserID)
	if isMember {
		return errors.New(errors.CodeForbidden, "already a member", nil)
	}

	req := &domain.Request{
		ID:        uuid.New().String(),
		UserID:    targetUserID,
		ChannelID: channelID,
		Type:      "invite",
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.CreateRequest(ctx, req); err != nil {
		return errors.New(errors.CodeInternal, "failed to send invite", err)
	}
	return nil
}

func (s *channelService) SendJoin(ctx context.Context, userID, channelID string) error {
	//TODO check same request exists

	if exists := s.repo.CheckExistingRequest(ctx, userID, channelID, "join"); exists {
		return errors.New(errors.CodeForbidden, "request already exists", nil)
	}
	isMember, _ := s.repo.IsUserMember(ctx, channelID, userID)
	if isMember {
		return errors.New(errors.CodeForbidden, "already a member", nil)
	}

	req := &domain.Request{
		ID:        uuid.New().String(),
		UserID:    userID,
		ChannelID: channelID,
		Type:      "join",
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.CreateRequest(ctx, req); err != nil {
		return errors.New(errors.CodeInternal, "failed to request join", err)
	}
	return nil
}

func (s *channelService) RespondToRequest(ctx context.Context, requestID, userID, status string) error {
	request, err := s.repo.UpdateRequestStatus(ctx, requestID, status)
	if err != nil {
		return errors.New(errors.CodeInternal, "update status failed", err)
	}

	log.Println("status: ", status)
	if request.Status == "accepted" {
		// member will
		log.Println("insode if")
		s.AddMember(ctx, request.ChannelID, request.UserID)
	}

	return nil
}

func (s *channelService) ListUserInvites(ctx context.Context, userID string) ([]*domain.Request, error) {
	return s.repo.ListPendingRequests(ctx, userID, "")
}

func (s *channelService) ListChannelJoins(ctx context.Context, channelID string) ([]*domain.Request, error) {
	return s.repo.ListPendingRequests(ctx, "", channelID)
}

// =============================================================================
// ADMIN OPERATIONS
// =============================================================================

func (s *channelService) AdminListChannels(ctx context.Context, limit, offset int32) ([]*domain.ChannelWithMembers, error) {
	return s.repo.AdminListChannels(ctx, limit, offset)
}

func (s *channelService) AdminFreezeChannel(ctx context.Context, channelID string, freeze bool, reason string) error {
	if err := s.repo.FreezeChannel(ctx, channelID, freeze); err != nil {
		return errors.New(errors.CodeInternal, "freeze update failed", err)
	}

	logger.Log.Info("admin updated channel freeze status",
		zap.String("channel_id", channelID),
		zap.Bool("freeze", freeze),
		zap.String("reason", reason))

	return nil
}

func (s *channelService) AdminDeleteChannel(ctx context.Context, channelID string) error {

	if err := s.repo.DeleteChannel(ctx, channelID); err != nil {
		return errors.New(errors.CodeInternal, "admin delete failed", err)
	}
	return nil
}

func (s *channelService) NotifyAdminUserJoined(ctx context.Context, channelID, newUserID string) error {
	// 1. Load Channel
	channel, err := s.repo.GetChannel(ctx, channelID)
	if err != nil || channel == nil {
		return errors.New(errors.CodeNotFound, "channel not found", err)
	}

	// 2. Load New User
	user, err := s.clients.Auth.GetUser(ctx, &authpb.GetUserRequest{Query: &authpb.GetUserRequest_UserId{newUserID}})
	if err != nil || user == nil {
		return errors.New(errors.CodeNotFound, "user not found", err)
	}

	// 3. Load Channel Admin
	admin, err := s.clients.Auth.GetUser(ctx, &authpb.GetUserRequest{Query: &authpb.GetUserRequest_UserId{channel.CreatedBy}})
	if err != nil || admin == nil {
		return errors.New(errors.CodeNotFound, "admin not found", err)
	}

	log.Println("insode notify")

	// 4. Prepare Email
	from := mail.NewEmail("StreamHub", "ak001mob@gmail.com")
	to := mail.NewEmail(admin.User.Username, admin.User.Email)

	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.Subject = "New Member Joined Your Channel"
	message.SetTemplateID(s.cfg.SendGrid.AdminInfo) // Template ID from SendGrid

	p := mail.NewPersonalization()
	p.AddTos(to)

	p.SetDynamicTemplateData("admin_name", admin.User.Username)
	p.SetDynamicTemplateData("new_user_name", user.User.Username)
	p.SetDynamicTemplateData("channel_name", channel.Name)
	p.SetDynamicTemplateData("joined_at", time.Now().Format("02 Jan 2006 15:04"))
	p.SetDynamicTemplateData("support_email", "support@streamhub.com")

	message.AddPersonalizations(p)

	client := sendgrid.NewSendClient(s.cfg.SendGrid.Key)
	resp, err := client.Send(message)

	if err != nil || resp.StatusCode >= 300 {
		return errors.New(errors.CodeInternal, "failed to send new member email", err)
	}

	logger.Log.Info("user joined informed to admin" + admin.User.Email)

	log.Println("insode notify")

	return nil
}

// for other services
func (s *channelService) GetChannelStorage(
	ctx context.Context,
	channelID string,
) (usedMB int64, limitMB int64, err error) {
	return s.repo.GetChannelStorage(ctx, channelID)

}
func (s *channelService) UpdateUsedMB(ctx context.Context, channelID string, usedMB int64) error {
	return s.repo.UpdateUsedMB(ctx, channelID, usedMB)

}
func (s *channelService) UpdateChannelPlan(ctx context.Context, channelID string, planID string, limitMB int64,
) error {

	return s.repo.UpdateChannelPlan(ctx, channelID, planID, limitMB)

}
