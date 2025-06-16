package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	_ "time/tzdata"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/dtos"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/services"
	"github.com/gin-gonic/gin"
	"github.com/google/shlex"
)

func mmReminderCreate(
	app *app.Application,
	req dtos.MMRequest,
	tokens []string,
) (string, error) {
	if err := services.MMReminderCreate(app, req, tokens); err != nil {
		return "", err
	}
	return "Reminder successfully created", nil
}

func mmReminderTimeZone(
	app *app.Application,
	req dtos.MMRequest,
	tokens []string,
) (string, error) {
	if len(tokens) <= 1 {
		return services.MMReminderTimeZoneGet(app, req), nil
	}
	str, err := services.MMReminderTimeZoneSet(app, req, tokens)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Time zone set to %s", str), nil
}

func authorize(c *gin.Context) error {
	err := fmt.Errorf("invalid token")

	authFull := c.Request.Header.Get("Authorization")
	tokens := strings.Split(authFull, " ")
	if len(tokens) != 2 {
		return err
	}

	auth := tokens[1]
	if !strings.EqualFold(auth, os.Getenv("MM_SC_TOKEN")) {
		return err
	}
	return nil
}

func processCommands(
	app *app.Application,
	req dtos.MMRequest,
	tokens []string,
) string {
	var str string
	var err error
	if len(tokens) > 0 && strings.EqualFold(req.Command, "/reminder") {
		switch tokens[0] {
		case "add", "create":
			str, err = mmReminderCreate(app, req, tokens)
		case "list", "ls":
			str, err = services.MMReminderList(app, req)
		case "delete", "del", "remove", "rm":
			str, err = services.MMReminderDelete(app, req, tokens)
		case "timezone", "tz":
			str, err = mmReminderTimeZone(app, req, tokens)
		case "wh", "webhook":
			str, err = services.MMReminderSetWebhook(app, req, tokens)
		case "own", "chown", "steal", "snatch":
			str, err = services.MMReminderChangeOwner(app, req, tokens)
		case "help", "h":
			str = help(tokens)
		}
	}
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	if str == "" {
		str = usage()
	}

	user, err := services.GetUser(app, req.UserName)
	if err != nil || !user.Webhook.Valid {
		str = "WARNING: Your webhook is not set! This might prevent your" +
			" reminds from being sent. Follow the `/reminder help webhook`" +
			" guide to create your webhook.\n\n" + str
	}
	return str
}

func MattermostReminder(c *gin.Context) {
	if err := authorize(c); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": fmt.Sprintf("Authorization error: %s", err)},
		)
		return
	}

	var req dtos.MMRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	app, err := extractApp(c)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{"text": err.Error()},
		)
		return
	}

	tokens, err := shlex.Split(req.Text)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": fmt.Sprintf("Error tokenizing request: %s", err)},
		)
	}

	c.JSON(http.StatusOK, gin.H{"text": processCommands(app, req, tokens)})
}
