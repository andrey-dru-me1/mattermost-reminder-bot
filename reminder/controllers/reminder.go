package controllers

import (
	"net/http"
	"strconv"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/dtos"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/services"
	"github.com/gin-gonic/gin"
)

func CreateReminder(c *gin.Context) {
	app, err := extractApp(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var request dtos.ReminderDTO

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := services.CreateReminder(app, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reminder created successfully",
		"id":      id,
	})
}

func GetReminders(c *gin.Context) {
	app, err := extractApp(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reminders, err := services.GetReminders(app)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, reminders)
}

func DeleteReminder(c *gin.Context) {
	app, err := extractApp(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reminderIDString := c.Param("id")
	reminderID, err := strconv.ParseInt(reminderIDString, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.DeleteReminder(app, reminderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reminder deleted successfully"})
}
