// Test run requires running db instance. It was tested with `docker compose up db -d` executed in git root directory.
package rman_test

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/internal/rman"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/models"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lpernett/godotenv"
	"github.com/stretchr/testify/suite"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type TestSuite struct {
	suite.Suite
	db *sql.DB
}

func TestRemindManager(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupSuite() {
	godotenv.Load(filepath.Join("..", "..", "..", ".env"))

	cfg := mysql.Config{
		User:            getEnvOrDefault("MYSQL_USER", "reminders"),
		Passwd:          getEnvOrDefault("MYSQL_PASSWORD", "XXXXXX"),
		Net:             "tcp",
		Addr:            fmt.Sprintf("%s:%s", "localhost", "3306"), // todo: load address and port from test env veriables
		DBName:          getEnvOrDefault("DB_NAME", "reminders"),
		MultiStatements: true,
	}

	var err error
	s.db, err = sql.Open("mysql", cfg.FormatDSN())
	s.Require().NoError(err, "failed to open mysql connection")

	err = s.db.Ping()
	s.Require().NoError(err, "failed to ping mysql connection")

	driver, err := migratemysql.WithInstance(s.db, &migratemysql.Config{})
	s.Require().NoError(err, "failed to initiate migrate driver")

	migrateInstance, err := migrate.NewWithDatabaseInstance(
		"file://"+filepath.Join("..", "..", "migrations"),
		"mysql",
		driver,
	)
	s.Require().NoError(err, "failed to initiate migrate instance")

	err = migrateInstance.Up()
	s.Require().True(err == nil || errors.Is(err, migrate.ErrNoChange), "failed to migrate")
}

func (s *TestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *TestSuite) TestBasicOperations() {
	location, err := time.LoadLocation("UTC")
	s.Require().NoError(err)

	s.Run("TriggerReminds and GetReminds", func() {
		rm := rman.New(s.db, location)
		remind := models.Remind{
			ReminderId: 1,
			Name:       "Test Remind",
			Rule:       "* * * * *",
			Channel:    "test-channel",
			Message:    "Test message",
		}

		rm.TriggerReminds(remind)
		reminds := rm.GetReminds()

		s.Len(reminds, 1)
		s.Equal(remind, reminds[0])
	})

	s.Run("CompleteReminds", func() {
		rm := rman.New(s.db, location)
		reminder := models.Reminder{
			ID:      1,
			Name:    "Test Reminder",
			Rule:    "* * * * * * *",
			Channel: "test-channel",
			Message: "Test message",
		}

		rm.AddReminders(reminder)
		rm.CompleteReminds(reminder.ID)

		reminds := rm.GetReminds()
		s.Empty(reminds)
	})

	s.Run("AddReminders and RemoveReminders", func() {
		rm := rman.New(s.db, location)
		reminder := models.Reminder{
			ID:      1,
			Name:    "Test Reminder",
			Rule:    "* * * * *",
			Channel: "test-channel",
			Message: "Test message",
			Owner:   sql.NullString{String: "test-owner", Valid: true},
		}

		rm.AddReminders(reminder)

		rm.RemoveReminders(reminder.ID)
		reminds := rm.GetReminds()

		s.Empty(reminds)
	})
}

func (s *TestSuite) TestRemindManagerModifications() {
	location, err := time.LoadLocation("UTC")
	s.Require().NoError(err)

	s.Run("UpdateReminderOwner", func() {
		rm := rman.New(s.db, location)
		remind := models.Remind{
			ReminderId: 1,
			Name:       "Test Remind",
			Rule:       "* * * * *",
			Channel:    "test-channel",
			Message:    "Test message",
		}

		rm.TriggerReminds(remind)
		rm.UpdateReminderOwner(remind.ReminderId, "new-owner")

		reminds := rm.GetReminds()
		s.Len(reminds, 1)
		s.Equal("new-owner", reminds[0].Owner.String)
		s.True(reminds[0].Owner.Valid)
	})

	s.Run("UpdateRemindWebhook", func() {
		rm := rman.New(s.db, location)
		remind := models.Remind{
			ReminderId: 1,
			Name:       "Test Remind",
			Rule:       "* * * * *",
			Channel:    "test-channel",
			Message:    "Test message",
		}

		rm.TriggerReminds(remind)
		rm.UpdateRemindWebhook(remind.ReminderId, "http://test-webhook.com")

		reminds := rm.GetReminds()
		s.Len(reminds, 1)
		s.Equal("http://test-webhook.com", reminds[0].Webhook)
	})
}

func (s *TestSuite) TestConcurrentAccess() {
	location, err := time.LoadLocation("UTC")
	s.Require().NoError(err)

	s.Run("concurrent trigger and complete", func() {
		rm := rman.New(s.db, location)
		const n = 100

		// Create test reminds
		reminders := make([]models.Reminder, n)
		reminds := make([]models.Remind, n)
		for i := range n {
			reminders[i] = models.Reminder{
				ID:      int64(i + 1),
				Name:    "TestReminder",
				Rule:    "* * * * * * *",
				Channel: "test-channel",
				Message: "Test message",
			}
			reminds[i] = models.Remind{
				ReminderId: 0,
				Name:       "TestReminder",
				Rule:       "* * * * * * *",
				Channel:    "test-channel",
				Message:    "Test message",
			}
		}

		// Trigger all reminds concurrently
		var wg sync.WaitGroup
		wg.Add(n * 2)
		for _, r := range reminders {
			reminder := r
			go func() {
				rm.AddReminders(reminder)
				wg.Done()
			}()
		}

		// Complete all reminds concurrently
		for _, r := range reminds {
			remind := r
			go func() {
				rm.CompleteReminds(remind.ReminderId)
				wg.Done()
			}()
		}

		// Wait for all completes
		wg.Wait()

		// Verify all reminds are removed
		remainingReminds := rm.GetReminds()
		s.Empty(remainingReminds)
	})
}
