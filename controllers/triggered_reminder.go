package controllers

import (
	"log"
	"net/http"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/models"
	"github.com/gin-gonic/gin"
)

func GetTriggeredReminders(c *gin.Context) {
	triggeredReminders := c.MustGet("triggeredReminders").(chan models.Reminder)
	var reminders []models.Reminder
	for {
		select {
		case reminder := <-triggeredReminders:
			log.Printf("Got a triggered reminder: %v\n", reminder)
			reminders = append(reminders, reminder)
		default:
			log.Println("Reminder list complete")
			c.JSON(http.StatusOK, reminders)
			return
		}
	}
}
