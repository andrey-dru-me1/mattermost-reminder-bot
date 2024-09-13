package controllers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/dtos"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/services"
	"github.com/gin-gonic/gin"
)

const usage = `Usage: /reminder COMMAND OPTIONS
  commands:
  - add, create NAME CRON-RULE MESSAGE - creates new reminder
  - list, ls - lists all reminders
  - delete, del, remove, rm ID - deletes a reminder with ID identifier
  - timezone, tz LOCATION - updates channel timezone

CRON-RULE: "Seconds Minutes Hours DayOfMonth Month DayOfWeek Year"
`

type mattermostRequest struct {
	ChannelName string `form:"channel_name"`
	Command     string `form:"command"`
	Text        string `form:"text"`
}

func mattermostReminderCreate(c *gin.Context, app *app.Application, req mattermostRequest, tokens []string) {
	if len(tokens) < 4 {
		c.JSON(http.StatusOK, gin.H{"text": usage})
		return
	}

	rem := dtos.ReminderDTO{
		Name:    tokens[1],
		Rule:    tokens[2],
		Message: tokens[3],
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

func mattermostReminderList(c *gin.Context, app *app.Application, req mattermostRequest) {
	reminders, err := services.GetRemindersByChannel(app, req.ChannelName)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{"text": err.Error()},
		)
		return
	}

	if len(reminders) > 0 {
		var sb strings.Builder
		sb.WriteString("|Id|Name|Channel|Rule|Message|\n|-|-|-|-|-|\n")
		for _, reminder := range reminders {
			sb.WriteString(
				fmt.Sprintf(
					"|%d|%s|%s|%s|%s|\n",
					reminder.ID,
					reminder.Name,
					reminder.Channel,
					reminder.Rule,
					reminder.Message,
				),
			)
		}

		c.JSON(
			http.StatusOK,
			gin.H{"text": sb.String()},
		)
	} else {
		c.JSON(
			http.StatusOK,
			gin.H{"text": "There are no reminders yet in this channel! Add a new one using `/reminder add ...`"},
		)
	}
}

func mattermostReminderDelete(c *gin.Context, app *app.Application, tokens []string) {
	if len(tokens) < 2 {
		c.JSON(http.StatusOK, gin.H{"text": usage})
		return
	}
	reminderIDString := tokens[1]
	reminderID, err := strconv.ParseInt(reminderIDString, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"text": usage})
		return
	}

	if err := services.DeleteReminder(app, reminderID); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"text": fmt.Sprintf("Error removing reminder %d: %s", reminderID, err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"text": "Reminder successfully deleted"})
}

func mattermostReminderTimeZone(c *gin.Context, app *app.Application, req mattermostRequest, tokens []string) {
	if len(tokens) < 2 {
		c.JSON(http.StatusOK, gin.H{"text": usage})
		return
	}
	timeZone := tokens[1]
	if _, err := time.LoadLocation(timeZone); err != nil {
		c.JSON(http.StatusOK, gin.H{"text": fmt.Sprintf("Error parsing timezone: %s", err)})
		return
	}
	if err := services.InsertChannel(
		app,
		models.Channel{Name: req.ChannelName, TimeZone: timeZone},
	); err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{"text": fmt.Sprintf("Error inserting channel: %s", err)},
		)
		return
	}
	c.JSON(http.StatusOK, gin.H{"text": "Timezone set"})
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
			mattermostReminderList(c, app, req)
			return
		case "delete", "del", "remove", "rm":
			mattermostReminderDelete(c, app, tokens)
			return
		case "timezone", "tz":
			mattermostReminderTimeZone(c, app, req, tokens)
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
