package paymentredis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ak-repo/stream-hub/internal/payment_service/domain"
	"github.com/ak-repo/stream-hub/internal/payment_service/port"
	"github.com/redis/go-redis/v9"
)

type PaymentRedis struct {
	client *redis.Client
	ttl    time.Duration
}

func NewPaymentRedis(client *redis.Client, ttl time.Duration) port.Redis {
	return &PaymentRedis{client: client, ttl: ttl}
}

func (r *PaymentRedis) SavePaymentSession(
	ctx context.Context,
	session *domain.PaymentSession,
) error {

	key := "payment:order:" + session.RazorpayOrderID

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, r.ttl).Err()
}

func (r *PaymentRedis) GetSessionByOrderID(
	ctx context.Context,
	razorpayOrderID string,
) (*domain.PaymentSession, error) {

	key := "payment:order:" + razorpayOrderID

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("no payment session found")
		}
		return nil, err
	}

	var session domain.PaymentSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}
