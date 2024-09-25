package services

import (
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
)

func GetReminds(app *app.Application) []models.Remind {
	return app.RemindManager.GetReminds()
}

func CompleteReminds(app *app.Application, ids []int64) {
	app.RemindManager.CompleteReminds(ids...)
}
