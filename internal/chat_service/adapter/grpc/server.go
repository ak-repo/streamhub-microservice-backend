package chatgrpc

import (
	"context"
	"fmt"
	"io"

	"github.com/ak-repo/stream-hub/gen/chatpb"
	"github.com/ak-repo/stream-hub/internal/chat_service/port"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"go.uber.org/zap"
)

type ChatServer struct {
	chatpb.UnimplementedChatServiceServer
	service port.ChatService
}

func NewChatServer(service port.ChatService) *ChatServer {
	return &ChatServer{service: service}
}

// Connect handles bidirectional streaming for real-time chat.
// This method manages both incoming (JOIN/MESSAGE) and outgoing (broadcast) flows.
func (s *ChatServer) Connect(stream chatpb.ChatService_ConnectServer) error {
	ctx := stream.Context()
	errChan := make(chan error, 1)

	// Goroutine to receive client messages (JOIN/MESSAGE)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Log.Error("Recovered from panic in Connect receiver")
				errChan <- nil
			}
		}()
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				errChan <- nil
				return
			}
			if err != nil {
				logger.Log.Error("Stream receive error: " + err.Error())
				errChan <- err
				return
			}

			switch payload := req.Payload.(type) {
			case *chatpb.StreamRequest_Join:
				// User joins a channel - start streaming messages to them
				logger.Log.Info(fmt.Sprintf("User %s joining channel %s", payload.Join.UserId, payload.Join.ChannelId))
				go s.streamMessagesToClient(stream, payload.Join.ChannelId)
			case *chatpb.StreamRequest_Message:
				_, err := s.service.PostMessage(
					ctx,
					payload.Message.UserId,
					payload.Message.ChannelId,
					payload.Message.Content,
				)
				if err != nil {
					logger.Log.Error("Error posting message", zap.Error(err))
					// Don't kill the stream - just log and continue

				}

			}
		}

	}()
	return <-errChan

}

// streamMessagesToClient subscribes to Redis for a channel and
// pushes messages back to the connected client via gRPC stream.
func (s *ChatServer) streamMessagesToClient(stream chatpb.ChatService_ConnectServer, channelID string) {
	msgChan, err := s.service.SubscribeToChannel(stream.Context(), channelID)
	if err != nil {
		logger.Log.Error("Failed to subscribe to channel", zap.Error(err))
		return
	}

	for msg := range msgChan {
		resp := &chatpb.StreamResponse{
			MessageId:   msg.ID,
			ChannelId:   msg.ChannelID,
			SenderId:    msg.SenderID,
			Content:     msg.Content,
			TimestampMs: msg.CreatedAt.UnixMilli(), // FIXED: Convert time.Time to milliseconds
		}
		if err := stream.Send(resp); err != nil {
			logger.Log.Error("Failed to send message to client:", zap.Error(err))
		}
	}
}

// sent history to user
func (s *ChatServer) ListMessages(ctx context.Context, req *chatpb.ListMessagesRequest) (*chatpb.ListMessagesResponse, error) {

	messages, err := s.service.GetHistory(ctx, req.ChannelId)
	if err != nil {
		return nil, err
	}

	resp := []*chatpb.MessageInfo{}
	for _, msg := range messages {
		m := &chatpb.MessageInfo{
			MessageId:   msg.ID,
			SenderId:    msg.SenderID,
			Content:     msg.Content,
			TimestampMs: msg.CreatedAt.Unix(),
			Username:    msg.Username,
		}
		resp = append(resp, m)
	}
	return &chatpb.ListMessagesResponse{Messages: resp}, nil
}

// CreateChannel handles channel creation requests.
func (s *ChatServer) CreateChannel(ctx context.Context, req *chatpb.CreateChannelRequest) (*chatpb.CreateChannelResponse, error) {
	ch, err := s.service.CreateChannel(ctx, req.GetName(), req.GetCreatorId())
	if err != nil {
		return nil, err
	}

	return &chatpb.CreateChannelResponse{
		ChannelId:   ch.ID,
		Name:        ch.Name,
		CreatedBy:   ch.CreatedBy,
		CreatedAtMs: ch.CreatedAt.UnixMilli(),
	}, nil
}

// GetChannel retrieves channel information.
func (s *ChatServer) GetChannel(ctx context.Context, req *chatpb.GetChannelRequest) (*chatpb.GetChannelResponse, error) {
	ch, err := s.service.GetChannel(ctx, req.GetChannelId())
	if err != nil {
		return nil, err
	}

	resp := &chatpb.ChannelInfo{
		ChannelId:   ch.ID,
		Name:        ch.Name,
		CreatedBy:   ch.CreatedBy,
		CreatedAtMs: ch.CreatedAt.UnixMilli(),
	}
	return &chatpb.GetChannelResponse{
		Channel: resp,
	}, nil
}

// Get All channels of a user
func (s *ChatServer) ListChannels(ctx context.Context, req *chatpb.ListChannelsRequest) (*chatpb.ListChannelsResponse, error) {

	// Call service to get grouped channels
	chans, err := s.service.ListChannels(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	resp := &chatpb.ListChannelsResponse{
		Channels: []*chatpb.ChannelInfo{},
	}

	for _, ch := range chans {

		// Convert members
		var members []*chatpb.MemberInfo
		for _, m := range ch.Members {
			members = append(members, &chatpb.MemberInfo{
				UserId:     m.UserID,
				JoinedAtMs: m.JoinedAt.UnixMilli(),
			})
		}

		// Convert channel
		chanInfo := &chatpb.ChannelInfo{
			ChannelId:   ch.Channel.ID,
			Name:        ch.Channel.Name,
			CreatedBy:   ch.Channel.CreatedBy,
			CreatedAtMs: ch.Channel.CreatedAt.UnixMilli(),
			Members:     members,
		}

		resp.Channels = append(resp.Channels, chanInfo)
	}

	return resp, nil
}

// AddMember adds a user to a channel.
func (s *ChatServer) AddMember(ctx context.Context, req *chatpb.AddMemberRequest) (*chatpb.AddMemberResponse, error) {
	m, err := s.service.AddMember(ctx, req.GetChannelId(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &chatpb.AddMemberResponse{
		MemberId:   m.ID,
		ChannelId:  m.ChannelID,
		UserId:     m.UserID,
		JoinedAtMs: m.JoinedAt.UnixMilli(),
	}, nil
}

// RemoveMember removes a user from a channel.
func (s *ChatServer) RemoveMember(ctx context.Context, req *chatpb.RemoveMemberRequest) (*chatpb.RemoveMemberResponse, error) {
	err := s.service.RemoveMember(ctx, req.GetChannelId(), req.GetUserId())
	return &chatpb.RemoveMemberResponse{Success: err == nil}, err
}

// ListMembers returns all members of a channel.
func (s *ChatServer) ListMembers(ctx context.Context, req *chatpb.ListMembersRequest) (*chatpb.ListMembersResponse, error) {
	members, err := s.service.ListMembers(ctx, req.GetChannelId())
	if err != nil {
		return nil, err
	}

	chatpbMembers := make([]*chatpb.MemberInfo, len(members))
	for i, m := range members {
		chatpbMembers[i] = &chatpb.MemberInfo{
			UserId:     m.UserID,
			Username:   m.Username,
			JoinedAtMs: m.JoinedAt.UnixMilli(),
		}
	}

	return &chatpb.ListMembersResponse{Members: chatpbMembers}, nil
}
