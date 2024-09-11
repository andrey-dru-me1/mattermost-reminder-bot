package main

import (
	"log"
	"net/http"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/controllers"
	"github.com/gin-gonic/gin"
)

func main() {
	app, err := app.SetupApplication()
	if err != nil {
		log.Fatal(err)
	}
	defer app.Db.Close()

	router := gin.Default()

	router.Use(func(ctx *gin.Context) {
		ctx.Set("app", app)
		ctx.Next()
	})

	router.GET("/healthcheck", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
	router.GET("/reminders", controllers.GetReminders)
	router.PUT("/reminder/:id", controllers.UpdateReminder)
	router.POST("/reminders", controllers.CreateReminder)
	router.DELETE("/reminder/:id", controllers.DeleteReminder)

	router.GET("/reminders/triggered", controllers.GetTriggeredReminders)
	router.POST("/reminders/triggered", controllers.CompleteReminds)

	router.POST("/mattermost/reminders", controllers.MattermostReminder)

	router.Run(":8080")
}
