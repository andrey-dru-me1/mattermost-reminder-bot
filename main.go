package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// type dayOfWeek uint8
// type hour uint8
// type min uint8

// const (
// 	monday = iota
// 	tuesday
// 	wednesday
// 	thursday
// 	friday
// 	saturday
// 	sunday
// )

type reminder struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Rule       string    `json:"rule"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

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

	router.GET("/reminders", getReminders)
	router.PUT("/reminders/:id", updateReminder)
	router.POST("/reminders", createReminder)
	router.DELETE("/reminders/:id", deleteReminder)

	// router.Run("0.0.0.0:8080")
	router.Run("localhost:8080")
}

func setupDatabase() (*sql.DB, error) {

	dbPassword := os.Getenv("MYSQL_PASSWORD")

	dbUrl := fmt.Sprintf("reminders:%s@tcp(localhost:3306)/reminders", dbPassword)

	db, err := sql.Open("mysql", dbUrl)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
