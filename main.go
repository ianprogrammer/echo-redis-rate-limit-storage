package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ianprogrammer/echo-redis-rate-limit-storage/limit"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.RateLimiterWithConfig(RateLimitConfig()))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "happy path")
	})

	if err := e.Start(":8080"); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func RateLimitConfig() middleware.RateLimiterConfig {

	rdb := redis.NewClient(&redis.Options{
		Addr:     ":6379",
		Password: "secret",
	})
	rdConf := &limit.RateLimiterRedisStoreConfig{
		Rate:        10,
		Burst:       15,
		ExpiresIn:   10 * time.Second,
		RedisClient: rdb,
	}
	config := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store:   limit.NewRedisLimitStore(*rdConf),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return &echo.HTTPError{
				Code:     http.StatusForbidden,
				Message:  "Nãp",
				Internal: err,
			}
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return &echo.HTTPError{
				Code:     http.StatusTooManyRequests,
				Message:  "tooMany",
				Internal: err,
			}
		},
	}
	return config
}
