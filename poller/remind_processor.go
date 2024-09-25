package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
)

func handleRemind(c context.Context, wg *sync.WaitGroup, reminder remind) {
	defer wg.Done()
	logger := log.With().Interface("reminder", reminder).Logger()

	resp, err := sendRemindToMM(c, reminder)
	if err != nil {
		logger.Error().Err(err).Msg("Could not send remind to mattermost")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if err := markRemindCompleted(c, reminder); err != nil {
			logger.Error().Err(err).Msg("Could not mark remind completed")
		}
	} else {
		logger.Error().
			Interface(respStatus, resp.Status).
			Interface(respHeader, resp.Header).
			Msg("Error sending request to a reminder service to mark reminds completed")
	}
}

func processReminds(c context.Context, wg *sync.WaitGroup) error {
	resp, err := http.Get("http://reminder:8080/reminders/triggered")
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	var reminders []remind

	if err := json.Unmarshal(body, &reminders); err != nil {
		return err
	}
	log.Info().Interface("reminders", reminders).Msg("Got reminds")

	for _, reminder := range reminders {
		wg.Add(1)
		go handleRemind(c, wg, reminder)
	}

	return nil
}
