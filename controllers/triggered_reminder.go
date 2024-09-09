package controllers

import (
	"log"
	"net/http"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/models"
	"github.com/gin-gonic/gin"
)

func GetTriggeredReminders(c *gin.Context) {
	reminderChan := c.MustGet("reminderChan").(chan models.Reminder)
	var reminders []models.Reminder
	for {
		select {
		case reminder := <-reminderChan:
			log.Printf("Got a triggered reminder: %v\n", reminder)
			reminders = append(reminders, reminder)
		default:
			log.Println("Reminder list complete")
			c.JSON(http.StatusOK, reminders)
			return
		}
	}
}
