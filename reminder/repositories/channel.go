package repositories

import (
	"database/sql"
	"fmt"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
)

func GetChannels(db *sql.DB) ([]models.Channel, error) {
	rows, err := db.Query(`SELECT name, time_zone FROM channels`)
	if err != nil {
		return nil, fmt.Errorf("get channels: execute query: %w", err)
	}
	defer rows.Close()

	var channels []models.Channel

	for rows.Next() {
		var channel models.Channel
		if err := rows.Scan(&channel.Name, &channel.TimeZone); err != nil {
			return nil, fmt.Errorf("get channels: scan row: %w", err)
		}

		channels = append(channels, channel)
	}

	return channels, nil
}

func GetChannel(db *sql.DB, name string) (*models.Channel, error) {
	row := db.QueryRow(
		`SELECT name, time_zone FROM channels WHERE name = ?`,
		name,
	)

	var channel models.Channel
	if err := row.Scan(&channel.Name, &channel.TimeZone); err != nil {
		return nil, fmt.Errorf("get channel: scan row: %w", err)
	}

	return &channel, nil
}

func InsertChannel(db *sql.DB, channel models.Channel) error {
	_, err := db.Exec(`
		INSERT INTO channels (name, time_zone)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE
			time_zone = VALUES(time_zone)
		`,
		channel.Name,
		channel.TimeZone,
	)
	if err != nil {
		return fmt.Errorf("insert channel: execute query: %w", err)
	}
	return nil
}

func DeleteChannel(db *sql.DB, name string) error {
	_, err := db.Exec(`DELETE FROM channels WHERE name = ?`, name)
	if err != nil {
		return fmt.Errorf("delete channel: execute query: %w", err)
	}
	return nil
}
