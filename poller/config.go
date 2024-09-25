package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

const reqBody = "request body"
const respHeader = "response header"
const respBody = "response body"
const respStatus = "status"

func setupTicker() *time.Ticker {
	var ticker *time.Ticker

	durationString := os.Getenv("POLL_PERIOD")
	if duration, err := time.ParseDuration(durationString); err != nil {
		log.Warn().
			Err(err).
			Str("duration string", durationString).
			Msg("Warning: could not parse env `POLL_PERIOD`, using default value: 1m")
		ticker = time.NewTicker(1 * time.Minute)
	} else {
		log.Info().
			Str("duration string", durationString).
			Str("duration", duration.String()).
			Msg("Duration successfully parsed")
		ticker = time.NewTicker(duration)
	}

	return ticker
}

func setupGracefulShutdown(
	cancelCtx context.CancelFunc,
	ticker *time.Ticker,
) chan<- bool {
	var cancel chan bool

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigs:
			log.Info().Msg("Received shutdown signal, terminating...")
			cancelCtx()
			ticker.Stop()
		case <-cancel:
		}
	}()

	return cancel
}
