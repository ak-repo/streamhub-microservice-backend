package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ak-repo/stream-hub/gen/channelpb"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"github.com/gofiber/websocket/v2"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

type WSMessage struct {
	Type      string `json:"type"` // JOIN or MESSAGE
	UserID    string `json:"userId"`
	ChannelID string `json:"channelId"`
	Content   string `json:"content"` // Only for MESSAGE
}

func (h *ChannelHandler) WsHandler(conn *websocket.Conn) {
	// ✔ Validate user
	userID := conn.Query("userId")
	if userID == "" {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"userId is required"}`))
		conn.Close()
		return
	}

	logger.Log.Info("WebSocket connected", zap.String("user", userID))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ✔ Open gRPC stream
	stream, err := h.client.Connect(ctx)
	if err != nil {
		logger.Log.Error("failed to connect to grpc stream", zap.Error(err))
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"internal server error"}`))
		conn.Close()
		return
	}
	defer stream.CloseSend()

	// -------------------------------------------------------------------------
	// WS → GRPC
	// -------------------------------------------------------------------------
	go func() {
		defer cancel()

		conn.SetReadLimit(maxMessageSize)
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})

		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				logger.Log.Warn("ws read error", zap.Error(err))
				return
			}

			var m WSMessage
			if err := json.Unmarshal(raw, &m); err != nil {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"invalid json"}`))
				continue
			}

			// Inherit userID from query if missing
			if m.UserID == "" {
				m.UserID = userID
			}

			var req *channelpb.StreamRequest = &channelpb.StreamRequest{}

			switch m.Type {

			// -----------------------------------------------------------------
			// JOIN → StreamConnect
			// -----------------------------------------------------------------
			case "JOIN":
				req.Payload = &channelpb.StreamRequest_Connect{
					Connect: &channelpb.StreamConnect{
						UserId:    m.UserID,
						ChannelId: m.ChannelID,
					},
				}

			// -----------------------------------------------------------------
			// MESSAGE → StreamSendMessage
			// -----------------------------------------------------------------
			case "MESSAGE":
				req.Payload = &channelpb.StreamRequest_Message{
					Message: &channelpb.StreamSendMessage{
						Content: m.Content,
					},
				}

			default:
				_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"unknown type"}`))
				continue
			}

			if err := stream.Send(req); err != nil {
				logger.Log.Error("grpc stream send error", zap.Error(err))
				return
			}
		}
	}()

	// -------------------------------------------------------------------------
	// PING KEEPALIVE
	// -------------------------------------------------------------------------
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// -------------------------------------------------------------------------
	// GRPC → WS
	// -------------------------------------------------------------------------
	for {
		res, err := stream.Recv()
		if err != nil {
			break
		}

		data, _ := json.Marshal(res)

		_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			break
		}
	}

	_ = conn.Close()
}
