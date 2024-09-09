package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/controllers"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/models"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/services"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/gorhill/cronexpr"
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

	ch := make(chan models.Reminder, 255)
	err = generateReminds(db, ch)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	router.Use(func(ctx *gin.Context) {
		ctx.Set("db", db)
		ctx.Set("reminderChan", ch)
		ctx.Next()
	})

	router.GET("/reminders", controllers.GetReminders)
	router.PUT("/reminder/:id", controllers.UpdateReminder)
	router.POST("/reminders", controllers.CreateReminder)
	router.DELETE("/reminder/:id", controllers.DeleteReminder)

	router.GET("/reminders/triggered", controllers.GetTriggeredReminders)

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

func generateReminds(db *sql.DB, ch chan models.Reminder) error {
	reminders, err := services.GetReminders(db)
	if err != nil {
		return err
	}

	for _, reminder := range reminders {
		go func() {
			expr, err := cronexpr.Parse(reminder.Rule)
			if err != nil {
				log.Printf(
					"Error while parsing cron expression '%s': %s\n", reminder.Rule, err,
				)
				return
			}
			log.Printf("Reminder detected: %v\n", reminder)

			for {
				nextTime := expr.Next(time.Now())
				timer := time.NewTimer(time.Until(nextTime))
				log.Printf("Next trigger time for reminder '%s' is: %v", reminder.Name, nextTime)
				<-timer.C
				go func() {
					ch <- reminder
				}()
			}
		}()
	}

	return nil
}
