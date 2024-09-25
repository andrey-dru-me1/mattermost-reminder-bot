package main

type remind struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Rule    string `json:"rule"`
	Channel string `json:"channel"`
	Message string `json:"message"`
	Webhook string `json:"webhook"`
}
