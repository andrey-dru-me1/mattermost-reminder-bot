package main

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
)

func extractDB(c *gin.Context) (*sql.DB, error) {
	db, exists := c.MustGet("db").(*sql.DB)
	if !exists {
		return nil, fmt.Errorf("database connection not found")
	}
	return db, nil
}
