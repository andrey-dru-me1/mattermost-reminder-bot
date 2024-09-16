package services

import (
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
)

func GetTriggeredReminders(app *app.Application) []models.Reminder {
	return app.RemindManager.GetReminds()
}

func CompleteReminds(app *app.Application, ids []int64) {
	app.RemindManager.CompleteReminds(ids...)
}
