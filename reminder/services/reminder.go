package services

import (
	"fmt"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/dtos"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/repositories"
	"github.com/gorhill/cronexpr"
)

func CreateReminder(
	app *app.Application,
	reminderDTO dtos.ReminderDTO,
) (int64, error) {
	if _, err := cronexpr.Parse(reminderDTO.Rule); err != nil {
		return 0, fmt.Errorf("parse cron expr: %w", err)
	}

	id, err := repositories.CreateReminder(app.Db, reminderDTO)
	if err != nil {
		return 0, fmt.Errorf("create reminder: %w", err)
	}

	reminder, err := repositories.GetReminder(app.Db, id)
	if err != nil {
		return 0, fmt.Errorf("get created reminder: %w", err)
	}

	app.RemindManager.AddReminders(*reminder)

	return id, nil
}

func UpdateReminderOwner(
	app *app.Application,
	reminderID int64,
	userName string,
) error {
	app.RemindManager.UpdateRemindOwner(reminderID, userName)
	return repositories.UpdateReminderOwner(app.Db, reminderID, userName)
}

func DeleteReminder(app *app.Application, reminderID int64) error {
	app.RemindManager.RemoveReminders(reminderID)
	return repositories.DeleteReminder(app.Db, reminderID)
}

func GetReminder(
	app *app.Application,
	reminderID int64,
) (*models.Reminder, error) {
	return repositories.GetReminder(app.Db, reminderID)
}

func GetRemindersByChannel(
	app *app.Application,
	channel string,
) ([]models.Reminder, error) {
	return repositories.GetRemindersByChannel(app.Db, channel)
}

func GetReminders(app *app.Application) ([]models.Reminder, error) {
	return repositories.GetReminders(app.Db)
}
