package rman

import (
	"log"
	"time"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/internal/syncmap"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
	"github.com/gorhill/cronexpr"
)

type RemindManager interface {
	AddReminders(reminders ...models.Reminder)
	RemoveReminders(ids ...int)
	TriggerReminds(reminders ...models.Reminder)
	GetReminds() []models.Reminder
	CompleteReminds(ids ...int)
}

type defaultRemindManager struct {
	cancels   *syncmap.Map[int, chan bool]
	completes *syncmap.Map[int, chan bool]
	reminds   *syncmap.Map[int, models.Reminder]
}

func New() RemindManager {
	return &defaultRemindManager{
		cancels:   syncmap.New[int, chan bool](),
		completes: syncmap.New[int, chan bool](),
		reminds:   syncmap.New[int, models.Reminder]()}
}

func (drm *defaultRemindManager) TriggerReminds(reminders ...models.Reminder) {
	for _, reminder := range reminders {
		drm.reminds.Set(reminder.ID, reminder)
	}
}

func (drm *defaultRemindManager) GetReminds() []models.Reminder {
	var reminders []models.Reminder
	drm.reminds.Range(func(key int, value models.Reminder) bool {
		reminders = append(reminders, value)
		return true
	})
	return reminders
}

func (drm *defaultRemindManager) CompleteReminds(ids ...int) {
	for _, id := range ids {
		if complete, ok := drm.completes.Get(id); ok {
			complete <- true
			drm.reminds.Delete(id)
		}
	}
}

func (drm *defaultRemindManager) AddReminders(reminders ...models.Reminder) {
	for _, reminder := range reminders {
		drm.cancels.Set(reminder.ID, make(chan bool))
		drm.completes.Set(reminder.ID, make(chan bool))
		go drm.generateReminds(reminder)
	}
}

func (drm *defaultRemindManager) RemoveReminders(ids ...int) {
	for _, id := range ids {
		if cancel, ok := drm.cancels.Get(id); ok {
			cancel <- true
			close(cancel)

			if complete, ok := drm.completes.Get(id); ok {
				select {
				case complete <- true:
				default:
				}
				close(complete)
			}

			drm.reminds.Delete(id)
			drm.completes.Delete(id)
			drm.cancels.Delete(id)
		}
	}
}

func (drm *defaultRemindManager) generateReminds(reminder models.Reminder) {
	expr, err := cronexpr.Parse(reminder.Rule)
	if err != nil {
		log.Printf(
			"Error while parsing cron expression '%s': %s\n", reminder.Rule, err,
		)
		return
	}
	log.Printf("Reminder %d (%s) starts generating reminds\n", reminder.ID, reminder.Name)

	cancel, ok := drm.cancels.Get(reminder.ID)
	if !ok {
		log.Printf("Cancel channel for a reminder with id %d was not created\n", reminder.ID)
		return
	}

	complete, ok := drm.completes.Get(reminder.ID)
	if !ok {
		log.Printf("Cancel channel for a reminder with id %d was not created\n", reminder.ID)
		return
	}

	for {
		nextTime := expr.Next(time.Now())
		timer := time.NewTimer(time.Until(nextTime))
		log.Printf("Next trigger time for reminder '%s' is: %v", reminder.Name, nextTime)
		select {
		case <-timer.C:
			go func() {
				drm.reminds.Set(reminder.ID, reminder)
			}()
			<-complete
		case <-cancel:
			return
		}
	}
}
