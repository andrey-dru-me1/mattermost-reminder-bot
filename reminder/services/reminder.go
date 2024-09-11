package services

import (
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/dtos"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/repositories"
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

	app.RemindManager.AddReminders(*reminder)

	return id, nil
}

func UpdateReminder(app *app.Application, reminderID int64, reminderDTO dtos.ReminderDTO) error {
	return repositories.UpdateReminder(app.Db, reminderID, reminderDTO)
}

func DeleteReminder(app *app.Application, reminderID int64) error {
	app.RemindManager.RemoveReminders(int(reminderID))
	return repositories.DeleteReminder(app.Db, reminderID)
}

func GetReminder(app *app.Application, reminderID int64) (*models.Reminder, error) {
	return repositories.GetReminder(app.Db, reminderID)
}

func GetRemindersByChannel(app *app.Application, channel string) ([]models.Reminder, error) {
	return repositories.GetRemindersByChannel(app.Db, channel)
}

func GetReminders(app *app.Application) ([]models.Reminder, error) {
	return repositories.GetReminders(app.Db)
}
