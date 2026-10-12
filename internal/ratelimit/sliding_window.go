package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const slidingWindowLua = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local clearBefore = now - window

redis.call('ZREMRANGEBYSCORE', key, '-inf', clearBefore)
local currentRequests = redis.call('ZCARD', key)

if currentRequests < limit then
			redis.call('ZADD', key, now, now)
			redis.call('PEXPIRE', key, window)
			return 1
else 
			return 0
end
`

type RedisSlidingWindowLimiter struct {
	client *redis.Client
	script *redis.Script
	window time.Duration
	limit int
}

func NewRedisSlidingWindowLimiter(client *redis.Client, window time.Duration, limit int) 	*RedisSlidingWindowLimiter {
		return &RedisSlidingWindowLimiter {
			client: client,
			script: redis.NewScript(slidingWindowLua),
			window: window,
			limit: limit,
		}
}

func (r *RedisSlidingWindowLimiter) Allow(ctx context.Context, identifier string) (bool, error) {
	key := "ratelimit:" + identifier
	now := time.Now().UnixMilli()
	windowMs := r.window.Milliseconds()

	result, err := r.script.Run(ctx, r.client, []string{key}, now, windowMs, r.limit).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}
