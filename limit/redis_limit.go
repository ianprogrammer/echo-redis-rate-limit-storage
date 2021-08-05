package limit

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

type RateLimiterRedisStore struct {
	mutex       sync.Mutex
	rate        rate.Limit
	burst       int
	expiresIn   time.Duration
	lastCleanup time.Time
	db          *redis.Client
}

type RateLimiterRedisStoreConfig struct {
	Rate        rate.Limit    // Rate of requests allowed to pass as req/s
	Burst       int           // Burst additionally allows a number of requests to pass when rate limit is reached
	ExpiresIn   time.Duration // ExpiresIn is the duration after that a rate limiter is cleaned up
	RedisClient *redis.Client
}

// var (
// 	// ErrRateLimitExceeded denotes an error raised when rate limit is exceeded
// 	ErrRateLimitExceeded = error(http.StatusTooManyRequests, "rate limit exceeded")
// 	// ErrExtractorError denotes an error raised when extractor function is unsuccessful
// 	ErrExtractorError = echo.NewHTTPError(http.StatusForbidden, "error while extracting identifier")
// )

type Visitor struct {
	*rate.Limiter
	lastSeen time.Time
}

func (store *RateLimiterRedisStore) getVisitorsByIdentifier(identifier string) (*Visitor, bool) {
	return nil, false
}

func (store *RateLimiterRedisStore) saveVisitor(ctx context.Context, key string, value interface{}) {
	marshalledValue, err := json.Marshal(value)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = store.db.HSet(ctx, "rate_limit", key, marshalledValue).Err()
	if err != nil {
		log.Fatal(err)
		//c.logger.Error(err)
	}
}

func (store *RateLimiterRedisStore) Allow(identifier string) (bool, error) {
	store.mutex.Lock()
	limiter, exists := store.getVisitorsByIdentifier(identifier)
	if !exists {
		limiter = new(Visitor)
		limiter.Limiter = rate.NewLimiter(store.rate, store.burst)
		store.saveVisitor(context.Background(), identifier, limiter)
	}
	limiter.lastSeen = now()
	if now().Sub(store.lastCleanup) > store.expiresIn {
		store.deleteVisitors()
	}
	store.mutex.Unlock()
	return limiter.AllowN(now(), 1), nil
}

func (store *RateLimiterRedisStore) deleteVisitors() error {
	return store.db.HDel(context.Background(), "rate_limit").Err()
}

func NewRedisLimitStore(config RateLimiterRedisStoreConfig) (store *RateLimiterRedisStore) {
	store = &RateLimiterRedisStore{}
	store.burst = config.Burst
	store.rate = config.Rate
	store.expiresIn = config.ExpiresIn
	store.db = config.RedisClient

	if config.Burst == 0 {
		store.burst = int(config.Rate)
	}
	store.lastCleanup = now()
	return

}

var now = time.Now
