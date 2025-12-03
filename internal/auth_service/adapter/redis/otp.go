package otpredis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type OTPStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewOTPStore(client *redis.Client, ttl time.Duration) *OTPStore {
	return &OTPStore{client: client, ttl: ttl}
}

func (s *OTPStore) SaveOTP(ctx context.Context, email, otp string) error {
	key := "otp:" + email
	return s.client.Set(ctx, key, otp, s.ttl).Err()
}

func (s *OTPStore) VerifyOTP(ctx context.Context, email string) (string, error) {
	key := "otp:" + email
	return s.client.Get(ctx, key).Result()
}

func (s *OTPStore) DeleteOTP(ctx context.Context, email string) error {
	key := "otp:" + email
	return s.client.Del(ctx, key).Err()
}
