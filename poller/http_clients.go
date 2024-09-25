package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

func sendRemindToMM(
	c context.Context,
	remind remind,
) (*http.Response, error) {
	logger := log.With().Interface("reminder", remind).Logger()

	type message struct {
		Channel string `json:"channel"`
		Message string `json:"text"`
	}

	rem := message{
		Channel: remind.Channel,
		Message: remind.Message,
	}

	jsonStr, err := json.Marshal(rem)
	if err != nil {
		return nil, fmt.Errorf("parse json from reminder: %w", err)
	}

	req, err := http.NewRequestWithContext(
		c,
		"POST",
		"http://test_mm:8065/hooks/"+remind.Webhook,
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

func markRemindCompleted(c context.Context, reminder remind) error {
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
