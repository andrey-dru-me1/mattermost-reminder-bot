package dtos

type ReminderDTO struct {
	Name    string `json:"name"`
	Rule    string `json:"rule"`
	Channel string `json:"channel"`
	Message string `json:"message"`
}

type MMRequest struct {
	ChannelName string `form:"channel_name"`
	Command     string `form:"command"`
	Text        string `form:"text"`
}
