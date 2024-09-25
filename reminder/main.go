package main

import (
	"net/http"
	"os"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/controllers"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	app, err := app.SetupApplication()
	if err != nil {
		log.Err(err).Msg("Error setting application up")
	}
	defer app.Db.Close()

	router := gin.Default()

	router.Use(func(ctx *gin.Context) {
		ctx.Set("app", app)
		ctx.Next()
	})

	router.GET(
		"/healthcheck",
		func(ctx *gin.Context) { ctx.Status(http.StatusOK) },
	)
	router.GET("/reminders", controllers.GetReminders)
	router.POST("/reminders", controllers.CreateReminder)
	router.DELETE("/reminder/:id", controllers.DeleteReminder)

	router.GET("/reminders/triggered", controllers.GetTriggeredReminders)
	router.POST("/reminders/triggered", controllers.CompleteReminds)

	router.POST("/mattermost/reminders", controllers.MattermostReminder)

	router.Run()
}
