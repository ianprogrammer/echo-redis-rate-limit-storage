## Rate limiter redis storage for echo

This package provides a rate limiter storage using redis  for [echo web framework](https://echo.labstack.com/ "echo web framework").

## Installation

This package  requires a Go version with modules support, make sure you have initilized your go module.

``` 
go mod init github.com/user/project
```

After that install echo_redis_rate_limit

```
go get github.com/ianprogrammer/echo-redis-rate-limit-storage
```

## Example

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	echo_redis_rate_limit "github.com/ianprogrammer/echo-redis-rate-limit-storage"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.RateLimiterWithConfig(RateLimitConfig()))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK,"hello world")
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
	rdConf := &echo_redis_rate_limit.RateLimiterRedisStoreConfig{
		Rate:        10,
		ExpiresIn:   30 * time.Second,
		RedisClient: rdb,
	}
	config := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store:   echo_redis_rate_limit.NewRedisLimitStore(context.Background(),*rdConf),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return &echo.HTTPError{
				Code:     http.StatusForbidden,
				Message:  "error",
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

```

## Testing

You can use a benchmarking tool like [autocannon](https://www.npmjs.com/package/autocannon "autocannon").
If you don't want to use a benchmarking tool you can the run the following bash code.

```shell
for counter in $(seq 1 200); do  curl --write-out '%{http_code}\n' --silent --output /dev/null https://localhost:8081 --insecure; done
```

