package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/dtos"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
)

const reminderCols = "id, owner, name, rule, channel, message, created_at, modified_at"

type multiScanner interface {
	Scan(dest ...any) error
}

func extractReminderFromRow(row multiScanner) (*models.Reminder, error) {
	var id int64
	var name, rule, channel, message, createdAtString, modifiedAtString string
	var owner sql.NullString

	if err := row.Scan(
		&id,
		&owner,
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
		Owner:      owner,
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
		"SELECT "+reminderCols+" FROM reminders WHERE id = ?",
		reminderID,
	)

	reminder, err := extractReminderFromRow(row)
	if err != nil {
		return nil, err
	}

	return reminder, nil
}

func GetReminders(db *sql.DB) ([]models.Reminder, error) {
	rows, err := db.Query(
		"SELECT " + reminderCols + " FROM reminders",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"execute query to extract data from reminders table: %w",
			err,
		)
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

func GetRemindersByChannel(
	db *sql.DB,
	channel string,
) ([]models.Reminder, error) {
	rows, err := db.Query(
		`SELECT `+reminderCols+` FROM reminders WHERE channel = ?`,
		channel,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"execute query to extract data from reminders table: %w",
			err,
		)
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

func UpdateReminderOwner(db *sql.DB, reminderID int64, userName string) error {
	res, err := db.Exec(
		`UPDATE reminders SET owner = ? WHERE id = ?`,
		userName,
		reminderID,
	)
	if err != nil {
		return fmt.Errorf("update reminder owner: execute query: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update reminder owner: get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("update reminder owner: no rows was affected")
	}

	return nil
}

func CreateReminder(db *sql.DB, req dtos.ReminderDTO) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO reminders (name, owner, rule, channel, message) VALUES (?, ?, ?, ?, ?)`,
		req.Name,
		req.Owner,
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
	res, err := db.Exec(`DELETE FROM reminders WHERE id = ?`, reminderID)
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
