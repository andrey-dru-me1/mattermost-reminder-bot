package services

import (
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/repositories"
)

func GetChannels(app *app.Application) ([]models.Channel, error) {
	return repositories.GetChannels(app.Db)
}

func GetChannel(app *app.Application, name string) (*models.Channel, error) {
	return repositories.GetChannel(app.Db, name)
}

func InsertChannel(app *app.Application, channel models.Channel) error {
	return repositories.InsertChannel(app.Db, channel)
}

func DeleteChannel(app *app.Application, name string) error {
	return repositories.DeleteChannel(app.Db, name)
}
