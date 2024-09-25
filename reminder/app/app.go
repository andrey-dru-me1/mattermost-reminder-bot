package app

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/internal/rman"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/repositories"
	"github.com/go-sql-driver/mysql"
)

type TriggeredReminder struct {
	Reminder models.Reminder
	Complete chan bool
}

type Application struct {
	Db              *sql.DB
	RemindManager   rman.RemindManager
	DefaultLocation *time.Location
}

func SetupApplication() (*Application, error) {
	db, err := setupDatabase()
	if err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation(os.Getenv("DEFAULT_TZ"))
	if err != nil {
		loc = time.UTC
	}

	rman := rman.New(db, loc)
	err = setupRemindGenerator(db, rman)
	if err != nil {
		return nil, err
	}

	return &Application{
		Db:              db,
		RemindManager:   rman,
		DefaultLocation: loc,
	}, nil
}

func setupDatabase() (*sql.DB, error) {
	cfg := mysql.Config{
		User:   os.Getenv("MYSQL_USER"),
		Passwd: os.Getenv("MYSQL_PASSWORD"),
		Net:    "tcp",
		Addr: fmt.Sprintf(
			"%s:%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
		),
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

func setupRemindGenerator(db *sql.DB, rman rman.RemindManager) error {
	reminders, err := repositories.GetReminders(db)
	if err != nil {
		return err
	}

	go func() {
		rman.AddReminders(reminders...)
	}()

	return nil
}
