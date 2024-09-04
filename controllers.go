package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func updateReminderController(c *gin.Context) {
	db, err := extractDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reminderID := c.Param("id")

	var request reminderDTO

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := updateReminderService(db, reminderID, request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func createReminderController(c *gin.Context) {
	db, err := extractDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var request reminderDTO

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := createReminderService(db, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reminder created successfully",
		"id":      id,
	})
}

func getReminders(c *gin.Context) {
	db, err := extractDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := db.Query("SELECT * FROM reminders")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var reminders []reminder

	for rows.Next() {
		var id int
		var name, rule, channel, createdAtString, modifiedAtString string

		if err := rows.Scan(
			&id,
			&name,
			&rule,
			&createdAtString,
			&modifiedAtString,
			&channel,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		modifiedAt, err := time.Parse("2006-01-02 15:04:05", modifiedAtString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		reminders = append(reminders, reminder{
			ID:         id,
			Name:       name,
			Rule:       rule,
			Channel:    channel,
			CreatedAt:  createdAt,
			ModifiedAt: modifiedAt,
		})
	}

	c.IndentedJSON(http.StatusOK, reminders)
}

func deleteReminder(c *gin.Context) {
	db, err := extractDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID := c.Param("id")

	res, err := db.Exec("DELETE FROM reminders WHERE id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Reminder not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reminder deleted successfully"})
}
