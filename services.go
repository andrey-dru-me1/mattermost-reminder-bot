package main

import (
	"database/sql"
)

type reminderDTO struct {
	Name    string `json:"name"`
	Rule    string `json:"rule"`
	Channel string `json:"channel"`
}

func updateReminderService(db *sql.DB, reminderID string, req reminderDTO) error {
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

func createReminderService(db *sql.DB, req reminderDTO) (int64, error) {
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
