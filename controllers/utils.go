package controllers

import (
	"fmt"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/app"
	"github.com/gin-gonic/gin"
)

func extractApp(c *gin.Context) (*app.Application, error) {
	app, exists := c.MustGet("app").(*app.Application)
	if !exists {
		return nil, fmt.Errorf("application not provided")
	}
	return app, nil
}
