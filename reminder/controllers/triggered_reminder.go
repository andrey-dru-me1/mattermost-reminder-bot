package controllers

import (
	"fmt"
	"net/http"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/services"
	"github.com/gin-gonic/gin"
)

func GetTriggeredReminders(c *gin.Context) {
	app := c.MustGet("app").(*app.Application)
	c.JSON(http.StatusOK, services.GetTriggeredReminders(app))
}

func CompleteReminds(c *gin.Context) {
	app := c.MustGet("app").(*app.Application)

	var ids []int64
	if err := c.BindJSON(&ids); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": fmt.Sprintf("Request should consist of an id array: %s", err.Error())},
		)
		return
	}

	services.CompleteReminds(app, ids)
	c.Status(http.StatusOK)
}
