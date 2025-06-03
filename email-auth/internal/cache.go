package internal

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	SetValue(ctx context.Context, key string, value []byte, exp time.Duration) error
	GetValue(ctx context.Context, key string, value interface{}) error
	ValueExists(ctx context.Context, key string) (*string, error)
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCatch(client *redis.Client) Cache {
	return &RedisCache{
		client: client,
	}
}

func (r *RedisCache) SetValue(ctx context.Context, key string, value []byte, exp time.Duration) error {
	_, err := r.client.Set(ctx, key, value, exp).Result()
	if err != nil {
		return err
	}

	log.Println("cache set")
	return nil
}

func (r *RedisCache) GetValue(ctx context.Context, key string, value interface{}) error {
	result, err := r.ValueExists(ctx, key)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(*result), value); err != nil {
		return err
	}

	log.Println("cache get")
	return nil
}

func (r *RedisCache) ValueExists(ctx context.Context, key string) (*string, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return &result, nil
}
