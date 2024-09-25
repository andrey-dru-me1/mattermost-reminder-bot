package dtos

type ReminderDTO struct {
	Name    string `json:"name"`
	Owner   string `json:"owner"`
	Rule    string `json:"rule"`
	Channel string `json:"channel"`
	Message string `json:"message"`
}

type UserDTO struct {
	Name    string `json:"name"`
	Webhook string `json:"webhook"`
}

type MMRequest struct {
	ChannelName string `form:"channel_name"`
	UserName    string `form:"user_name"`
	Command     string `form:"command"`
	Text        string `form:"text"`
}
