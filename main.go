package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	//e.Use(middleware.RateLimiterWithConfig(RateLimitConfig()))

	if err := e.Start(":8080"); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func RateLimitConfig() middleware.RateLimiterConfig {

	config := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		// Store:   limit.NewRedisLimitStore(),
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
