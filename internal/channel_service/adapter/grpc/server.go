package chatgrpc

import (
	"context"
	"io"
	"log"

	"github.com/ak-repo/stream-hub/gen/channelpb"
	"github.com/ak-repo/stream-hub/internal/channel_service/domain"
	"github.com/ak-repo/stream-hub/internal/channel_service/port"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ChannelServer struct {
	channelpb.UnimplementedChannelServiceServer
	service port.ChannelService
}

func NewChannelServer(service port.ChannelService) *ChannelServer {
	return &ChannelServer{service: service}
}

// Connect handles bidirectional streaming for real-time chat.
// - receives StreamRequest (Join or Message)
// - on Join: subscribes the client to channel messages and starts sending via stream.Send()
// - on Message: forwards to service.PostMessa
func (s *ChannelServer) Connect(stream channelpb.ChannelService_ConnectServer) error {
	ctx := stream.Context()
	errChan := make(chan error, 1)

	// create a context that cancels when stream is done
	// (stream.Context() already cancels on client disconnect)

	// We'll use a goroutine to receive messages from client and process them.
	// For each join we spawn a goroutine that listens on the service subscription and writes back to stream.
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

			// handle the oneof payload (Join or Message)
			switch payload := req.Payload.(type) {
			case *channelpb.StreamRequest_Join:
				join := payload.Join

				logger.Log.Info("Join received", zap.String("user", join.UserId), zap.String("channel", join.ChannelId))

				// subscribe and stream messages back to this client
				go func(userID, channelID string) {
					if err := s.streamMessagesToClient(stream, userID, channelID); err != nil {
						// streamMessagesToClient will log; here we also surface error
						logger.Log.Error("streamMessagesToClient error", zap.Error(err))
					}
				}(join.UserId, join.ChannelId)

			case *channelpb.StreamRequest_Message:
				msg := payload.Message

				// extract text or file attachment
				log.Println("msg:", msg.GetContent())

				var text string
				var attachment *domain.FileAttachment
				if txt := msg.GetContent(); txt != "" {
					text = txt

				} else if f := msg.GetFile(); f != nil {
					attachment = &domain.FileAttachment{
						ID:       uuid.New().String(),
						FileID:   f.GetFileId(),
						URL:      f.GetUrl(),
						MimeType: f.GetMimeType(),
						Size:     f.GetSize(),
					}
				}
				// log.Println(attachment)
				if _, err := s.service.PostMessage(ctx, msg.GetUserId(), msg.GetChannelId(), text, attachment); err != nil {
					logger.Log.Error("PostMessage failed", zap.Error(err))
				}
			default:
				logger.Log.Warn("unknown payload type in stream request")

			}
		}

	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}

}

// streamMessagesToClient subscribes to Redis for a channel and
// pushes messages back to the connected client via gRPC stream.
func (s *ChannelServer) streamMessagesToClient(stream channelpb.ChannelService_ConnectServer, userID, channelID string) error {
	ctx := stream.Context()
	msgChan, err := s.service.SubscribeToChannel(ctx, channelID)
	if err != nil {
		logger.Log.Error("SubscribeToChannel failed", zap.Error(err))
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case m, ok := <-msgChan:
			if !ok {
				return nil
			}
			created := &channelpb.MessageCreated{
				MessageId:   m.ID,
				SenderId:    m.SenderID,
				Content:     m.Content,
				TimestampMs: m.CreatedAt.UnixMilli(),
			}

			if m.Attachment != nil {
				created.Attachment = &channelpb.FileAttachment{
					FileId:   m.Attachment.FileID,
					Url:      m.Attachment.URL,
					MimeType: m.Attachment.MimeType,
					Size:     m.Attachment.Size,
				}
			}

			resp := &channelpb.StreamResponse{
				ChannelId:   m.ChannelID,
				TimestampMs: m.CreatedAt.UnixMilli(),
				Event:       &channelpb.StreamResponse_Created{Created: created},
			}

			if err := stream.Send(resp); err != nil {
				logger.Log.Error("failed to send message to client", zap.Error(err))
				return err
			}

		}
	}

}

// sent history to user
func (s *ChannelServer) ListMessages(ctx context.Context, req *channelpb.ListMessagesRequest) (*channelpb.ListMessagesResponse, error) {
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 50
	}

	messages, err := s.service.GetHistory(ctx, req.GetChannelId(), limit, int(req.Offset))
	if err != nil {
		return nil, err
	}

	respMessages := make([]*channelpb.MessageInfo, 0, len(messages))
	for _, m := range messages {
		var attach *channelpb.FileAttachment
		if m.Attachment != nil {
			attach = &channelpb.FileAttachment{
				FileId:   m.Attachment.FileID,
				Url:      m.Attachment.URL,
				MimeType: m.Attachment.MimeType,
				Size:     m.Attachment.Size,
			}
		}

		respMessages = append(respMessages, &channelpb.MessageInfo{
			MessageId:   m.ID,
			SenderId:    m.SenderID,
			Content:     m.Content,
			TimestampMs: m.CreatedAt.UnixMilli(),
			Username:    m.Username,
			Attachment:  attach,
		})
	}
	return &channelpb.ListMessagesResponse{
		Messages: respMessages,
	}, nil
}

// func (s *ChannelServer) EditMessage(ctx context.Context, req *channelpb.EditMessageRequest) (*channelpb.EditMessageResponse, error) {
// 	if err := s.service.EditMessage(ctx, req.GetMessageId(), req.GetEditorId(), req.GetNewContent()); err != nil {
// 		return &channelpb.EditMessageResponse{Success: false}, err
// 	}
// 	return &channelpb.EditMessageResponse{Success: true}, nil
// }

// func (s *ChannelServer) DeleteMessage(ctx context.Context, req *channelpb.DeleteMessageRequest) (*channelpb.DeleteMessageResponse, error) {
// 	if err := s.service.DeleteMessage(ctx, req.GetMessageId(), req.GetRequesterId()); err != nil {
// 		return &channelpb.DeleteMessageResponse{Success: false}, err
// 	}
// 	return &channelpb.DeleteMessageResponse{Success: true}, nil
// }

// CreateChannel handles channel creation requests.
func (s *ChannelServer) CreateChannel(ctx context.Context, req *channelpb.CreateChannelRequest) (*channelpb.CreateChannelResponse, error) {
	ch, err := s.service.CreateChannel(ctx, req.GetName(), req.GetCreatorId())
	if err != nil {
		return nil, err
	}

	return &channelpb.CreateChannelResponse{
		ChannelId:   ch.ID,
		Name:        ch.Name,
		CreatedBy:   ch.CreatedBy,
		CreatedAtMs: ch.CreatedAt.UnixMilli(),
	}, nil
}

// GetChannel retrieves channel information.
func (s *ChannelServer) GetChannel(ctx context.Context, req *channelpb.GetChannelRequest) (*channelpb.GetChannelResponse, error) {
	ch, err := s.service.GetChannel(ctx, req.GetChannelId())
	if err != nil {
		return nil, err
	}

	resp := &channelpb.ChannelInfo{
		ChannelId:   ch.ID,
		Name:        ch.Name,
		CreatedBy:   ch.CreatedBy,
		CreatedAtMs: ch.CreatedAt.UnixMilli(),
	}
	return &channelpb.GetChannelResponse{
		Channel: resp,
	}, nil
}

// Get All channels of a user
func (s *ChannelServer) ListChannels(ctx context.Context, req *channelpb.ListChannelsRequest) (*channelpb.ListChannelsResponse, error) {

	// Call service to get grouped channels
	chans, err := s.service.ListChannels(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	resp := &channelpb.ListChannelsResponse{
		Channels: []*channelpb.ChannelInfo{},
	}

	for _, ch := range chans {

		// Convert members
		var members []*channelpb.MemberInfo
		for _, m := range ch.Members {
			members = append(members, &channelpb.MemberInfo{
				UserId:     m.UserID,
				JoinedAtMs: m.JoinedAt.UnixMilli(),
			})
		}

		// Convert channel
		chanInfo := &channelpb.ChannelInfo{
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
func (s *ChannelServer) AddMember(ctx context.Context, req *channelpb.AddMemberRequest) (*channelpb.AddMemberResponse, error) {
	m, err := s.service.AddMember(ctx, req.GetChannelId(), req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &channelpb.AddMemberResponse{
		Member: &channelpb.MemberInfo{
			UserId:     m.UserID,
			Username:   m.Username,
			JoinedAtMs: m.JoinedAt.UnixMilli(),
		},
	}, nil
}

// RemoveMember removes a user from a channel.
func (s *ChannelServer) RemoveMember(ctx context.Context, req *channelpb.RemoveMemberRequest) (*channelpb.RemoveMemberResponse, error) {
	err := s.service.RemoveMember(ctx, req.GetChannelId(), req.GetUserId())
	return &channelpb.RemoveMemberResponse{Success: err == nil}, err
}

// ListMembers returns all members of a channel.
func (s *ChannelServer) ListMembers(ctx context.Context, req *channelpb.ListMembersRequest) (*channelpb.ListMembersResponse, error) {
	members, err := s.service.ListMembers(ctx, req.GetChannelId())
	if err != nil {
		return nil, err
	}

	channelpbMembers := make([]*channelpb.MemberInfo, len(members))
	for i, m := range members {
		channelpbMembers[i] = &channelpb.MemberInfo{
			UserId:     m.UserID,
			Username:   m.Username,
			JoinedAtMs: m.JoinedAt.UnixMilli(),
		}
	}

	return &channelpb.ListMembersResponse{Members: channelpbMembers}, nil
}

// Delete channel only by owner
func (s *ChannelServer) DeleteChannel(ctx context.Context, req *channelpb.DeleteChannelRequest) (*channelpb.DeleteChannelResponse, error) {

	if err := s.service.DeleteChannel(ctx, req.ChannelId, req.RequesterId); err != nil {
		return nil, err
	}

	return &channelpb.DeleteChannelResponse{Success: true}, nil

}

// // --- ADMIN OPERATIONS ---
// func (s *ChannelServer) AdminStats(ctx context.Context, req *channelpb.AdminStatsRequest) (*channelpb.AdminStatsResponse, error) {
// 	stats, err := s.service.GetAdminStats(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &channelpb.AdminStatsResponse{
// 		TotalUsers:       stats.TotalUsers,
// 		TotalChannels:    stats.TotalChannels,
// 		TotalMessages:    stats.TotalMessages,
// 		TotalStorageBytes: stats.TotalStorageBytes,
// 	}, nil
// }

// func (s *ChannelServer) AdminListAllChannels(ctx context.Context, req *channelpb.AdminListAllChannelsRequest) (*channelpb.AdminListAllChannelsResponse, error) {
// 	channels, err := s.service.ListAllChannels(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	resp := make([]*channelpb.ChannelInfo, 0, len(channels))
// 	for _, ch := range channels {
// 		members := make([]*channelpb.MemberInfo, 0, len(ch.Members))
// 		for _, m := range ch.Members {
// 			members = append(members, &channelpb.MemberInfo{
// 				UserId:     m.UserID,
// 				Username:   m.Username,
// 				JoinedAtMs: m.JoinedAt.UnixMilli(),
// 				Role:       channelpb.UserRole(m.Role),
// 			})
// 		}
// 		resp = append(resp, &channelpb.ChannelInfo{
// 			ChannelId:   ch.Channel.ID,
// 			Name:        ch.Channel.Name,
// 			CreatedBy:   ch.Channel.CreatedBy,
// 			CreatedAtMs: ch.Channel.CreatedAt.UnixMilli(),
// 			Members:     members,
// 		})
// 	}

// 	return &channelpb.AdminListAllChannelsResponse{Channels: resp}, nil
// }

// func (s *ChannelServer) AdminListAllUsers(ctx context.Context, req *channelpb.AdminListAllUsersRequest) (*channelpb.AdminListAllUsersResponse, error) {
// 	users, err := s.service.ListAllUsers(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	resp := make([]*channelpb.MemberInfo, 0, len(users))
// 	for _, u := range users {
// 		resp = append(resp, &channelpb.MemberInfo{
// 			UserId:     u.UserID,
// 			Username:   u.Username,
// 			JoinedAtMs: u.JoinedAt.UnixMilli(),
// 			Role:       channelpb.UserRole(u.Role),
// 		})
// 	}
// 	return &channelpb.AdminListAllUsersResponse{Users: resp}, nil
// }

// func (s *ChannelServer) AdminBanUser(ctx context.Context, req *channelpb.AdminBanUserRequest) (*channelpb.AdminBanUserResponse, error) {
// 	if err := s.service.BanUser(ctx, req.GetUserId(), req.GetBannedBy()); err != nil {
// 		return &channelpb.AdminBanUserResponse{Success: false}, err
// 	}
// 	return &channelpb.AdminBanUserResponse{Success: true}, nil
// }

// func (s *ChannelServer) AdminUnbanUser(ctx context.Context, req *channelpb.AdminUnbanUserRequest) (*channelpb.AdminUnbanUserResponse, error) {
// 	if err := s.service.UnbanUser(ctx, req.GetUserId(), req.GetUnbannedBy()); err != nil {
// 		return &channelpb.AdminUnbanUserResponse{Success: false}, err
// 	}
// 	return &channelpb.AdminUnbanUserResponse{Success: true}, nil
// }
