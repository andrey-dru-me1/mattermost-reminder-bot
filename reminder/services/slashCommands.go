package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/dtos"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
)

type undeleted struct {
	id  int64
	err error
}

func MMReminderCreate(
	app *app.Application,
	req dtos.MMRequest,
	tokens []string,
) error {
	if len(tokens) < 4 {
		return fmt.Errorf("wrong argument count")
	}

	rem := dtos.ReminderDTO{
		Name:    tokens[1],
		Rule:    tokens[2],
		Message: tokens[3],
		Channel: req.ChannelName,
	}
	_, err := CreateReminder(app, rem)
	if err != nil {
		return fmt.Errorf("create reminder: %w", err)
	}

	return nil
}

func MMReminderList(app *app.Application, req dtos.MMRequest) (string, error) {
	reminders, err := GetRemindersByChannel(app, req.ChannelName)
	if err != nil {
		return "", fmt.Errorf("get reminders: %w", err)
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

		return sb.String(), nil
	}
	return "There are no reminders in this channel yet! Add a new one using `/reminder add ...`", nil
}

func deleteRemindersAndCollect(
	app *app.Application,
	req dtos.MMRequest,
	tokens []string,
) (deleted []int64, undels []undeleted) {
	for _, reminderIDString := range tokens[1:] {
		id, err := strconv.ParseInt(reminderIDString, 10, 64)
		if err != nil {
			undels = append(
				undels,
				undeleted{id: id, err: fmt.Errorf("parse id: %w", err)},
			)
			continue
		}

		reminder, err := GetReminder(app, id)
		if err != nil {
			undels = append(
				undels,
				undeleted{id: id, err: fmt.Errorf("get reminder: %w", err)},
			)
			continue
		}
		if !strings.EqualFold(reminder.Channel, req.ChannelName) {
			undels = append(
				undels,
				undeleted{
					id: id,
					err: fmt.Errorf(
						"get reminder: invalid access: reminder %d belongs to channel "+
							"'%s' and cannot be deleted from channel '%s'",
						id,
						reminder.Channel,
						req.ChannelName,
					),
				},
			)
			continue
		}

		if err := DeleteReminder(app, id); err != nil {
			undels = append(
				undels,
				undeleted{
					id:  id,
					err: fmt.Errorf("delete reminder from database: %w", err),
				},
			)
			continue
		}

		deleted = append(deleted, id)
	}
	return
}

func constructMessage(deleted []int64, undels []undeleted) string {
	var sb strings.Builder
	for _, undel := range undels {
		sb.WriteString(
			fmt.Sprintf(
				"Error deleting %d reminder: %s\n",
				undel.id,
				undel.err,
			),
		)
	}
	if len(deleted) > 0 {
		sb.WriteString("\nSuccessfully deleted: ")
		for i, del := range deleted {
			if i != 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%d", del))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func MMReminderDelete(
	app *app.Application,
	req dtos.MMRequest,
	tokens []string,
) (string, error) {
	if len(tokens) < 2 {
		return "", fmt.Errorf("wrong argument count")
	}

	deleted, undels := deleteRemindersAndCollect(app, req, tokens)

	return constructMessage(deleted, undels), nil
}

func MMReminderTimeZoneSet(
	app *app.Application,
	req dtos.MMRequest,
	tokens []string,
) (string, error) {
	timeZone := tokens[1]
	if _, err := time.LoadLocation(timeZone); err != nil {
		return "", fmt.Errorf("parse timezone: %w", err)
	}

	if err := InsertChannel(
		app,
		models.Channel{Name: req.ChannelName, TimeZone: timeZone},
	); err != nil {
		return "", fmt.Errorf("insert channel: %w", err)
	}
	return timeZone, nil
}

func MMReminderTimeZoneGet(app *app.Application, req dtos.MMRequest) string {
	channel, err := GetChannel(app, req.ChannelName)
	if err != nil {
		return fmt.Sprintf(
			"Time zone is not set for the channel '%s'. Using default time zone: %v.\n",
			req.ChannelName,
			app.DefaultLocation,
		)
	}
	return fmt.Sprintf("Time zone: %s", channel.TimeZone)
}
