package app

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/repositories"
	"github.com/go-sql-driver/mysql"
	"github.com/gorhill/cronexpr"
)

type TriggeredReminder struct {
	Reminder models.Reminder
	Complete chan bool
}

type Application struct {
	Db                 *sql.DB
	TriggeredReminders <-chan TriggeredReminder
	NewReminders       chan<- models.Reminder
}

func SetupApplication() (*Application, error) {
	db, err := setupDatabase()
	if err != nil {
		return nil, err
	}

	trigRems, newRems, err := setupRemindGenerator(db)
	if err != nil {
		return nil, err
	}

	return &Application{
		Db:                 db,
		TriggeredReminders: trigRems,
		NewReminders:       newRems,
	}, nil
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

func setupRemindGenerator(db *sql.DB) (<-chan TriggeredReminder, chan<- models.Reminder, error) {
	triggeredReminders := make(chan TriggeredReminder, 255)
	newReminders := make(chan models.Reminder)

	reminders, err := repositories.GetReminders(db)
	if err != nil {
		return nil, nil, err
	}
	go func() {
		for _, reminder := range reminders {
			log.Printf("Push reminder %d (%s) to a channel\n", reminder.ID, reminder.Name)
			newReminders <- reminder
		}
	}()

	go generateReminds(triggeredReminders, newReminders)

	return triggeredReminders, newReminders, nil
}

func generateReminds(triggeredReminders chan<- TriggeredReminder, newReminders <-chan models.Reminder) {
	log.Printf("Begin generating reminds")
	for reminder, more := <-newReminders; more; reminder, more = <-newReminders {
		go func() {
			expr, err := cronexpr.Parse(reminder.Rule)
			if err != nil {
				log.Printf(
					"Error while parsing cron expression '%s': %s\n", reminder.Rule, err,
				)
				return
			}
			log.Printf("Reminder %d (%s) starts generating reminds\n", reminder.ID, reminder.Name)

			for {
				nextTime := expr.Next(time.Now())
				timer := time.NewTimer(time.Until(nextTime))
				log.Printf("Next trigger time for reminder '%s' is: %v", reminder.Name, nextTime)
				<-timer.C
				complete := make(chan bool)
				go func() {
					triggeredReminders <- TriggeredReminder{
						Reminder: reminder,
						Complete: complete,
					}
				}()
				<-complete
			}
		}()
	}
	log.Printf("Finish generating reminds")
}
