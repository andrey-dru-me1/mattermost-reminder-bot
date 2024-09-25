package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

func sendRemindToMM(
	c context.Context,
	reminder reminder,
) (*http.Response, error) {
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
		return nil, fmt.Errorf("parse json from reminder: %w", err)
	}

	req, err := http.NewRequestWithContext(
		c,
		"POST",
		os.Getenv("MM_IN_HOOK"),
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create request to a mattermost webhook: %w",
			err,
		)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send message to a mattermost webhook: %w", err)
	}

	logger.Info().
		Bytes(reqBody, jsonStr).
		Interface(respStatus, resp.Status).
		Interface(respHeader, resp.Header).
		Msg("Remind sent to mattermost")

	return resp, nil
}

func markRemindCompleted(c context.Context, reminder reminder) error {
	logger := log.With().Interface("reminder", reminder).Logger()

	jsonStr := []byte(fmt.Sprintf(
		`[%d]`,
		reminder.ID,
	))

	req, err := http.NewRequestWithContext(
		c,
		"POST",
		"http://reminder:8080/reminders/triggered",
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		return fmt.Errorf("create request to a reminder service: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send message to a reminder service: %w", err)
	}
	defer resp.Body.Close()

	logger.Info().Bytes(reqBody, jsonStr).
		Interface(respStatus, resp.Status).
		Interface(respHeader, resp.Header).
		Msg("Remind marked completed")

	return nil
}
