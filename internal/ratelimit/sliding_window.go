package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)


type RateLimitResult struct {
	Allowed bool
	Limit int
	Remaining int
	Reset int64
}

const slidingWindowLua = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local clearBefore = now - window

redis.call('ZREMRANGEBYSCORE', key, '-inf', clearBefore)
local currentRequests = redis.call('ZCARD', key)
local resetTime = math.floor((now + window) / 1000)

if currentRequests < limit then
			redis.call('ZADD', key, now, now)
			redis.call('PEXPIRE', key, window)
			local remaining = limit - currentRequests - 1
			return {1, remaining, resetTime}
else 
			return {0, 0, resetTime}
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

func (r *RedisSlidingWindowLimiter) Allow(ctx context.Context, identifier string) (*RateLimitResult, error) {
	key := "ratelimit:" + identifier
	now := time.Now().UnixMilli()
	windowMs := r.window.Milliseconds()

	result, err := r.script.Run(ctx, r.client, []string{key}, now, windowMs, r.limit).Slice()
	if err != nil {
		return nil, err
	}

	allowed := result[0].(int64) == 1
	remaining := int(result[1].(int64))
	resetTime := result[2].(int64)

	return &RateLimitResult{
		Allowed: allowed,
		Limit: r.limit,
		Remaining: remaining,
		Reset: resetTime,
	}, nil
}
