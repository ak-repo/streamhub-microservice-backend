package port

import (
	"context"

	"github.com/ak-repo/stream-hub/internal/channel_service/domain"
)

// PubSub defines real-time message broadcasting capabilities.
type PubSub interface {
	Publish(ctx context.Context, channelID string, msg *domain.Message) error
	Subscribe(ctx context.Context, channelID string) (<-chan *domain.Message, error)
}
