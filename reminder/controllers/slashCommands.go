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
Commands:
- add, create NAME CRON-RULE MESSAGE - creates new reminder
- list, ls - lists all reminders
- delete, del, remove, rm ID - deletes a reminder with ID identifier
- timezone, tz LOCATION - updates channel timezone
- timezone, tz - shows current location

CRON-RULE:
- "Seconds Minutes Hours DayOfMonth Month DayOfWeek Year"
- "Minutes Hours DayOfMonth Month DayOfWeek Year" (Seconds default to 0)
- "Minutes Hours DayOfMonth Month DayOfWeek" (Year defaults to *)
- Month: 1-12 or JAN-DEC
- DayOfWeek 0-6 or SUN-SAT
- ` + "`*`" + ` - any value ("0 12 * * *" - 12:00 every day every month every year)
- ` + "`/`" + ` - time period ("*/5 * * * *" - every 5 minute of every day every month every year)
- ` + "`,`" + ` - list separator ("0 12 10,25 * *" - 12:00 every 10th and 25th day of every month)
- ` + "`-`" + ` - range ("0 12 * MON-FRI *" - 12:00 every workday)
- ` + "`L`" + ` - last ("0 12 * 5L *" - 12:00 last friday every month)
- ` + "`#`" + ` - numbered ("0 12 * TUE#2 *" - 12:00 second tuesday of every month)
LOCATION: TZ identifier (for example "Asia/Novosibirsk")
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

func mattermostReminderTimeZoneSet(c *gin.Context, app *app.Application, req mattermostRequest, tokens []string) {
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

func mattermostReminderTimeZoneGet(c *gin.Context, app *app.Application, req mattermostRequest) {
	channel, err := services.GetChannel(app, req.ChannelName)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{
				"text": fmt.Sprintf(
					"Time zone is not set for the channel '%s'. Used default time zone: UTC.\n",
					req.ChannelName,
				),
			},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		gin.H{"text": fmt.Sprintf("Time zone: %s", channel.TimeZone)},
	)
}

func mattermostReminderTimeZone(c *gin.Context, app *app.Application, req mattermostRequest, tokens []string) {
	if len(tokens) <= 1 {
		mattermostReminderTimeZoneGet(c, app, req)
	} else {
		mattermostReminderTimeZoneSet(c, app, req, tokens)
	}
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
