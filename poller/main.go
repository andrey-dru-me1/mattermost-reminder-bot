package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
		os.Getenv("MM_IN_HOOK"),
		"application/json",
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Response from mattermost server:", resp)
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
			log.Println(err)
			return
		}

		log.Println("Response from complete reminds:", resp)
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
