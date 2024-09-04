package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"example.com/colleague/graph/controllers"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := setupDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := gin.Default()

	router.Use(func(ctx *gin.Context) {
		ctx.Set("db", db)
		ctx.Next()
	})

	router.GET("/reminders", controllers.GetReminders)
	router.PUT("/reminders/:id", controllers.UpdateReminder)
	router.POST("/reminders", controllers.CreateReminder)
	router.DELETE("/reminders/:id", controllers.DeleteReminder)

	router.POST("/mattermost/reminders", controllers.MattermostReminder)

	router.Run(":8080")
}

func setupDatabase() (*sql.DB, error) {
	cfg := mysql.Config{
		User:   os.Getenv("MYSQL_USER"),
		Passwd: os.Getenv("MYSQL_PASSWORD"),
		Net:    "tcp",
		Addr:   fmt.Sprintf("%s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		DBName: os.Getenv("DB_NAME"),
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
