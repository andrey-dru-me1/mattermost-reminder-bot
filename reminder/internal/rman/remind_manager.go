package rman

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/internal/syncmap"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/repositories"
	"github.com/gorhill/cronexpr"
	"github.com/rs/zerolog/log"
)

type RemindManager interface {
	TriggerReminds(reminders ...models.Remind)
	GetReminds() []models.Remind
	CompleteReminds(ids ...int64)
	AddReminders(reminders ...models.Reminder)
	UpdateReminderOwner(id int64, owner string)
	UpdateRemindWebhook(id int64, webhook string)
	RemoveReminders(ids ...int64)
}

type defaultRemindManager struct {
	cancels         *syncmap.Map[int64, chan<- bool]
	completes       *syncmap.Map[int64, chan<- bool]
	reminds         *syncmap.Map[int64, models.Remind]
	defaultLocation *time.Location
	db              *sql.DB
}

func New(
	db *sql.DB,
	defaultLocation *time.Location,
) RemindManager {
	return &defaultRemindManager{
		cancels:         syncmap.New[int64, chan<- bool](),
		completes:       syncmap.New[int64, chan<- bool](),
		reminds:         syncmap.New[int64, models.Remind](),
		db:              db,
		defaultLocation: defaultLocation,
	}
}

func (rm *defaultRemindManager) TriggerReminds(reminds ...models.Remind) {
	for _, remind := range reminds {
		rm.reminds.Set(remind.ReminderId, remind)
	}
}

func (rm *defaultRemindManager) GetReminds() []models.Remind {
	var reminders []models.Remind
	rm.reminds.Range(func(key int64, value models.Remind) error {
		reminders = append(reminders, value)
		return nil
	})
	return reminders
}

func (rm *defaultRemindManager) CompleteReminds(ids ...int64) {
	for _, id := range ids {
		if complete, ok := rm.completes.Get(id); ok {
			complete <- true
			rm.reminds.Delete(id)
		}
	}
}

func (rm *defaultRemindManager) AddReminders(reminders ...models.Reminder) {
	for _, reminder := range reminders {
		expr, err := cronexpr.Parse(reminder.Rule)
		if err != nil {
			log.Err(err).
				Str("rule", reminder.Rule).
				Msg("Cannot parse cron expression")
			continue
		}

		cancel := make(chan bool)
		complete := make(chan bool)

		rm.cancels.Set(reminder.ID, cancel)
		rm.completes.Set(reminder.ID, complete)

		go rm.generateReminds(reminder, expr, cancel, complete)
	}
}

func (rm *defaultRemindManager) RemoveReminders(ids ...int64) {
	for _, id := range ids {
		if cancel, ok := rm.cancels.Get(id); ok {
			select {
			case cancel <- true:
			default:
			}
			close(cancel)

			if complete, ok := rm.completes.Get(id); ok {
				select {
				case complete <- true:
				default:
				}
				close(complete)
			}

			rm.reminds.Delete(id)
			rm.completes.Delete(id)
			rm.cancels.Delete(id)
		}
	}
}

func (rm *defaultRemindManager) UpdateReminderOwner(id int64, owner string) {
	rm.reminds.Apply(id, func(remind models.Remind) models.Remind {
		remind.Owner = sql.NullString{String: owner, Valid: true}
		return remind
	})
}

func (rm *defaultRemindManager) UpdateRemindWebhook(id int64, webhook string) {
	rm.reminds.Apply(id, func(remind models.Remind) models.Remind {
		remind.Webhook = webhook
		return remind
	})
}

func (rm *defaultRemindManager) generateReminds(
	reminder models.Reminder,
	expr *cronexpr.Expression,
	cancel <-chan bool,
	complete <-chan bool,
) {
	log.Info().
		Str("Reminder", fmt.Sprintf("%v", reminder)).
		Msg("Starts generating reminds")

	for {
		now := time.Now().In(rm.defaultLocation)
		if channel, err := repositories.GetChannel(rm.db, reminder.Channel); err == nil {
			if loc, err := time.LoadLocation(channel.TimeZone); err == nil {
				now = now.In(loc)
			} else {
				log.Warn().
					Err(err).
					Str("Location", loc.String()).
					Any("Channel", channel).
					Interface("Default location", rm.defaultLocation).
					Msg("Cannot parse location, using default TZ")
			}
		} else {
			log.Error().Err(err).Any("Channel", reminder.Channel).Msg("Channel not found in db")
		}

		nextTime := expr.Next(now).UTC()
		if nextTime.IsZero() {
			rm.RemoveReminders(reminder.ID)
			repositories.DeleteReminder(rm.db, reminder.ID)
			return
		}

		timer := time.NewTimer(time.Until(nextTime))
		log.Info().
			Any("Reminder", reminder).
			Time("Next time", nextTime).
			Msg("Next trigger time calculated")
		select {
		case <-timer.C:
			go func() {
				rm.reminds.Set(reminder.ID, rm.reminderToRemind(reminder))
			}()
			<-complete
		case <-cancel:
			return
		}
	}
}

func (rm *defaultRemindManager) reminderToRemind(
	reminder models.Reminder,
) models.Remind {
	remind := models.Remind{
		ReminderId: reminder.ID,
		Owner:      reminder.Owner,
		Name:       reminder.Name,
		Rule:       reminder.Rule,
		Channel:    reminder.Channel,
		Message:    reminder.Message,
	}

	if reminder.Owner.Valid {
		user, err := repositories.GetUser(rm.db, reminder.Owner.String)
		if err == nil && user.Webhook.Valid {
			remind.Webhook = user.Webhook.String
		}
	}

	return remind
}
