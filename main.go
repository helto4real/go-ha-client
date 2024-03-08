package main

import (
	"os"
	"os/signal"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	client "github.com/helto4real/go-ha-client/client"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	conn, err := client.ConnectHomeAssistant("localhost", 8124, false, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIwNWYyOGFlZjkzYWU0MDUwODljMTYyYWNmNzZmNDk2MCIsImlhdCI6MTcwOTc4OTQ2OSwiZXhwIjoyMDI1MTQ5NDY5fQ.BqUndJESpV8jLTZO8aAT2uqDxmCb5cR0ZhJDewHQPAY")
	if err != nil {
		println(err)
		return
	}

	sub, err := conn.SubscribeEvents("state_changed", func(event *client.HaEvent) {
		log.Info().Msgf("Event: %v", event)
	})
	defer sub.Unsubscribe()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	select {
	case <-done:
		log.Info().Msg("Got interrupt signal")
		conn.Close()
	}
}
