package limit

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
}
type RateLimiterRedisStoreConfig struct {
	Rate        rate.Limit    // Rate of requests allowed to pass as req/s
	ExpiresIn   time.Duration // ExpiresIn is the duration after that a rate limiter is cleaned up
	RedisClient *redis.Client
}

const (
	rate_limit_prefix_key = "rate_limit"
)

func (store *RateLimiterRedisStore) Incr(identifier string) error {
	return store.db.IncrBy(context.Background(), rate_limit_prefix_key+identifier, 1).Err()
}

func (store *RateLimiterRedisStore) getVisitorsByIdentifier(identifier string) (int, bool, error) {
	val, err := store.db.Get(context.Background(), rate_limit_prefix_key+identifier).Result()

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
	limiter, exists, err := store.getVisitorsByIdentifier(identifier)

	if err != nil {
		return false, err
	}

	if !exists {
		err = store.saveVisitor(context.Background(), identifier, 0, store.expiresIn)
		if err != nil {
			return false, err
		}
	}

	if limiter <= int(store.rate) {
		err = store.Incr(identifier)
		if err != nil {
			return false, err
		}
		allow = true
	}

	store.mutex.Unlock()
	return allow, nil
}

func NewRedisLimitStore(config RateLimiterRedisStoreConfig) (store *RateLimiterRedisStore) {
	return &RateLimiterRedisStore{
		rate:      config.Rate,
		expiresIn: config.ExpiresIn,
		db:        config.RedisClient,
	}
}
