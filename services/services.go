package services

import (
	"database/sql"
	"time"

	"example.com/colleague/graph/controllers/dtos"
	"example.com/colleague/graph/models"
)

func GetReminders(db *sql.DB) ([]models.Reminder, error) {
	rows, err := db.Query("SELECT * FROM reminders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []models.Reminder

	for rows.Next() {
		var id int
		var name, rule, channel, createdAtString, modifiedAtString string

		if err := rows.Scan(
			&id,
			&name,
			&rule,
			&createdAtString,
			&modifiedAtString,
			&channel,
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

		reminders = append(reminders, models.Reminder{
			ID:         id,
			Name:       name,
			Rule:       rule,
			Channel:    channel,
			CreatedAt:  createdAt,
			ModifiedAt: modifiedAt,
		})
	}

	return reminders, nil
}

func UpdateReminder(db *sql.DB, reminderID string, req dtos.ReminderDTO) error {
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
