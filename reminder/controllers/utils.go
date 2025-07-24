package controllers

import (
	"fmt"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/gin-gonic/gin"
)

func extractApp(c *gin.Context) (*app.Application, error) {
	val, exists := c.Get("app")
	if !exists {
		return nil, fmt.Errorf("application not provided")
	}
	app, ok := val.(*app.Application)
	if !ok {
		return nil, fmt.Errorf("application has wrong type")
	}
	return app, nil
}
