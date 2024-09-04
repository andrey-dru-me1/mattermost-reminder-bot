package dtos

type ReminderDTO struct {
	Name    string `json:"name"`
	Rule    string `json:"rule"`
	Channel string `json:"channel"`
}
