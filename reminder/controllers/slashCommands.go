package controllers

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	_ "time/tzdata"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/dtos"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/services"
	"github.com/gin-gonic/gin"
)

const usage = `Usage: /reminder COMMAND OPTIONS
Commands:
- add, create NAME CRON-RULE MESSAGE - creates new reminder
- list, ls - lists all reminders
- delete, del, remove, rm ID... - deletes a reminders with ID... identifiers
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

func mmReminderCreate(c *gin.Context, app *app.Application, req dtos.MMRequest, tokens []string) {
	if err := services.MMReminderCreate(app, req, tokens); err != nil {
		txt := err.Error()
		c.JSON(http.StatusOK, gin.H{"text": strings.ToUpper(txt[:1]) + txt[1:]})
	} else {
		c.JSON(
			http.StatusOK,
			gin.H{"text": "Reminder successfully created"},
		)
	}
}

func mmReminderList(c *gin.Context, app *app.Application, req dtos.MMRequest) {
	if str, err := services.MMReminderList(app, req); err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{"text": fmt.Sprintf("Error: %s", err)},
		)
	} else {
		c.JSON(
			http.StatusOK,
			gin.H{"text": str},
		)
	}
}

func mmReminderDelete(c *gin.Context, app *app.Application, tokens []string) {
	if str, err := services.MMReminderDelete(app, tokens); err != nil {
		txt := err.Error()
		c.JSON(http.StatusOK, gin.H{"text": strings.ToUpper(txt[:1]) + txt[1:]})
	} else {
		c.JSON(http.StatusOK, gin.H{"text": str})
	}
}

func mmReminderTimeZone(c *gin.Context, app *app.Application, req dtos.MMRequest, tokens []string) {
	if len(tokens) <= 1 {
		c.JSON(
			http.StatusOK,
			gin.H{"text": services.MMReminderTimeZoneGet(app, req)},
		)
	} else {
		if str, err := services.MMReminderTimeZoneSet(app, req, tokens); err != nil {
			c.JSON(http.StatusOK, gin.H{"text": fmt.Sprintf("Error: %s", err)})
		} else {
			c.JSON(http.StatusOK, gin.H{"text": fmt.Sprintf("Time zone set to %s", str)})
		}
	}
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

	tokens := tokenize(req.Text)

	if len(tokens) > 0 && strings.EqualFold(req.Command, "/reminder") {
		switch tokens[0] {
		case "add", "create":
			mmReminderCreate(c, app, req, tokens)
			return
		case "list", "ls":
			mmReminderList(c, app, req)
			return
		case "delete", "del", "remove", "rm":
			mmReminderDelete(c, app, tokens)
			return
		case "timezone", "tz":
			mmReminderTimeZone(c, app, req, tokens)
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
