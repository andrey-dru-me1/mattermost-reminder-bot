package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type mattermostRequest struct {
	ChannelID string `form:"channel_id"`
	Command   string `form:"command"`
	Text      string `form:"text"`
}

func mattermostReminderCreate(c *gin.Context) {
	var req mattermostRequest

	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens := strings.Split(req.Text, " ")

	if strings.EqualFold(req.Command, "/reminder") && strings.EqualFold(tokens[0], "create") {
		db, err := extractDB(c)
		if err != nil {
			c.JSON(
				http.StatusOK,
				gin.H{"text": fmt.Sprintf("Error: %s", err)},
			)
			return
		}

		rem := reminderDTO{Name: tokens[1], Rule: tokens[2], Channel: req.ChannelID}
		_, err = createReminderService(db, rem)
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
		return
	}
	c.Status(http.StatusOK)
}
