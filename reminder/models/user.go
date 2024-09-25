package models

import "database/sql"

type User struct {
	Name    string         `json:"name"`
	Webhook sql.NullString `json:"webhook"`
}
