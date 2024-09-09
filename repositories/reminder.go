package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/dtos"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/models"
)

type multiScanner interface {
	Scan(dest ...any) error
}

func extractReminderFromRow(row multiScanner) (*models.Reminder, error) {
	var id int
	var name, rule, channel, createdAtString, modifiedAtString string

	if err := row.Scan(
		&id,
		&name,
		&rule,
		&channel,
		&createdAtString,
		&modifiedAtString,
	); err != nil {
		return nil, err
	}

	createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtString)
	if err != nil {
		return nil, err
	}
	modifiedAt, err := time.Parse("2006-01-02 15:04:05", modifiedAtString)
	if err != nil {
		return nil, err
	}

	return &models.Reminder{
		ID:         id,
		Name:       name,
		Rule:       rule,
		Channel:    channel,
		CreatedAt:  createdAt,
		ModifiedAt: modifiedAt,
	}, nil
}

func GetReminder(db *sql.DB, reminderID int64) (*models.Reminder, error) {
	row := db.QueryRow("SELECT id, name, rule, channel, created_at, modified_at FROM reminders")

	reminder, err := extractReminderFromRow(row)
	if err != nil {
		return nil, err
	}

	return reminder, nil
}

func GetReminders(db *sql.DB) ([]models.Reminder, error) {
	rows, err := db.Query("SELECT id, name, rule, channel, created_at, modified_at FROM reminders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []models.Reminder

	for rows.Next() {
		reminder, err := extractReminderFromRow(rows)
		if err != nil {
			return nil, err
		}

		reminders = append(reminders, *reminder)
	}

	return reminders, nil
}

func UpdateReminder(db *sql.DB, reminderID int64, req dtos.ReminderDTO) error {
	res, err := db.Exec(
		"UPDATE reminders SET name = ?, rule = ?, channel = ? WHERE id = ?",
		req.Name,
		req.Rule,
		req.Channel,
		reminderID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return err
	}

	return nil
}

func CreateReminder(db *sql.DB, req dtos.ReminderDTO) (int64, error) {
	res, err := db.Exec(
		"INSERT INTO reminders (name, rule, channel) VALUES (?, ?, ?)",
		req.Name,
		req.Rule,
		req.Channel,
	)
	if err != nil {
		return 0, err
	}

	lastInsertID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

func DeleteReminder(db *sql.DB, reminderID int64) error {
	res, err := db.Exec("DELETE FROM reminders WHERE id = ?", reminderID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("reminder not found")
	}

	return nil
}
