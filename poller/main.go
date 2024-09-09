package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
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
	jsonStr := []byte(fmt.Sprintf(
		`{"channel": "%s", "text": "%s"}`,
		reminder.Channel,
		reminder.Message,
	))
	resp, err := http.Post(
		"http://test_mm:8065/hooks/993ksxbh1jrhiqsnqe4ed6w1ay",
		"application/json",
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(resp)
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

	for _, reminder := range reminders {
		go printRemind(reminder)
	}

	return nil
}

func main() {
	ticker := time.NewTicker(1 * time.Minute)

	for {
		<-ticker.C
		if err := processReminders(); err != nil {
			log.Println(err)
		}
	}
}
