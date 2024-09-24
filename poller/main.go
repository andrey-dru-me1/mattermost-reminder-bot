package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type reminder struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Rule       string    `json:"rule"`
	Channel    string    `json:"channel"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

func printRemind(reminder reminder) {
	const reqBody = "request body"

	logger := log.With().Interface("reminder", reminder).Logger()

	type remind struct {
		Channel string `json:"channel"`
		Message string `json:"text"`
	}

	rem := remind{
		Channel: reminder.Channel,
		Message: reminder.Message,
	}

	jsonStr, err := json.Marshal(rem)
	if err != nil {
		logger.Err(err).Msg("Error parsing json from reminder")
		return
	}

	resp, err := http.Post(
		os.Getenv("MM_IN_HOOK"),
		"application/json",
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		logger.Err(err).
			Bytes(reqBody, jsonStr).
			Msg("Error sending message to a mattermost webhook")
		return
	}

	logger.Info().
		Bytes(reqBody, jsonStr).
		Interface("response header", resp.Header).
		Interface("response body", resp.Body).
		Msg("Got a response from mattermost")

	if resp.StatusCode == http.StatusOK {
		jsonStr := []byte(fmt.Sprintf(
			`[%d]`,
			reminder.ID,
		))
		resp, err := http.Post(
			"http://reminder:8080/reminders/triggered",
			"application/json",
			bytes.NewBuffer(jsonStr),
		)
		if err != nil {
			logger.Err(err).
				Bytes(reqBody, jsonStr).
				Msg("Error marking remind completed")
			return
		}

		logger.Info().Bytes(reqBody, jsonStr).
			Interface("response header", resp.Header).
			Interface("response body", resp.Body).
			Msg("Got a response from reminder service")
	}
}

func processReminders() error {
	resp, err := http.Get("http://reminder:8080/reminders/triggered")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var reminders []reminder

	if err := json.Unmarshal(body, &reminders); err != nil {
		return err
	}
	log.Info().Interface("reminders", reminders).Msg("Got reminders")

	for _, reminder := range reminders {
		go printRemind(reminder)
	}

	return nil
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var ticker *time.Ticker

	durationString := os.Getenv("POLL_PERIOD")
	if duration, err := time.ParseDuration(durationString); err != nil {
		log.Err(err).
			Str("duration string", durationString).
			Msg("Error parsing duration from ENV")
		ticker = time.NewTicker(1 * time.Minute)
	} else {
		log.Info().
			Str("duration string", durationString).
			Str("duration", duration.String()).
			Msg("Duration successfully parsed")
		ticker = time.NewTicker(duration)
	}

	for {
		if err := processReminders(); err != nil {
			log.Err(err)
		}
		<-ticker.C
	}
}
