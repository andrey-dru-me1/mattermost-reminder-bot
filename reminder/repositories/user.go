package repositories

import (
	"database/sql"
	"fmt"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
)

func GetUsers(db *sql.DB) ([]models.User, error) {
	rows, err := db.Query(`SELECT name, webhook FROM users`)
	if err != nil {
		return nil, fmt.Errorf("get users: execute query: %w", err)
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.Name, &user.Webhook); err != nil {
			return nil, fmt.Errorf("get channels: scan row: %w", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func GetUser(db *sql.DB, name string) (models.User, error) {
	row := db.QueryRow(
		`SELECT name, webhook FROM users WHERE name = ?`,
		name,
	)

	var user models.User
	if err := row.Scan(&user.Name, &user.Webhook); err != nil {
		return user, fmt.Errorf("get user: scan row: %w", err)
	}

	return user, nil
}

func InsertUser(db *sql.DB, user models.User) error {
	_, err := db.Exec(`
		INSERT INTO users (name, webhook)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE
			webhook = VALUES(webhook)
	`,
		user.Name,
		user.Webhook,
	)
	if err != nil {
		return fmt.Errorf("insert user: execute query: %w", err)
	}
	return nil
}

func DeleteUser(db *sql.DB, name string) error {
	_, err := db.Exec(`DELETE FROM users WHERE name = ?`, name)
	if err != nil {
		return fmt.Errorf("delete user: execute query: %w", err)
	}
	return nil
}
