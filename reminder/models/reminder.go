package models

import "time"

type Reminder struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Rule       string    `json:"rule"`
	Channel    string    `json:"channel"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}
