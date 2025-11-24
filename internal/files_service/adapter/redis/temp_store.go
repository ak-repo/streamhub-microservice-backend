package redisstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ak-repo/stream-hub/internal/files_service/domain"
	"github.com/ak-repo/stream-hub/internal/files_service/port"
	"github.com/redis/go-redis/v9"
)

type TempStore struct {
	client     *redis.Client
	expiration time.Duration
}

func NewTempStore(client *redis.Client, ttl time.Duration) port.TempFileStore {
	return &TempStore{client: client, expiration: ttl}
}

func (s *TempStore) SaveTemp(ctx context.Context, f *domain.File) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, "temp:file:"+f.ID, b, s.expiration).Err()
}

func (s *TempStore) GetTemp(ctx context.Context, fileID string) (*domain.File, error) {
	val, err := s.client.Get(ctx, "temp:file:"+fileID).Result()
	if err != nil {
		return nil, err
	}
	var f domain.File
	if err := json.Unmarshal([]byte(val), &f); err != nil {
		return nil, err
	}
	return &f, nil
}

func (s *TempStore) DeleteTemp(ctx context.Context, fileID string) error {
	return s.client.Del(ctx, "temp:file:"+fileID).Err()
}
