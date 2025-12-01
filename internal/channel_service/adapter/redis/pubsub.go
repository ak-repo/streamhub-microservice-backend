package chatredis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ak-repo/stream-hub/internal/channel_service/domain"
	"github.com/ak-repo/stream-hub/internal/channel_service/port"
	"github.com/redis/go-redis/v9"
)

const RedisChannelPrefix = "chat:channel:"

type redisPubSub struct {
	client *redis.Client
}

func NewRedisPubSub(client *redis.Client) port.PubSub {
	return &redisPubSub{client: client}
}

// Publish broadcasts a message to all subscribers of a channel.
// Redis pub/sub enables real-time delivery to all connected clients
// across multiple service instances.
func (r *redisPubSub) Publish(ctx context.Context, channelID string, msg *domain.Message) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	topic := fmt.Sprintf("%s%s", RedisChannelPrefix, channelID)
	return r.client.Publish(ctx, topic, payload).Err()
}

// Subscribe creates a real-time message stream for a channel.
// Returns a Go channel that receives messages as they're published.
func (r *redisPubSub) Subscribe(ctx context.Context, channelID string) (<-chan *domain.Message, error) {
	topic := fmt.Sprintf("%s%s", RedisChannelPrefix, channelID)
	pubsub := r.client.Subscribe(ctx, topic)

	msgChan := make(chan *domain.Message, 10) // Buffered to prevent blocking

	go func() {
		defer pubsub.Close()
		defer close(msgChan)

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case redisMsg, ok := <-ch:
				if !ok {
					return
				}

				var domainMsg domain.Message
				if err := json.Unmarshal([]byte(redisMsg.Payload), &domainMsg); err == nil {
					select {
					case msgChan <- &domainMsg:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return msgChan, nil
}
