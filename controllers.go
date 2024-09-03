package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func updateReminder(c *gin.Context) {
	db, exists := c.MustGet("db").(*sql.DB)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not found"})
		return
	}

	userID := c.Param("id")

	var request struct {
		Name string `json:"name"`
		Rule string `json:"rule"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res, err := db.Exec(
		"UPDATE reminders SET name = ?, rule = ? WHERE id = ?",
		request.Name,
		request.Rule,
		userID,
	)
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

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func createReminder(c *gin.Context) {
	db, exists := c.MustGet("db").(*sql.DB)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not found"})
		return
	}

	var request struct {
		Name string `json:"name"`
		Rule string `json:"rule"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res, err := db.Exec(
		"INSERT INTO reminders (name, rule) VALUES (?, ?)",
		request.Name,
		request.Rule,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	lastInsertID, err := res.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reminder created successfully",
		"id":      lastInsertID,
	})
}

func getReminders(c *gin.Context) {
	db, exists := c.MustGet("db").(*sql.DB)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get database connection"})
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM reminders")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var reminders []reminder

	for rows.Next() {
		var id int
		var name, rule, createdAtString, modifiedAtString string

		if err := rows.Scan(
			&id,
			&name,
			&rule,
			&createdAtString,
			&modifiedAtString,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		modifiedAt, err := time.Parse("2006-01-02 15:04:05", modifiedAtString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		reminders = append(reminders, reminder{
			ID:         id,
			Name:       name,
			Rule:       rule,
			CreatedAt:  createdAt,
			ModifiedAt: modifiedAt,
		})
	}

	c.IndentedJSON(http.StatusOK, reminders)
}
