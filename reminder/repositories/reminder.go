package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/dtos"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
)

type multiScanner interface {
	Scan(dest ...any) error
}

func extractReminderFromRow(row multiScanner) (*models.Reminder, error) {
	var id int
	var name, rule, channel, message, createdAtString, modifiedAtString string

	if err := row.Scan(
		&id,
		&name,
		&rule,
		&channel,
		&message,
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
		Message:    message,
		CreatedAt:  createdAt,
		ModifiedAt: modifiedAt,
	}, nil
}

func GetReminder(db *sql.DB, reminderID int64) (*models.Reminder, error) {
	row := db.QueryRow(
		`SELECT id, name, rule, channel, message, created_at, modified_at FROM reminders WHERE id = ?`,
		reminderID,
	)

	reminder, err := extractReminderFromRow(row)
	if err != nil {
		return nil, err
	}

	return reminder, nil
}

func GetReminders(db *sql.DB) ([]models.Reminder, error) {
	rows, err := db.Query(`SELECT id, name, rule, channel, message, created_at, modified_at FROM reminders`)
	if err != nil {
		return nil, fmt.Errorf("execute query to extract data from reminders table: %w", err)
	}
	defer rows.Close()

	var reminders []models.Reminder

	for rows.Next() {
		reminder, err := extractReminderFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("extract reminder from row: %w", err)
		}

		reminders = append(reminders, *reminder)
	}

	return reminders, nil
}

func GetRemindersByChannel(db *sql.DB, channel string) ([]models.Reminder, error) {
	rows, err := db.Query(
		`SELECT id, name, rule, channel, message, created_at, modified_at FROM reminders WHERE channel = ?`,
		channel,
	)
	if err != nil {
		return nil, fmt.Errorf("execute query to extract data from reminders table: %w", err)
	}
	defer rows.Close()

	var reminders []models.Reminder

	for rows.Next() {
		reminder, err := extractReminderFromRow(rows)
		if err != nil {
			return nil, fmt.Errorf("extract reminder from row: %w", err)
		}

		reminders = append(reminders, *reminder)
	}

	return reminders, nil
}

func UpdateReminder(db *sql.DB, reminderID int64, req dtos.ReminderDTO) error {
	res, err := db.Exec(
		`UPDATE reminders SET name = ?, rule = ?, channel = ?, message = ? WHERE id = ?`,
		req.Name,
		req.Rule,
		req.Channel,
		req.Message,
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
		`INSERT INTO reminders (name, rule, channel, message) VALUES (?, ?, ?, ?)`,
		req.Name,
		req.Rule,
		req.Channel,
		req.Message,
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
	res, err := db.Exec(`--sql DELETE FROM reminders WHERE id = ?`, reminderID)
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
