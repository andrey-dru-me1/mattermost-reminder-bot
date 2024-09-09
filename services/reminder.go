package services

import (
	"github.com/andrey-dru-me1/mattermost-reminder-bot/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/dtos"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/models"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/repositories"
)

func CreateReminder(app *app.Application, reminderDTO dtos.ReminderDTO) (int64, error) {
	id, err := repositories.CreateReminder(app.Db, reminderDTO)
	if err != nil {
		return 0, err
	}

	reminder, err := repositories.GetReminder(app.Db, id)
	if err != nil {
		return 0, err
	}

	app.NewReminders <- *reminder

	return id, nil
}

func UpdateReminder(app *app.Application, reminderID int64, reminderDTO dtos.ReminderDTO) error {
	return repositories.UpdateReminder(app.Db, reminderID, reminderDTO)
}

func DeleteReminder(app *app.Application, reminderID int64) error {
	return repositories.DeleteReminder(app.Db, reminderID)
}

func GetReminder(app *app.Application, reminderID int64) (*models.Reminder, error) {
	return repositories.GetReminder(app.Db, reminderID)
}

func GetReminders(app *app.Application) ([]models.Reminder, error) {
	return repositories.GetReminders(app.Db)
}
