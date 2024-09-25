package services

import (
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/repositories"
)

func Getusers(app *app.Application) ([]models.User, error) {
	return repositories.GetUsers(app.Db)
}

func GetUser(app *app.Application, name string) (models.User, error) {
	return repositories.GetUser(app.Db, name)
}

func InsertUser(app *app.Application, user models.User) error {
	reminders, err := repositories.GetRemindersByUser(app.Db, user.Name)
	if err == nil && user.Webhook.Valid {
		for _, reminder := range reminders {
			app.RemindManager.UpdateRemindWebhook(
				reminder.ID,
				user.Webhook.String,
			)
		}
	}
	return repositories.InsertUser(app.Db, user)
}

func DeleteUser(app *app.Application, name string) error {
	return repositories.DeleteUser(app.Db, name)
}
