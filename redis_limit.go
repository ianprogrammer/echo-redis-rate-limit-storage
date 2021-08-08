// Package provide the redis rate limiter storage
package echo_redis_rate_limit

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

type RateLimiterRedisStore struct {
	mutex     sync.Mutex
	rate      rate.Limit
	expiresIn time.Duration
	db        *redis.Client
	ctx       context.Context
}
type RateLimiterRedisStoreConfig struct {
	Rate        rate.Limit    // Rate of requests allowed to pass as req/s
	ExpiresIn   time.Duration // ExpiresIn is the duration after that a rate limiter is cleaned up
	RedisClient *redis.Client
}

const (
	rate_limit_prefix_key = "rate_limit"
)

func (store *RateLimiterRedisStore) incr(ctx context.Context, identifier string) error {
	return store.db.IncrBy(ctx, rate_limit_prefix_key+identifier, 1).Err()
}

func (store *RateLimiterRedisStore) getVisitorsByIdentifier(ctx context.Context, identifier string) (int, bool, error) {
	val, err := store.db.Get(ctx, rate_limit_prefix_key+identifier).Result()

	if err == redis.Nil {
		return 0, false, nil
	} else if err != nil {
		return 0, false, err
	}

	result, err := strconv.Atoi(val)
	if err != nil {
		return 0, false, err
	}

	return result, true, nil
}

func (store *RateLimiterRedisStore) saveVisitor(ctx context.Context, key string, remaining_requests int, exp time.Duration) error {
	return store.db.Set(ctx, rate_limit_prefix_key+key, remaining_requests, exp).Err()
}

func (store *RateLimiterRedisStore) Allow(identifier string) (bool, error) {
	store.mutex.Lock()
	allow := false
	ctx := context.Background()

	if store.ctx != nil {
		ctx = store.ctx
	}

	limiter, exists, err := store.getVisitorsByIdentifier(ctx, identifier)

	if err != nil {
		return false, err
	}

	if !exists {
		err = store.saveVisitor(ctx, identifier, 0, store.expiresIn)
		if err != nil {
			return false, err
		}
	}

	if limiter <= int(store.rate) {
		err = store.incr(ctx, identifier)
		if err != nil {
			return false, err
		}
		allow = true
	}

	store.mutex.Unlock()
	return allow, nil
}

func NewRedisLimitStore(ctx context.Context, config RateLimiterRedisStoreConfig) (store *RateLimiterRedisStore) {
	return &RateLimiterRedisStore{
		rate:      config.Rate,
		expiresIn: config.ExpiresIn,
		db:        config.RedisClient,
		ctx:       ctx,
	}
}
