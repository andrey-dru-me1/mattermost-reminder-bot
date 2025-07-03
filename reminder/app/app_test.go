// Test run requires running db instance. It was tested with `docker compose up db -d` executed in git root directory.
package app_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/lpernett/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupApplication(t *testing.T) {
	currentDir, err := os.Getwd()
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		os.Chdir(filepath.Join(currentDir, ".."))
		godotenv.Load((filepath.Join("..", ".env")))
		stringLocation := "America/Toronto"
		location, err := time.LoadLocation(stringLocation)
		require.NoError(t, err)

		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "3306")
		os.Setenv("DEFAULT_TZ", stringLocation)

		application, err := app.SetupApplication()
		require.NoError(t, err)
		assert.Equal(t, location, application.DefaultLocation)
		assert.NotNil(t, application.Db)
		assert.NoError(t, application.Db.Ping())
	})

	t.Run("Load location failure", func(t *testing.T) {
		os.Chdir(filepath.Join(currentDir, ".."))
		godotenv.Load((filepath.Join("..", ".env")))
		stringLocation := "unknown location"

		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "3306")
		os.Setenv("DEFAULT_TZ", stringLocation)

		application, err := app.SetupApplication()
		require.NoError(t, err)
		assert.Equal(t, time.UTC, application.DefaultLocation)
		assert.NotNil(t, application.Db)
		assert.NoError(t, application.Db.Ping())
	})

	t.Run("Run migrations failure", func(t *testing.T) {
		os.Chdir(currentDir)
		godotenv.Load((filepath.Join(currentDir, "..", "..", ".env")))

		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "3306")

		application, err := app.SetupApplication()
		require.Error(t, err)
		assert.Nil(t, application)
		assert.ErrorContains(t, err, "run migrations")
	})

	t.Run("Setup database failure", func(t *testing.T) {
		os.Chdir(filepath.Join(currentDir, ".."))
		godotenv.Load((filepath.Join("..", ".env")))

		os.Setenv("DB_HOST", "unknown")

		application, err := app.SetupApplication()
		require.Error(t, err)
		assert.Nil(t, application)
		assert.ErrorContains(t, err, "setup database")
	})
}
