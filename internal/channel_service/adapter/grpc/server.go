package channelgrpc

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/ak-repo/stream-hub/gen/channelpb"
	"github.com/ak-repo/stream-hub/internal/channel_service/domain"
	"github.com/ak-repo/stream-hub/internal/channel_service/port"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// =============================================================================
// SERVER FACTORY
// =============================================================================

type Server struct {
	channelpb.UnimplementedChannelServiceServer
	channelpb.UnimplementedAdminChannelServiceServer
	service port.ChannelService
}

// NewServer returns a server that implements BOTH ChannelService and AdminChannelService
func NewServer(service port.ChannelService) *Server {
	return &Server{service: service}
}

// =============================================================================
// 1. REAL-TIME STREAMING
// =============================================================================

func (s *Server) Connect(stream channelpb.ChannelService_ConnectServer) error {
	ctx := stream.Context()

	// State variables for this specific connection
	var currentUserID string
	var currentChannelID string
	var isConnected bool

	// Channel to coordinate sending messages back to the client
	// We use a separate goroutine for sending to avoid blocking reads
	sendChan := make(chan *channelpb.StreamResponse, 100)
	errChan := make(chan error, 1)

	// 1. Start Sender Goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case resp := <-sendChan:
				if err := stream.Send(resp); err != nil {
					logger.Log.Error("failed to send to stream", zap.Error(err))
					errChan <- err
					return
				}
			}
		}
	}()

	// 2. Main Receiver Loop
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch payload := req.Payload.(type) {

		// A. Handle Connection Setup
		case *channelpb.StreamRequest_Connect:
			if isConnected {
				continue // Already connected, ignore or handle as re-connect
			}
			conn := payload.Connect
			currentUserID = conn.UserId
			currentChannelID = conn.ChannelId
			isConnected = true

			// Subscribe to the PubSub via the Service Layer
			msgChan, err := s.service.SubscribeToChannel(ctx, currentChannelID)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to subscribe: %v", err)
			}

			// Start listening to PubSub events
			go func() {
				for msg := range msgChan {
					// Map Domain Message to Proto
					pbMsg := mapMessageToProto(msg)
					sendChan <- &channelpb.StreamResponse{
						ChannelId: currentChannelID,
						Timestamp: time.Now().UnixMilli(),
						Event: &channelpb.StreamResponse_MessageCreated{
							MessageCreated: pbMsg,
						},
					}
				}
			}()

			logger.Log.Info("Stream connected", zap.String("user", currentUserID), zap.String("channel", currentChannelID))

		// B. Handle Incoming Messages
		case *channelpb.StreamRequest_Message:
			if !isConnected {
				return status.Error(codes.FailedPrecondition, "must send StreamConnect before sending messages")
			}

			msgReq := payload.Message

			// Call Service
			_, err := s.service.PostMessage(ctx, currentUserID, currentChannelID, msgReq.Content)
			if err != nil {
				logger.Log.Error("failed to post message", zap.Error(err))
				// Optionally send an error response back via stream if your proto supports it
			}
		}
	}
}

// =============================================================================
// 2. CHANNEL CRUD
// =============================================================================

// TEST
func (s *Server) CreateChannel(ctx context.Context, req *channelpb.CreateChannelRequest) (*channelpb.CreateChannelResponse, error) {

	ch, err := s.service.CreateChannel(ctx, req.Name, req.Description, req.Visibility, req.CreatorId)
	if err != nil {
		return nil, err
	}

	return &channelpb.CreateChannelResponse{Channel: mapChannelToProto(ch)}, nil
}

// TEST
func (s *Server) GetChannel(ctx context.Context, req *channelpb.GetChannelRequest) (*channelpb.GetChannelResponse, error) {
	ch, err := s.service.GetChannel(ctx, req.ChannelId)
	if err != nil {
		return nil, err
	}
	return &channelpb.GetChannelResponse{Channel: mapChannelToProto(ch)}, nil
}

// TEST
func (s *Server) ListUserChannels(ctx context.Context, req *channelpb.ListUserChannelsRequest) (*channelpb.ListUserChannelsResponse, error) {
	channels, err := s.service.ListUserChannels(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	var pbChannels []*channelpb.Channel
	for _, c := range channels {
		pbChannels = append(pbChannels, mapChannelToProto(c))
	}
	return &channelpb.ListUserChannelsResponse{Channels: pbChannels}, nil
}

func (s *Server) DeleteChannel(ctx context.Context, req *channelpb.DeleteChannelRequest) (*channelpb.DeleteChannelResponse, error) {
	err := s.service.DeleteChannel(ctx, req.ChannelId, req.RequesterId)
	if err != nil {
		return nil, err
	}
	return &channelpb.DeleteChannelResponse{Success: true}, nil
}

func (s *Server) SearchChannels(ctx context.Context, req *channelpb.SearchChannelRequest) (*channelpb.SearchChannelResponse, error) {

	channels, err := s.service.SearchChannels(ctx, req.Query, req.Pagination.Limit, req.Pagination.Offset)
	if err != nil {
		return nil, err
	}

	var resp []*channelpb.Channel
	for _, ch := range channels {
		resp = append(resp, mapChannelToProto(ch))
	}
	return &channelpb.SearchChannelResponse{Channels: resp}, nil
}

// =============================================================================
// 3. MEMBER MANAGEMENT
// =============================================================================

func (s *Server) AddMember(ctx context.Context, req *channelpb.AddMemberRequest) (*channelpb.AddMemberResponse, error) {
	member, err := s.service.AddMember(ctx, req.ChannelId, req.UserId)
	if err != nil {
		return nil, err
	}
	return &channelpb.AddMemberResponse{Member: mapMemberToProto(member)}, nil
}

func (s *Server) RemoveMember(ctx context.Context, req *channelpb.RemoveMemberRequest) (*channelpb.RemoveMemberResponse, error) {
	err := s.service.RemoveMember(ctx, req.ChannelId, req.UserId, req.RemovedBy)
	if err != nil {
		return nil, err
	}
	return &channelpb.RemoveMemberResponse{Success: true}, nil
}

func (s *Server) ListMembers(ctx context.Context, req *channelpb.ListMembersRequest) (*channelpb.ListMembersResponse, error) {
	members, err := s.service.ListMembers(ctx, req.ChannelId)
	if err != nil {
		return nil, err
	}

	var pbMembers []*channelpb.ChannelMember
	for _, m := range members {
		pbMembers = append(pbMembers, mapMemberToProto(m))
	}
	return &channelpb.ListMembersResponse{Members: pbMembers}, nil
}

// =============================================================================
// 4. REQUEST FLOW (INVITES & JOINS)
// =============================================================================

func (s *Server) SendInvite(ctx context.Context, req *channelpb.SendInviteRequest) (*channelpb.SendInviteResponse, error) {

	err := s.service.SendInvite(ctx, req.TargetUserId, req.ChannelId, req.SenderId)
	if err != nil {
		return nil, err
	}
	return &channelpb.SendInviteResponse{Success: true}, nil
}

func (s *Server) SendJoin(ctx context.Context, req *channelpb.SendJoinRequest) (*channelpb.SendJoinResponse, error) {
	err := s.service.SendJoin(ctx, req.UserId, req.ChannelId)
	if err != nil {
		return nil, err
	}
	return &channelpb.SendJoinResponse{RequestId: uuid.NewString(), Message: "Join request sent"}, nil
}

func (s *Server) UpdateRequestStatus(ctx context.Context, req *channelpb.UpdateRequestStatusRequest) (*channelpb.UpdateRequestStatusResponse, error) {
	err := s.service.RespondToRequest(ctx, req.RequestId, req.UserId, req.Status)
	if err != nil {
		return nil, err
	}
	return &channelpb.UpdateRequestStatusResponse{Success: true}, nil
}

func (s *Server) ListUserInvites(ctx context.Context, req *channelpb.ListUserInvitesRequest) (*channelpb.ListUserInvitesResponse, error) {
	requests, err := s.service.ListUserInvites(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	log.Println("reqs:", requests)
	return &channelpb.ListUserInvitesResponse{Requests: mapRequestsToProto(requests)}, nil
}

func (s *Server) ListChannelJoins(ctx context.Context, req *channelpb.ListChannelJoinsRequest) (*channelpb.ListChannelJoinsResponse, error) {
	requests, err := s.service.ListChannelJoins(ctx, req.ChannelId)
	if err != nil {
		return nil, err
	}
	return &channelpb.ListChannelJoinsResponse{Requests: mapRequestsToProto(requests)}, nil
}

// =============================================================================
// 5. CHAT HISTORY
// =============================================================================

func (s *Server) ListMessages(ctx context.Context, req *channelpb.ListMessagesRequest) (*channelpb.ListMessagesResponse, error) {
	msgs, err := s.service.GetHistory(ctx, req.ChannelId, int(req.Pagination.Limit), int(req.Pagination.Offset))
	if err != nil {
		return nil, err
	}

	var pbMsgs []*channelpb.ChatMessage
	for _, m := range msgs {
		pbMsgs = append(pbMsgs, mapMessageToProto(m))
	}
	return &channelpb.ListMessagesResponse{Messages: pbMsgs}, nil
}

// =============================================================================
// 6. ADMIN SERVICE (Implemented on the same struct)
// =============================================================================

func (s *Server) AdminListChannels(ctx context.Context, req *channelpb.AdminListChannelsRequest) (*channelpb.AdminListChannelsResponse, error) {
	if req.Limit >= 0 {
		req.Limit = 10
	}

	channels, err := s.service.AdminListChannels(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	var resp []*channelpb.ChannelWithMembers
	for _, c := range channels {

		ch := &channelpb.ChannelWithMembers{
			Channel: mapChannelToProto(c.Channel),
		}
		for _, m := range c.Members {
			ch.Members = append(ch.Members, mapMemberToProto(m))
		}
		resp = append(resp, ch)
	}
	return &channelpb.AdminListChannelsResponse{Channels: resp}, nil
}

func (s *Server) AdminFreezeChannel(ctx context.Context, req *channelpb.AdminFreezeChannelRequest) (*channelpb.AdminFreezeChannelResponse, error) {
	err := s.service.AdminFreezeChannel(ctx, req.ChannelId, req.Freeze, req.Reason)
	if err != nil {
		return nil, err
	}
	return &channelpb.AdminFreezeChannelResponse{Success: true}, nil
}

func (s *Server) AdminDeleteChannel(ctx context.Context, req *channelpb.AdminDeleteChannelRequest) (*channelpb.AdminDeleteChannelResponse, error) {

	err := s.service.AdminDeleteChannel(ctx, req.ChannelId)
	if err != nil {
		return nil, err
	}
	return &channelpb.AdminDeleteChannelResponse{Success: true}, nil
}

// =============================================================================
// HELPERS & MAPPERS
// =============================================================================

func mapChannelToProto(c *domain.Channel) *channelpb.Channel {
	if c == nil {
		return nil
	}
	return &channelpb.Channel{
		Id:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Visibility:  c.Visibility,
		OwnerId:     c.CreatedBy,
		CreatedAt:   c.CreatedAt.UnixMilli(),
		IsFrozen:    c.IsFrozen,
		OwnerName:   c.OwnerName,
	}
}

func mapMemberToProto(m *domain.ChannelMember) *channelpb.ChannelMember {
	return &channelpb.ChannelMember{
		UserId:    m.UserID,
		Username:  m.Username,
		ChannelId: m.ChannelID,
		Role:      m.Role,
		JoinedAt:  m.JoinedAt.UnixMilli(),
	}
}

func mapMessageToProto(m *domain.Message) *channelpb.ChatMessage {
	msg := &channelpb.ChatMessage{
		Id:             m.ID,
		ChannelId:      m.ChannelID,
		SenderId:       m.SenderID,
		SenderUsername: m.Username,
		Content:        m.Content,
		CreatedAt:      m.CreatedAt.UnixMilli(),
	}
	return msg
}

func mapRequestsToProto(reqs []*domain.Request) []*channelpb.MembershipRequest {
	var res []*channelpb.MembershipRequest
	for _, r := range reqs {
		res = append(res, &channelpb.MembershipRequest{
			RequestId: r.ID,
			UserId:    r.UserID,
			ChannelId: r.ChannelID,
			Type:      r.Type,   // String in proto now
			Status:    r.Status, // String in proto now
			CreatedAt: r.CreatedAt.Unix(),
		})

	}
	return res
}
