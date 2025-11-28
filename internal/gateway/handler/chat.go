package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/ak-repo/stream-hub/gen/chatpb"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"go.uber.org/zap"
)

type ChatHandler struct {
	client chatpb.ChatServiceClient
}

func NewChatHandler(client chatpb.ChatServiceClient) *ChatHandler {
	return &ChatHandler{client: client}
}

const (
	HTTPPort        = ":8080"
	ChatServiceAddr = "localhost:50053"

	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

type WSMessage struct {
	Type      string `json:"type"`
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
	Content   string `json:"content"`
}

func (h *ChatHandler) CreateChannel(c *fiber.Ctx) error {
	var req struct {
		Name      string `json:"name"`
		CreatorID string `json:"creator_id"`
	}
	json.Unmarshal(c.Body(), &req)

	log.Println(req)
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid body")
	}

	resp, err := h.client.CreateChannel(context.Background(), &chatpb.CreateChannelRequest{
		Name:      req.Name,
		CreatorId: req.CreatorID,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(resp)
}

func (h *ChatHandler) JoinChannel(c *fiber.Ctx) error {
	var req struct {
		ChannelID string `json:"channel_id"`
		UserID    string `json:"user_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid body")
	}

	resp, err := h.client.AddMember(context.Background(), &chatpb.AddMemberRequest{
		ChannelId: req.ChannelID,
		UserId:    req.UserID,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(resp)
}

// ---------------WebSocket Handler -----------------------
func (h *ChatHandler) WsHandler(c *websocket.Conn) {
	userID := c.Query("user_id")
	if userID == "" {
		c.WriteMessage(websocket.TextMessage, []byte(`{"error":"userid required"}`))
		c.Close()
		return
	}

	logger.Log.Info(fmt.Sprintf("WebSocket connected: user=%s", userID))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := h.client.Connect(ctx)
	if err != nil {
		logger.Log.Error("stream error ", zap.Error(err))
		c.Close()
		return
	}

	// Reader goroutine: WS → gRPC
	go func() {
		c.SetReadLimit(maxMessageSize)
		c.SetReadDeadline(time.Now().Add(pongWait))
		c.SetPongHandler(func(string) error {
			c.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				cancel()
				return
			}
			var wsMsg WSMessage
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				continue
			}

			grpcReq := &chatpb.StreamRequest{}

			switch wsMsg.Type {
			case "JOIN":
				grpcReq.Payload = &chatpb.StreamRequest_Join{
					Join: &chatpb.JoinPayload{
						UserId:    userID,
						ChannelId: wsMsg.ChannelID,
					},
				}
			case "MESSAGE":
				grpcReq.Payload = &chatpb.StreamRequest_Message{
					Message: &chatpb.MessagePayload{
						UserId:    userID,
						ChannelId: wsMsg.ChannelID,
						Content:   wsMsg.Content,
					},
				}
			}
			if err := stream.Send(grpcReq); err != nil {
				return
			}
		}

	}()

	// Ping (keepalive) goroutine
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	go func() {
		select {
		case <-ticker.C:
			c.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait))
		case <-ctx.Done():
			return
		}
	}()

	// Writer: gRPC → WS
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		data, _ := json.Marshal(resp)
		c.WriteMessage(websocket.TextMessage, data)
	}

	logger.Log.Info(fmt.Sprintf("WebSocket disconnected: user=%s", userID))

}
