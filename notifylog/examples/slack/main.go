package main

import (
	"os"

	"github.com/TranVuGiang/gin-pkg/notifylog"
	"github.com/TranVuGiang/gin-pkg/notifylog/notifier"
	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
)

func main() {
	slack := notifier.NewSlackNotifier(
		zerolog.InfoLevel,
		os.Getenv("SLACK_CHANNEL"),
		slack.New(os.Getenv("SLACK_TOKEN")),
	)
	log := notifylog.New("test", notifylog.JSON, slack)

	log.Info().Str("foo", "bar").Msg("Hello world ddd")
}
