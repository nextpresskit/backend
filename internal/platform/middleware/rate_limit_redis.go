package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimitCounterStore interface {
	IncrementWindow(ctx context.Context, key string, window time.Duration) (count int64, ttl time.Duration, err error)
}

type RedisCounterStore struct {
	client *redis.Client
	prefix string
}

func NewRedisCounterStore(client *redis.Client, prefix string) *RedisCounterStore {
	return &RedisCounterStore{
		client: client,
		prefix: prefix,
	}
}

var redisFixedWindowScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("PTTL", KEYS[1])
return {current, ttl}
`)

func (s *RedisCounterStore) IncrementWindow(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	if s == nil || s.client == nil {
		return 0, 0, fmt.Errorf("redis store not initialized")
	}
	fullKey := key
	if s.prefix != "" {
		fullKey = s.prefix + ":" + key
	}
	res, err := redisFixedWindowScript.Run(ctx, s.client, []string{fullKey}, window.Milliseconds()).Result()
	if err != nil {
		return 0, 0, err
	}
	items, ok := res.([]any)
	if !ok || len(items) != 2 {
		return 0, 0, fmt.Errorf("unexpected redis script result")
	}
	count, ok := items[0].(int64)
	if !ok {
		return 0, 0, fmt.Errorf("unexpected redis count type")
	}
	ttlMs, ok := items[1].(int64)
	if !ok {
		return 0, 0, fmt.Errorf("unexpected redis ttl type")
	}
	ttl := time.Duration(ttlMs) * time.Millisecond
	if ttl < 0 {
		ttl = 0
	}
	return count, ttl, nil
}
