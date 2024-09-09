package controllers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/dtos"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/services"
	"github.com/gin-gonic/gin"
)

const usage = `Usage: /reminder COMMAND OPTIONS
  commands:
  - add, create NAME CRON-RULE - creates new reminder
  - list, ls - lists all reminders
  - delete, del, remove, rm ID - deletes a reminder with ID identifier

CRON-RULE: "Seconds Minutes Hours DayOfMonth Month DayOfWeek Year"
`

type mattermostRequest struct {
	ChannelName string `form:"channel_name"`
	Command     string `form:"command"`
	Text        string `form:"text"`
}

func mattermostReminderCreate(c *gin.Context, app *app.Application, req mattermostRequest, tokens []string) {
	if len(tokens) < 3 {
		c.JSON(http.StatusOK, gin.H{"text": `Usage: '/reminder create NAME "CRON-RULE"'`})
		return
	}

	rem := dtos.ReminderDTO{
		Name:    tokens[1],
		Rule:    tokens[2],
		Channel: req.ChannelName,
	}
	_, err := services.CreateReminder(app, rem)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{"text": fmt.Sprintf("Error: %s", err)},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{"text": "Reminder successfully created"},
	)
}

func mattermostReminderList(c *gin.Context, app *app.Application) {
	reminders, err := services.GetReminders(app)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{"text": err.Error()},
		)
		return
	}

	var sb strings.Builder
	sb.WriteString("|Id|Name|Channel|Rule|\n|-|-|-|-|\n")
	for _, reminder := range reminders {
		sb.WriteString(
			fmt.Sprintf(
				"|%d|%s|%s|%s|\n",
				reminder.ID,
				reminder.Name,
				reminder.Channel,
				reminder.Rule,
			),
		)
	}

	c.JSON(
		http.StatusOK,
		gin.H{"text": sb.String()},
	)
}

func mattermostReminderDelete(c *gin.Context, app *app.Application, tokens []string) {
	reminderIDString := tokens[1]
	reminderID, err := strconv.ParseInt(reminderIDString, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"text": `Usage: '/reminder delete ID'`})
		return
	}

	if err := services.DeleteReminder(app, reminderID); err != nil {
		c.JSON(http.StatusOK, gin.H{"text": fmt.Sprintf("Id %d was not found", reminderID)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"text": "Reminder successfully deleted"})
}

func MattermostReminder(c *gin.Context) {
	var req mattermostRequest
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

	tokens := tokenize(req.Text)

	if len(tokens) > 0 && strings.EqualFold(req.Command, "/reminder") {
		switch tokens[0] {
		case "add", "create":
			mattermostReminderCreate(c, app, req, tokens)
			return
		case "list", "ls":
			mattermostReminderList(c, app)
			return
		case "delete", "del", "remove", "rm":
			mattermostReminderDelete(c, app, tokens)
			return
		}
	}
	c.JSON(
		http.StatusOK,
		gin.H{"text": usage},
	)
}

func tokenize(str string) []string {
	re := regexp.MustCompile(`'[^']*'|"[^"]*"|\S+`)
	tokens := re.FindAllString(str, -1)

	for i, token := range tokens {
		tokLen := len(token)
		if tokLen > 1 &&
			((token[0] == '"' && token[tokLen-1] == '"') ||
				(token[0] == '\'' && token[tokLen-1] == '\'')) {
			tokens[i] = token[1 : tokLen-1]
		}
	}
	return tokens
}
