package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"example.com/colleague/graph/controllers/dtos"
	"example.com/colleague/graph/services"
	"github.com/gin-gonic/gin"
)

type mattermostRequest struct {
	ChannelName string `form:"channel_name"`
	Command     string `form:"command"`
	Text        string `form:"text"`
}

func mattermostReminderCreate(c *gin.Context, req mattermostRequest, tokens []string) {
	if len(tokens) < 3 {
		c.JSON(http.StatusOK, gin.H{"text": `Usage: '/reminder create [NAME] "[RULE]"'`})
		return
	}

	db, err := extractDB(c)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{"text": err.Error()},
		)
		return
	}

	rem := dtos.ReminderDTO{
		Name:    tokens[1],
		Rule:    strings.Join(tokens[1:], " "),
		Channel: req.ChannelName,
	}
	_, err = services.CreateReminder(db, rem)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{"text": fmt.Sprintf("Error: %s", err)},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{"text": "Reminder successfully created"},
	)
}

func mattermostReminderList(c *gin.Context) {
	db, err := extractDB(c)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{"text": err.Error()},
		)
		return
	}

	reminders, err := services.GetReminders(db)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{"text": err.Error()},
		)
		return
	}

	var sb strings.Builder
	sb.WriteString("|Name|Channel|Rule|\n|-|-|-|\n")
	for _, reminder := range reminders {
		sb.WriteString(
			fmt.Sprintf("|%s|%s|%s|\n", reminder.Name, reminder.Channel, reminder.Rule),
		)
	}

	c.JSON(
		http.StatusOK,
		gin.H{"text": sb.String()},
	)
}

func MattermostReminder(c *gin.Context) {
	var req mattermostRequest

	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens := strings.Split(req.Text, " ")

	if len(tokens) < 1 || !strings.EqualFold(req.Command, "/reminder") {
		c.JSON(
			http.StatusOK,
			gin.H{"text": "Usage: '/reminder [create|list]'"},
		)
	} else if strings.EqualFold(tokens[0], "add") {
		mattermostReminderCreate(c, req, tokens)
	} else if strings.EqualFold(tokens[0], "list") {
		mattermostReminderList(c)
	}
}
