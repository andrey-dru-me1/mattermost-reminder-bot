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

const usage = `Usage: /reminder COMMAND OPTIONS
Commands:
- add, create NAME CRON_RULE MESSAGE - creates new reminder
- list, ls - lists all reminders
- delete, del, remove, rm ID... - deletes a reminders with ID... identifiers
- timezone, tz LOCATION - updates channel timezone
- timezone, tz - shows current location

CRON-RULE:
- "Seconds Minutes Hours DayOfMonth Month DayOfWeek Year"
- "Minutes Hours DayOfMonth Month DayOfWeek Year" (Seconds default to 0)
- "Minutes Hours DayOfMonth Month DayOfWeek" (Year defaults to *)
- Month: 1-12 or JAN-DEC
- DayOfWeek 0-7 or SUN-SAT (both 0 and 7 stand for SUN)
- ` + "`*`" + ` - any value ("0 12 * * *" - 12:00 every day every month every year)
- ` + "`/`" + ` - time period ("*/5 * * * *" - every 5 minute of every day every month every year)
- ` + "`,`" + ` - list separator ("0 12 10,25 * *" - 12:00 every 10th and 25th day of every month)
- ` + "`-`" + ` - range ("0 12 * MON-FRI *" - 12:00 every workday)
- ` + "`L`" + ` - last ("0 12 * 5L *" - 12:00 last friday every month)
- ` + "`#`" + ` - numbered ("0 12 * TUE#2 *" - 12:00 second tuesday of every month)
LOCATION: TZ identifier (for example "Asia/Novosibirsk")
`

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
) (string, bool) {
	if len(tokens) > 0 && strings.EqualFold(req.Command, "/reminder") {
		var str string
		var err error
		switch tokens[0] {
		case "add", "create":
			str, err = mmReminderCreate(app, req, tokens)
		case "list", "ls":
			str, err = services.MMReminderList(app, req)
		case "delete", "del", "remove", "rm":
			str, err = services.MMReminderDelete(app, req, tokens)
		case "timezone", "tz":
			str, err = mmReminderTimeZone(app, req, tokens)
		}
		if err != nil {
			return fmt.Sprintf("Error: %s", err), true
		} else if str != "" {
			return str, true
		}
	}
	return "", false
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

	if str, ok := processCommands(app, req, tokens); ok {
		c.JSON(http.StatusOK, gin.H{"text": str})
	} else {
		c.JSON(http.StatusOK, gin.H{"text": usage})
	}
}
