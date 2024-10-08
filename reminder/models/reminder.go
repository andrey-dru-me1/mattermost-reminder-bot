package models

import (
	"database/sql"
	"time"
)

type Reminder struct {
	ID         int64          `json:"id"`
	Owner      sql.NullString `json:"owner"`
	Name       string         `json:"name"`
	Rule       string         `json:"rule"`
	Channel    string         `json:"channel"`
	Message    string         `json:"message"`
	CreatedAt  time.Time      `json:"created_at"`
	ModifiedAt time.Time      `json:"modified_at"`
}
