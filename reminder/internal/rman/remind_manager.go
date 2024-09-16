package rman

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/internal/syncmap"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/repositories"
	"github.com/gorhill/cronexpr"
	"github.com/rs/zerolog"
)

type RemindManager interface {
	TriggerReminds(reminders ...models.Reminder)
	GetReminds() []models.Reminder
	CompleteReminds(ids ...int64)
	AddReminders(reminders ...models.Reminder)
	RemoveReminders(ids ...int64)
}

type defaultRemindManager struct {
	cancels         *syncmap.Map[int64, chan<- bool]
	completes       *syncmap.Map[int64, chan<- bool]
	reminds         *syncmap.Map[int64, models.Reminder]
	defaultLocation *time.Location
	logger          zerolog.Logger
	db              *sql.DB
}

func New(db *sql.DB, defaultLocation *time.Location, logger zerolog.Logger) RemindManager {
	return &defaultRemindManager{
		cancels:         syncmap.New[int64, chan<- bool](),
		completes:       syncmap.New[int64, chan<- bool](),
		reminds:         syncmap.New[int64, models.Reminder](),
		db:              db,
		logger:          logger,
		defaultLocation: defaultLocation,
	}
}

func (rm *defaultRemindManager) TriggerReminds(reminders ...models.Reminder) {
	for _, reminder := range reminders {
		rm.reminds.Set(reminder.ID, reminder)
	}
}

func (rm *defaultRemindManager) GetReminds() []models.Reminder {
	var reminders []models.Reminder
	rm.reminds.Range(func(key int64, value models.Reminder) bool {
		reminders = append(reminders, value)
		return true
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
			log.Printf(
				"Error while parsing cron expression '%s': %s\n", reminder.Rule, err,
			)
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

func (rm *defaultRemindManager) generateReminds(
	reminder models.Reminder,
	expr *cronexpr.Expression,
	cancel <-chan bool,
	complete <-chan bool,
) {
	rm.logger.Info().Str("Reminder", fmt.Sprintf("%v", reminder)).Msg("Starts generating reminds")

	for {
		now := time.Now().In(rm.defaultLocation)
		if channel, err := repositories.GetChannel(rm.db, reminder.Channel); err == nil {
			if loc, err := time.LoadLocation(channel.TimeZone); err == nil {
				now = now.In(loc)
			} else {
				rm.logger.Error().Err(err).Str("Location", loc.String()).Any("Channel", channel).Msg("Cannot parse location")
			}
		} else {
			rm.logger.Error().Err(err).Any("Channel", reminder.Channel).Msg("Channel ot found in db")
		}

		nextTime := expr.Next(now).UTC()
		if nextTime.IsZero() {
			rm.RemoveReminders(reminder.ID)
			repositories.DeleteReminder(rm.db, reminder.ID)
			return
		}

		timer := time.NewTimer(time.Until(nextTime))
		rm.logger.Info().Any("Reminder", reminder).Time("Next time", nextTime).Msg("Next trigger time calculated")
		select {
		case <-timer.C:
			go func() {
				rm.reminds.Set(reminder.ID, reminder)
			}()
			<-complete
		case <-cancel:
			return
		}
	}
}
