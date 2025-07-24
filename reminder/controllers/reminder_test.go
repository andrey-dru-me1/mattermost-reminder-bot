package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/app"
	"github.com/andrey-dru-me1/mattermost-reminder-bot/reminder/controllers"
	"github.com/gin-gonic/gin"
	"github.com/lpernett/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateReminder(t *testing.T) {
	os.Chdir("..")
	godotenv.Load(filepath.Join("..", ".env"))
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "3306")

	application, err := app.SetupApplication()
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("app", application)
		ctx.Request = httptest.NewRequest(
			"POST",
			"http://localhost/reminders",
			strings.NewReader(
				`{
					"name":    "test-name",
					"owner":   "test-owner",
					"rule":    "* * * * *",
					"channel": "test-channel",
					"message": "test-message"
				}`,
			),
		)

		controllers.CreateReminder(ctx)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		var resp struct {
			Message string `json:"message"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, resp.Message, "Reminder created successfully")
	})

	t.Run("Wrong cron expr", func(t *testing.T) {
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("app", application)
		ctx.Request = httptest.NewRequest(
			"POST",
			"http://localhost/reminders",
			strings.NewReader(
				`{
					"name":    "test-name",
					"owner":   "test-owner",
					"rule":    "wrong cron expr",
					"channel": "test-channel",
					"message": "test-message"
				}`,
			),
		)

		controllers.CreateReminder(ctx)

		assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
		var resp struct {
			Error string `json:"error"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Contains(t, resp.Error, "parse cron expr")
	})

	t.Run("Wrong json", func(t *testing.T) {
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("app", application)
		ctx.Request = httptest.NewRequest(
			"POST",
			"http://localhost/reminders",
			strings.NewReader(
				`Incorrect json`,
			),
		)

		controllers.CreateReminder(ctx)

		// fixme: this status code not uses what is set in tested function
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("No app", func(t *testing.T) {
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest(
			"POST",
			"http://localhost/reminders",
			strings.NewReader(
				`{
					"name":    "test-name",
					"owner":   "test-owner",
					"rule":    "* * * * *",
					"channel": "test-channel",
					"message": "test-message"
				}`,
			),
		)

		controllers.CreateReminder(ctx)

		assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
		var resp struct {
			Error string `json:"error"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Contains(t, resp.Error, "application not provided")
	})

	t.Run("Wrong app type", func(t *testing.T) {
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("app", 32)
		ctx.Request = httptest.NewRequest(
			"POST",
			"http://localhost/reminders",
			strings.NewReader(
				`{
					"name":    "test-name",
					"owner":   "test-owner",
					"rule":    "* * * * *",
					"channel": "test-channel",
					"message": "test-message"
				}`,
			),
		)

		controllers.CreateReminder(ctx)

		assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
		var resp struct {
			Error string `json:"error"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Contains(t, resp.Error, "application has wrong type")
	})

	application.RemindManager.Close() // todo: move remind manager setup out of application setup and use defer to close remind manager
}
