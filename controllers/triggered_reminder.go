package controllers

import (
	"log"
	"net/http"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/models"
	"github.com/gin-gonic/gin"
)

func GetTriggeredReminders(c *gin.Context) {
	app := c.MustGet("app").(*app.Application)
	triggeredReminders := app.TriggeredReminders

	var reminders []models.Reminder
	for {
		select {
		case triggered := <-triggeredReminders:
			log.Printf("Collect triggered reminder %d: %s\n", triggered.Reminder.ID, triggered.Reminder.Name)
			reminders = append(reminders, triggered.Reminder)
			triggered.Complete <- true
		default:
			c.JSON(http.StatusOK, reminders)
			return
		}
	}
}
