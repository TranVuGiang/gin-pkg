# gin_pkg

A Go library of reusable building blocks for [Gin](https://github.com/gin-gonic/gin)-based microservices. It bundles an HTTP server, metrics server, graceful-shutdown runner, middleware collection, structured logging with notifications, database/cache clients, and message streaming helpers.

## Packages

| Package | Description |
|---|---|
| `httpserver` | Gin HTTP server with built-in middleware and typed handler helpers |
| `metricserver` | Prometheus `/metrics` + `/status` server |
| `runner` | Concurrent service runner with graceful shutdown |
| `middleware` | Request ID, logger, error handler, metrics, JWKS JWT auth |
| `notifylog` | Zerolog-based logger with Slack/Telegram alert hooks |
| `postgres` | pgx connection pool that implements `runner.Service` |
| `goredis` | go-redis client that implements `runner.Service` |
| `redispub` | Redis Streams publisher (Watermill) |
| `redissub` | Redis Streams subscriber (Watermill) |
| `kctoken` | Keycloak service-account token client |
| `keyprovider` | JWKS key-function provider for JWT verification |
| `pagination` | Pagination helpers (normalize, compute totals) |
| `validator` | Singleton `go-playground/validator` wrapper |
| `testutil` | Test helpers for Postgres, Redis, and Gin |

---

## Quick start

```go
import (
    "github.com/TranVuGiang/gin_pkg/httpserver"
    "github.com/TranVuGiang/gin_pkg/metricserver"
    "github.com/TranVuGiang/gin_pkg/runner"
)

func main() {
    api := httpserver.New(&httpserver.Config{
        Host:        "0.0.0.0",
        Port:        8080,
        EnableCors:  true,
        CorsOrigins: []string{"https://example.com"},
    })

    api.Root.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    metrics := metricserver.New(&metricserver.Config{
        Host: "0.0.0.0",
        Port: 9090,
    })

    runner.New(
        runner.WithCoreService(api),
        runner.WithInfrastructureService(metrics),
    ).Run()
}
```

---

## runner

Starts every registered service in a goroutine and orchestrates graceful shutdown on `SIGINT`/`SIGTERM`. Infrastructure services are stopped before core services.

```go
r := runner.New(
    runner.WithCoreService(apiServer),
    runner.WithInfrastructureService(db),
    runner.WithInfrastructureService(redisClient),
    runner.WithGracefulShutdownTimeout(15 * time.Second),
)
r.Run()
```

Any type that satisfies `runner.Service` can be registered:

```go
type Service interface {
    Run()
    Stop(ctx context.Context) error
    Name() string
}
```

---

## httpserver

Gin server with the following middleware pre-installed: error handler, body-size limit, Prometheus metrics, and request logger.

```go
srv := httpserver.New(&httpserver.Config{
    Host:         "0.0.0.0",
    Port:         8080,
    BodyLimit:    "8M",   // supports K, M, G suffixes
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 30 * time.Second,
    GracePeriod:  10 * time.Second,
})
```

### Typed handlers with `ExecuteStandardized`

```go
type CreateUserRequest struct {
    Name  string `json:"name"  validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

type CreateUserResponse struct {
    ID string `json:"id"`
}

func createUser(
    log notifylog.NotifyLog,
    c *gin.Context,
    req *CreateUserRequest,
) (*httpserver.HandlerResponse[CreateUserResponse], *httpserver.AppError) {
    // ... business logic
    return &httpserver.HandlerResponse[CreateUserResponse]{
        Status: http.StatusCreated,
        Data:   CreateUserResponse{ID: "abc-123"},
    }, nil
}

srv.Root.POST("/users", func(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.Error(err)
        return
    }
    data, appErr := httpserver.ExecuteStandardized(c, &req, "createUser", createUser)
    if appErr != nil {
        c.Error(appErr)
        return
    }
    c.JSON(http.StatusCreated, data)
})
```

Response envelope:

```json
{
  "requestId": "3bf74527-...",
  "data": { "id": "abc-123" },
  "pagination": null
}
```

---

## middleware

### Request ID

Reads `X-Request-ID` from the incoming request (or generates one) and stores it in the Gin context.

```go
engine.Use(middleware.RequestID())
```

### JWKS JWT Authentication

Validates `Authorization: Bearer <token>` against a remote JWKS endpoint and stores parsed claims in context.

```go
auth := middleware.NewJwksAuth(logger, keyfuncProvider)
protected := srv.Root.Group("/api")
protected.Use(auth.Middleware())
```

Retrieve claims downstream:

```go
claims, err := middleware.GetExtendedClaimsFromContext(c)
```

### Context helpers

```go
middleware.GetRequestID(c)      // (string, bool)
middleware.GetTenantID(c)       // string
middleware.GetAPIKey(c)         // string
middleware.GetRateLimitRequests(c) // int
```

---

## notifylog

Structured logger built on [zerolog](https://github.com/rs/zerolog). Supports console (human-readable) and JSON output. Attach Slack or Telegram notifiers to receive alerts on error/fatal events.

```go
// JSON logger
log := notifylog.New("my-service", notifylog.JSON)

// Console logger with Slack alerts on error+
slackNotifier, _ := notifier.NewSlack(notifier.SlackConfig{
    WebhookURL: "https://hooks.slack.com/...",
    MinLevel:   zerolog.ErrorLevel,
})
log := notifylog.New("my-service", notifylog.Console, slackNotifier)

// Attach structured fields
log = log.With(map[string]string{"tenant_id": "abc", "trace_id": "xyz"})

log.Info().Msg("service started")
log.Error().Err(err).Msg("something went wrong") // triggers Slack alert
```

Log level is controlled by the `LOG_LEVEL` environment variable (`debug`, `info`, `warn`, `error`, `fatal`, `panic`). Defaults to `debug`.

---

## postgres

pgx connection pool that satisfies `runner.Service`.

```go
db, err := postgres.New(&postgres.Config{
    URL:                   "postgres://user:pass@localhost:5432/mydb",
    MaxConnection:         10,
    MinConnection:         2,
    MaxConnectionIdleTime: 5 * time.Minute,
    LogLevel:              tracelog.LogLevelInfo,
})

runner.New(runner.WithInfrastructureService(db)).Run()
```

---

## goredis

go-redis client that satisfies `runner.Service`.

```go
rdb, err := goredis.New(&goredis.Config{
    Host:         "localhost",
    Port:         6379,
    Password:     "",
    DB:           0,
    PingTimeout:  3 * time.Second,
    PoolSize:     10,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
})
```

---

## redispub / redissub

Publish and subscribe to Redis Streams using [Watermill](https://watermill.io/).

```go
// Publisher
pub, err := redispub.New(rdb.Client, redispub.Options{MaxStreamEntries: 1000})
pub.PublishToTopic("orders", `{"id":"1"}`, `{"id":"2"}`)

// Subscriber
sub, err := redissub.NewSubscriber(
    rdb.Client,
    "order-processor-group",
    "orders",
    func(ctx context.Context, payload message.Payload) error {
        fmt.Println(string(payload))
        return nil
    },
)
go sub.Start()
```

---

## kctoken

Fetches a Keycloak service-account access token using the client-credentials flow.

```go
client := kctoken.NewTokenClient(
    "https://keycloak.example.com/realms/myrealm/protocol/openid-connect/token",
    "my-client-id",
    "my-client-secret",
    kctoken.WithTimeout(5 * time.Second),
)

token, err := client.GetServiceToken()
```

---

## pagination

```go
page, pageSize, offset := pagination.Normalize(rawPage, rawPageSize)
totalPages := pagination.ComputeTotals(totalCount, pageSize)
```

Defaults: page = 1, pageSize = 100, max pageSize = 100.

---

## validator

```go
v := validator.DefaultRestValidator()
if err := v.Validate(myStruct); err != nil {
    // validation error
}
```

---

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `LOG_LEVEL` | `debug` | Global log level (`debug`, `info`, `warn`, `error`, `fatal`, `panic`) |
