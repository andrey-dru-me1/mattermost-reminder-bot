package main

import (
	"log"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/controllers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

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

	router.GET("/reminders", controllers.GetReminders)
	router.PUT("/reminder/:id", controllers.UpdateReminder)
	router.POST("/reminders", controllers.CreateReminder)
	router.DELETE("/reminder/:id", controllers.DeleteReminder)

	router.GET("/reminders/triggered", controllers.GetTriggeredReminders)

	router.POST("/mattermost/reminders", controllers.MattermostReminder)

	router.Run(":8080")
}
