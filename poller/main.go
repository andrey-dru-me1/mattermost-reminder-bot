package main

import (
	"context"
	"os"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	ticker := setupTicker()

	ctx, cancelCtx := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	cancelShutdownHandler := setupGracefulShutdown(cancelCtx, ticker)
	defer func() {
		select {
		case cancelShutdownHandler <- true:
		default:
		}
	}()

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			log.Info().Msg("All goroutines finished, exiting")
			return
		case <-ticker.C:
			if err := processReminds(ctx, wg); err != nil {
				log.Err(err).Msg("Error processing reminds")
			}
		}
	}
}
