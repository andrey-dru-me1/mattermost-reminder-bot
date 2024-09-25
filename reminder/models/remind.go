package models

import "database/sql"

type Remind struct {
	ReminderId int64          `json:"id"`
	Owner      sql.NullString `json:"owner"`
	Name       string         `json:"name"`
	Rule       string         `json:"rule"`
	Channel    string         `json:"channel"`
	Message    string         `json:"message"`
	Webhook    string         `json:"webhook"`
}
