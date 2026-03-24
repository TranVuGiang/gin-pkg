package notifylog

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/TranVuGiang/gin_pkg/notifylog/notifier"
	"github.com/rs/zerolog"
)

type Encoding int64

const (
	// Console outputs human-readable colored text via zerolog.ConsoleWriter.
	Console Encoding = iota
	// JSON outputs structured JSON logs to stderr.
	JSON
)

// Deprecated: Use Console instead.
const CBOR = Console

var (
	globalLevelOnce sync.Once
	globalLevel     zerolog.Level
)

type NotifyLog struct {
	zerolog.Logger
}

func New(name string, encoding Encoding, notifiers ...notifier.Notifier) NotifyLog {
	globalLevelOnce.Do(func() {
		globalLevel = parseLogLevel(os.Getenv("LOG_LEVEL"))
		zerolog.SetGlobalLevel(globalLevel)
	})

	var logger zerolog.Logger

	if encoding == Console {
		logger = console(name)
	} else {
		logger = json(name)
	}

	for _, not := range notifiers {
		logger = logger.Hook(not)
	}

	return NotifyLog{
		Logger: logger,
	}
}

func (l NotifyLog) With(fields map[string]string) NotifyLog {
	ctx := l.Logger.With()

	for k, v := range fields {
		ctx = ctx.Str(k, v)
	}

	return NotifyLog{Logger: ctx.Logger()}
}

func console(name string) zerolog.Logger {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339Nano,
	}

	return zerolog.New(output).With().Caller().Stack().Timestamp().Str("logger", name).Logger()
}

func json(name string) zerolog.Logger {
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }
	zerolog.TimeFieldFormat = time.RFC3339Nano

	return zerolog.New(os.Stderr).With().Caller().Stack().Timestamp().Str("logger", name).Logger()
}

func parseLogLevel(s string) zerolog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.DebugLevel
	}
}
