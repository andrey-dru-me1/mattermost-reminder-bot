package controllers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/controllers/dtos"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/services"
	"github.com/gin-gonic/gin"
)

type mattermostRequest struct {
	ChannelName string `form:"channel_name"`
	Command     string `form:"command"`
	Text        string `form:"text"`
}

func mattermostReminderCreate(c *gin.Context, req mattermostRequest, tokens []string) {
	if len(tokens) < 3 {
		c.JSON(http.StatusOK, gin.H{"text": `Usage: '/reminder create [NAME] "[CRON-RULE]"'`})
		return
	}

	db, err := extractDB(c)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{"text": err.Error()},
		)
		return
	}

	rem := dtos.ReminderDTO{
		Name:    tokens[1],
		Rule:    tokens[2],
		Channel: req.ChannelName,
	}
	_, err = services.CreateReminder(db, rem)
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

func mattermostReminderList(c *gin.Context) {
	db, err := extractDB(c)
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{"text": err.Error()},
		)
		return
	}

	reminders, err := services.GetReminders(db)
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

func mattemostReminderDelete(c *gin.Context, req mattermostRequest, tokens []string) {

}

func MattermostReminder(c *gin.Context) {
	var req mattermostRequest

	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens := tokenize(req.Text)

	if strings.EqualFold(req.Command, "/reminder") {
		switch tokens[0] {
		case "add":
			mattermostReminderCreate(c, req, tokens)
			return
		case "list":
			mattermostReminderList(c)
			return
		}
	}
	c.JSON(
		http.StatusOK,
		gin.H{"text": "Usage: '/reminder [add|list]'"},
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
