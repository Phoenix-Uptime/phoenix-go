package server

import (
	"bytes"
	"context"
	stdsql "database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entmonitorcheck "github.com/Phoenix-Uptime/phoenix-go/ent/monitorcheck"
	entmonitorstat "github.com/Phoenix-Uptime/phoenix-go/ent/monitorstat"
	entnotificationchannel "github.com/Phoenix-Uptime/phoenix-go/ent/notificationchannel"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	accountroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/account"
	authroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/auth"
	monitorroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/monitors"
	notificationchannelroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/notification_channels"
	tagroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/tags"
	"github.com/gofiber/fiber/v3"
)

func TestHealthRoute(t *testing.T) {
	app := New()

	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestAccountSettingsRoutesRequireAuth(t *testing.T) {
	app := New()

	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/account/settings"},
		{method: http.MethodPost, path: "/account/settings/smtp"},
		{method: http.MethodPost, path: "/account/settings/telegram"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
			}
		})
	}
}

func TestSignupAndLoginWithEnt(t *testing.T) {
	cleanup := useTestDatabase(t)
	defer cleanup()

	app := New()

	signupBody := []byte(`{"username":"exampleuser","email":"user@example.com","password":"examplepassword"}`)
	signupReq, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewReader(signupBody))
	if err != nil {
		t.Fatal(err)
	}
	signupReq.Header.Set("Content-Type", "application/json")

	signupResp, err := app.Test(signupReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if signupResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected signup status %d, got %d", http.StatusCreated, signupResp.StatusCode)
	}

	loginBody := []byte(`{"username":"exampleuser","password":"examplepassword"}`)
	loginReq, err := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatal(err)
	}
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := app.Test(loginReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("expected login status %d, got %d", http.StatusOK, loginResp.StatusCode)
	}
	defer loginResp.Body.Close()

	var body authroutes.LoginResponse
	if err := json.NewDecoder(loginResp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.ApiKey == "" {
		t.Fatal("expected login response to include api key")
	}
}

func TestAccountSettingsUseNotificationChannelRows(t *testing.T) {
	cleanup := useTestDatabase(t)
	defer cleanup()

	app := New()

	signupBody := []byte(`{"username":"settingsuser","email":"settings@example.com","password":"examplepassword"}`)
	signupReq, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewReader(signupBody))
	if err != nil {
		t.Fatal(err)
	}
	signupReq.Header.Set("Content-Type", "application/json")

	signupResp, err := app.Test(signupReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if signupResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected signup status %d, got %d", http.StatusCreated, signupResp.StatusCode)
	}

	loginBody := []byte(`{"username":"settingsuser","password":"examplepassword"}`)
	loginReq, err := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatal(err)
	}
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := app.Test(loginReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer loginResp.Body.Close()

	var login authroutes.LoginResponse
	if err := json.NewDecoder(loginResp.Body).Decode(&login); err != nil {
		t.Fatal(err)
	}
	if login.ApiKey == "" {
		t.Fatal("expected login response to include api key")
	}

	smtpBody := []byte(`{"smtp_server":"smtp.example.com","smtp_port":587,"from_address":"noreply@example.com","username":"mailer@example.com","password":"supersecret","use_tls":true}`)
	smtpReq, err := http.NewRequest(http.MethodPost, "/account/settings/smtp", bytes.NewReader(smtpBody))
	if err != nil {
		t.Fatal(err)
	}
	smtpReq.Header.Set("Content-Type", "application/json")
	smtpReq.Header.Set("x-api-key", login.ApiKey)

	smtpResp, err := app.Test(smtpReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if smtpResp.StatusCode != http.StatusOK {
		t.Fatalf("expected SMTP settings status %d, got %d", http.StatusOK, smtpResp.StatusCode)
	}

	telegramBody := []byte(`{"bot_token":"123456789:ABCdefGHIjklMNOpqrSTUvwxyz"}`)
	telegramReq, err := http.NewRequest(http.MethodPost, "/account/settings/telegram", bytes.NewReader(telegramBody))
	if err != nil {
		t.Fatal(err)
	}
	telegramReq.Header.Set("Content-Type", "application/json")
	telegramReq.Header.Set("x-api-key", login.ApiKey)

	telegramResp, err := app.Test(telegramReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if telegramResp.StatusCode != http.StatusOK {
		t.Fatalf("expected Telegram settings status %d, got %d", http.StatusOK, telegramResp.StatusCode)
	}

	settingsReq, err := http.NewRequest(http.MethodGet, "/account/settings", nil)
	if err != nil {
		t.Fatal(err)
	}
	settingsReq.Header.Set("x-api-key", login.ApiKey)

	settingsResp, err := app.Test(settingsReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if settingsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected settings status %d, got %d", http.StatusOK, settingsResp.StatusCode)
	}
	defer settingsResp.Body.Close()

	var settings accountroutes.SettingsResponse
	if err := json.NewDecoder(settingsResp.Body).Decode(&settings); err != nil {
		t.Fatal(err)
	}
	if settings.SMTPSettings == nil || settings.SMTPSettings.SMTPServer != "smtp.example.com" {
		t.Fatal("expected settings response to include SMTP settings")
	}
	if settings.TelegramBot == nil || settings.TelegramBot.BotToken == "" {
		t.Fatal("expected settings response to include Telegram bot settings")
	}

	channelCount, err := database.Client.NotificationChannel.Query().
		Where(entnotificationchannel.TypeIn(entnotificationchannel.TypeSMTP, entnotificationchannel.TypeTelegram)).
		Count(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if channelCount != 2 {
		t.Fatalf("expected 2 notification channel rows, got %d", channelCount)
	}
}

func TestResetAPIKeyRotatesAPIKeyRows(t *testing.T) {
	cleanup := useTestDatabase(t)
	defer cleanup()

	app := New()

	signupBody := []byte(`{"username":"keyuser","email":"key@example.com","password":"examplepassword"}`)
	signupReq, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewReader(signupBody))
	if err != nil {
		t.Fatal(err)
	}
	signupReq.Header.Set("Content-Type", "application/json")

	signupResp, err := app.Test(signupReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if signupResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected signup status %d, got %d", http.StatusCreated, signupResp.StatusCode)
	}

	loginBody := []byte(`{"username":"keyuser","password":"examplepassword"}`)
	loginReq, err := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatal(err)
	}
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := app.Test(loginReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer loginResp.Body.Close()

	var login authroutes.LoginResponse
	if err := json.NewDecoder(loginResp.Body).Decode(&login); err != nil {
		t.Fatal(err)
	}
	if login.ApiKey == "" {
		t.Fatal("expected login response to include api key")
	}

	resetReq, err := http.NewRequest(http.MethodPost, "/account/reset-api-key", nil)
	if err != nil {
		t.Fatal(err)
	}
	resetReq.Header.Set("x-api-key", login.ApiKey)

	resetResp, err := app.Test(resetReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if resetResp.StatusCode != http.StatusOK {
		t.Fatalf("expected reset status %d, got %d", http.StatusOK, resetResp.StatusCode)
	}
	defer resetResp.Body.Close()

	var resetBody accountroutes.ResetAPIKeyResponse
	if err := json.NewDecoder(resetResp.Body).Decode(&resetBody); err != nil {
		t.Fatal(err)
	}
	if resetBody.ApiKey == "" || resetBody.ApiKey == login.ApiKey {
		t.Fatal("expected reset response to include a new api key")
	}

	oldKeyReq, err := http.NewRequest(http.MethodGet, "/account/me", nil)
	if err != nil {
		t.Fatal(err)
	}
	oldKeyReq.Header.Set("x-api-key", login.ApiKey)

	oldKeyResp, err := app.Test(oldKeyReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if oldKeyResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected old key status %d, got %d", http.StatusUnauthorized, oldKeyResp.StatusCode)
	}

	newKeyReq, err := http.NewRequest(http.MethodGet, "/account/me", nil)
	if err != nil {
		t.Fatal(err)
	}
	newKeyReq.Header.Set("x-api-key", resetBody.ApiKey)

	newKeyResp, err := app.Test(newKeyReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if newKeyResp.StatusCode != http.StatusOK {
		t.Fatalf("expected new key status %d, got %d", http.StatusOK, newKeyResp.StatusCode)
	}
}

func TestMonitorRoutesCRUD(t *testing.T) {
	cleanup := useTestDatabase(t)
	defer cleanup()

	app := New()
	apiKey := signupAndLogin(t, app, "monitoruser", "monitor@example.com", "examplepassword")

	createBody := []byte(`{"name":"Website","type":"http","url":"https://example.com","interval":60,"timeout":10,"method":"GET","accepted_status_codes":["200-299"],"headers":{"x-test":"true"},"auth_password":"secret"}`)
	createReq, err := http.NewRequest(http.MethodPost, "/monitors", bytes.NewReader(createBody))
	if err != nil {
		t.Fatal(err)
	}
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("x-api-key", apiKey)

	createResp, err := app.Test(createReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create monitor status %d, got %d", http.StatusCreated, createResp.StatusCode)
	}
	defer createResp.Body.Close()

	var created monitorroutes.MonitorResponse
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.Name != "Website" || created.Type != "http" {
		t.Fatalf("unexpected created monitor response: %+v", created)
	}

	listReq, err := http.NewRequest(http.MethodGet, "/monitors", nil)
	if err != nil {
		t.Fatal(err)
	}
	listReq.Header.Set("x-api-key", apiKey)

	listResp, err := app.Test(listReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("expected list monitor status %d, got %d", http.StatusOK, listResp.StatusCode)
	}
	defer listResp.Body.Close()

	var list monitorroutes.MonitorListResponse
	if err := json.NewDecoder(listResp.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list.Monitors) != 1 || list.Monitors[0].ID != created.ID {
		t.Fatalf("expected created monitor in list, got %+v", list.Monitors)
	}

	if _, err := database.Client.MonitorCheck.Create().
		SetMonitorID(created.ID).
		SetStatus(entmonitorcheck.StatusUp).
		SetResponseTimeMs(123).
		Save(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Client.MonitorStat.Create().
		SetMonitorID(created.ID).
		SetPeriod(entmonitorstat.PeriodHour).
		SetPeriodStart(time.Now()).
		SetTotalChecks(1).
		SetUpChecks(1).
		SetUptimePercentage(100).
		Save(context.Background()); err != nil {
		t.Fatal(err)
	}

	checksReq, err := http.NewRequest(http.MethodGet, "/monitors/"+strconv.Itoa(created.ID)+"/checks", nil)
	if err != nil {
		t.Fatal(err)
	}
	checksReq.Header.Set("x-api-key", apiKey)

	checksResp, err := app.Test(checksReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if checksResp.StatusCode != http.StatusOK {
		t.Fatalf("expected monitor checks status %d, got %d", http.StatusOK, checksResp.StatusCode)
	}
	defer checksResp.Body.Close()

	var checks monitorroutes.MonitorChecksResponse
	if err := json.NewDecoder(checksResp.Body).Decode(&checks); err != nil {
		t.Fatal(err)
	}
	if len(checks.Checks) != 1 || checks.Checks[0].ResponseTimeMs != 123 {
		t.Fatalf("expected created monitor check, got %+v", checks.Checks)
	}

	statsReq, err := http.NewRequest(http.MethodGet, "/monitors/"+strconv.Itoa(created.ID)+"/stats?period=hour", nil)
	if err != nil {
		t.Fatal(err)
	}
	statsReq.Header.Set("x-api-key", apiKey)

	statsResp, err := app.Test(statsReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if statsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected monitor stats status %d, got %d", http.StatusOK, statsResp.StatusCode)
	}
	defer statsResp.Body.Close()

	var stats monitorroutes.MonitorStatsResponse
	if err := json.NewDecoder(statsResp.Body).Decode(&stats); err != nil {
		t.Fatal(err)
	}
	if len(stats.Stats) != 1 || stats.Stats[0].UptimePercentage != 100 {
		t.Fatalf("expected created monitor stat, got %+v", stats.Stats)
	}

	updateBody := []byte(`{"name":"Updated Website","timeout":15}`)
	updateReq, err := http.NewRequest(http.MethodPatch, "/monitors/"+strconv.Itoa(created.ID), bytes.NewReader(updateBody))
	if err != nil {
		t.Fatal(err)
	}
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("x-api-key", apiKey)

	updateResp, err := app.Test(updateReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("expected update monitor status %d, got %d", http.StatusOK, updateResp.StatusCode)
	}
	defer updateResp.Body.Close()

	var updated monitorroutes.MonitorResponse
	if err := json.NewDecoder(updateResp.Body).Decode(&updated); err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Updated Website" || updated.Timeout != 15 {
		t.Fatalf("unexpected updated monitor response: %+v", updated)
	}

	pauseReq, err := http.NewRequest(http.MethodPost, "/monitors/"+strconv.Itoa(created.ID)+"/pause", nil)
	if err != nil {
		t.Fatal(err)
	}
	pauseReq.Header.Set("x-api-key", apiKey)

	pauseResp, err := app.Test(pauseReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if pauseResp.StatusCode != http.StatusOK {
		t.Fatalf("expected pause monitor status %d, got %d", http.StatusOK, pauseResp.StatusCode)
	}
	defer pauseResp.Body.Close()

	var paused monitorroutes.MonitorResponse
	if err := json.NewDecoder(pauseResp.Body).Decode(&paused); err != nil {
		t.Fatal(err)
	}
	if paused.IsActive || paused.Status != "paused" {
		t.Fatalf("expected paused inactive monitor, got %+v", paused)
	}

	resumeReq, err := http.NewRequest(http.MethodPost, "/monitors/"+strconv.Itoa(created.ID)+"/resume", nil)
	if err != nil {
		t.Fatal(err)
	}
	resumeReq.Header.Set("x-api-key", apiKey)

	resumeResp, err := app.Test(resumeReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if resumeResp.StatusCode != http.StatusOK {
		t.Fatalf("expected resume monitor status %d, got %d", http.StatusOK, resumeResp.StatusCode)
	}
	defer resumeResp.Body.Close()

	var resumed monitorroutes.MonitorResponse
	if err := json.NewDecoder(resumeResp.Body).Decode(&resumed); err != nil {
		t.Fatal(err)
	}
	if !resumed.IsActive || resumed.Status != "pending" {
		t.Fatalf("expected resumed active monitor, got %+v", resumed)
	}

	deleteReq, err := http.NewRequest(http.MethodDelete, "/monitors/"+strconv.Itoa(created.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	deleteReq.Header.Set("x-api-key", apiKey)

	deleteResp, err := app.Test(deleteReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("expected delete monitor status %d, got %d", http.StatusOK, deleteResp.StatusCode)
	}
}

func TestTagRoutesAndMonitorAssignment(t *testing.T) {
	cleanup := useTestDatabase(t)
	defer cleanup()

	app := New()
	apiKey := signupAndLogin(t, app, "taguser", "tag@example.com", "examplepassword")

	createTagBody := []byte(`{"name":"production","description":"Production monitors","color":"#ff0000"}`)
	createTagReq, err := http.NewRequest(http.MethodPost, "/tags", bytes.NewReader(createTagBody))
	if err != nil {
		t.Fatal(err)
	}
	createTagReq.Header.Set("Content-Type", "application/json")
	createTagReq.Header.Set("x-api-key", apiKey)

	createTagResp, err := app.Test(createTagReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if createTagResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create tag status %d, got %d", http.StatusCreated, createTagResp.StatusCode)
	}
	defer createTagResp.Body.Close()

	var createdTag tagroutes.TagResponse
	if err := json.NewDecoder(createTagResp.Body).Decode(&createdTag); err != nil {
		t.Fatal(err)
	}
	if createdTag.ID == 0 || createdTag.Name != "production" {
		t.Fatalf("unexpected created tag response: %+v", createdTag)
	}

	listTagsReq, err := http.NewRequest(http.MethodGet, "/tags", nil)
	if err != nil {
		t.Fatal(err)
	}
	listTagsReq.Header.Set("x-api-key", apiKey)

	listTagsResp, err := app.Test(listTagsReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if listTagsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected list tags status %d, got %d", http.StatusOK, listTagsResp.StatusCode)
	}
	defer listTagsResp.Body.Close()

	var tagList tagroutes.TagListResponse
	if err := json.NewDecoder(listTagsResp.Body).Decode(&tagList); err != nil {
		t.Fatal(err)
	}
	if len(tagList.Tags) != 1 || tagList.Tags[0].ID != createdTag.ID {
		t.Fatalf("expected created tag in list, got %+v", tagList.Tags)
	}

	createMonitorBody := []byte(`{"name":"Tagged Website","type":"http","url":"https://example.com"}`)
	createMonitorReq, err := http.NewRequest(http.MethodPost, "/monitors", bytes.NewReader(createMonitorBody))
	if err != nil {
		t.Fatal(err)
	}
	createMonitorReq.Header.Set("Content-Type", "application/json")
	createMonitorReq.Header.Set("x-api-key", apiKey)

	createMonitorResp, err := app.Test(createMonitorReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if createMonitorResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create monitor status %d, got %d", http.StatusCreated, createMonitorResp.StatusCode)
	}
	defer createMonitorResp.Body.Close()

	var createdMonitor monitorroutes.MonitorResponse
	if err := json.NewDecoder(createMonitorResp.Body).Decode(&createdMonitor); err != nil {
		t.Fatal(err)
	}

	assignBody := []byte(`{"tag_ids":[` + strconv.Itoa(createdTag.ID) + `]}`)
	assignReq, err := http.NewRequest(http.MethodPut, "/monitors/"+strconv.Itoa(createdMonitor.ID)+"/tags", bytes.NewReader(assignBody))
	if err != nil {
		t.Fatal(err)
	}
	assignReq.Header.Set("Content-Type", "application/json")
	assignReq.Header.Set("x-api-key", apiKey)

	assignResp, err := app.Test(assignReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if assignResp.StatusCode != http.StatusOK {
		t.Fatalf("expected assign monitor tags status %d, got %d", http.StatusOK, assignResp.StatusCode)
	}
	defer assignResp.Body.Close()

	var assigned monitorroutes.MonitorTagsResponse
	if err := json.NewDecoder(assignResp.Body).Decode(&assigned); err != nil {
		t.Fatal(err)
	}
	if len(assigned.Tags) != 1 || assigned.Tags[0].ID != createdTag.ID {
		t.Fatalf("expected assigned monitor tag, got %+v", assigned.Tags)
	}

	getAssignedReq, err := http.NewRequest(http.MethodGet, "/monitors/"+strconv.Itoa(createdMonitor.ID)+"/tags", nil)
	if err != nil {
		t.Fatal(err)
	}
	getAssignedReq.Header.Set("x-api-key", apiKey)

	getAssignedResp, err := app.Test(getAssignedReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if getAssignedResp.StatusCode != http.StatusOK {
		t.Fatalf("expected monitor tags status %d, got %d", http.StatusOK, getAssignedResp.StatusCode)
	}
	defer getAssignedResp.Body.Close()

	var currentTags monitorroutes.MonitorTagsResponse
	if err := json.NewDecoder(getAssignedResp.Body).Decode(&currentTags); err != nil {
		t.Fatal(err)
	}
	if len(currentTags.Tags) != 1 || currentTags.Tags[0].Name != "production" {
		t.Fatalf("expected monitor tag, got %+v", currentTags.Tags)
	}

	updateTagBody := []byte(`{"name":"critical","description":""}`)
	updateTagReq, err := http.NewRequest(http.MethodPatch, "/tags/"+strconv.Itoa(createdTag.ID), bytes.NewReader(updateTagBody))
	if err != nil {
		t.Fatal(err)
	}
	updateTagReq.Header.Set("Content-Type", "application/json")
	updateTagReq.Header.Set("x-api-key", apiKey)

	updateTagResp, err := app.Test(updateTagReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if updateTagResp.StatusCode != http.StatusOK {
		t.Fatalf("expected update tag status %d, got %d", http.StatusOK, updateTagResp.StatusCode)
	}
	defer updateTagResp.Body.Close()

	var updatedTag tagroutes.TagResponse
	if err := json.NewDecoder(updateTagResp.Body).Decode(&updatedTag); err != nil {
		t.Fatal(err)
	}
	if updatedTag.Name != "critical" || updatedTag.Description != nil {
		t.Fatalf("unexpected updated tag response: %+v", updatedTag)
	}

	deleteTagReq, err := http.NewRequest(http.MethodDelete, "/tags/"+strconv.Itoa(createdTag.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	deleteTagReq.Header.Set("x-api-key", apiKey)

	deleteTagResp, err := app.Test(deleteTagReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if deleteTagResp.StatusCode != http.StatusOK {
		t.Fatalf("expected delete tag status %d, got %d", http.StatusOK, deleteTagResp.StatusCode)
	}
}

func TestNotificationChannelRoutesCRUD(t *testing.T) {
	cleanup := useTestDatabase(t)
	defer cleanup()

	app := New()
	apiKey := signupAndLogin(t, app, "channeluser", "channel@example.com", "examplepassword")

	createBody := []byte(`{"name":"Primary SMTP","type":"smtp","is_default":true,"smtp_server":"smtp.example.com","smtp_port":587,"smtp_from_address":"alerts@example.com","smtp_username":"alerts@example.com","smtp_password":"secret","smtp_use_tls":true}`)
	createReq, err := http.NewRequest(http.MethodPost, "/notification-channels", bytes.NewReader(createBody))
	if err != nil {
		t.Fatal(err)
	}
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("x-api-key", apiKey)

	createResp, err := app.Test(createReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create notification channel status %d, got %d", http.StatusCreated, createResp.StatusCode)
	}
	defer createResp.Body.Close()

	var created notificationchannelroutes.NotificationChannelResponse
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.Type != "smtp" || !created.IsDefault || !created.HasSMTPPassword {
		t.Fatalf("unexpected created notification channel response: %+v", created)
	}

	createSecondBody := []byte(`{"name":"Secondary SMTP","type":"smtp","is_default":true,"smtp_server":"smtp2.example.com","smtp_port":587,"smtp_from_address":"alerts2@example.com"}`)
	createSecondReq, err := http.NewRequest(http.MethodPost, "/notification-channels", bytes.NewReader(createSecondBody))
	if err != nil {
		t.Fatal(err)
	}
	createSecondReq.Header.Set("Content-Type", "application/json")
	createSecondReq.Header.Set("x-api-key", apiKey)

	createSecondResp, err := app.Test(createSecondReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if createSecondResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected second create notification channel status %d, got %d", http.StatusCreated, createSecondResp.StatusCode)
	}
	defer createSecondResp.Body.Close()

	var second notificationchannelroutes.NotificationChannelResponse
	if err := json.NewDecoder(createSecondResp.Body).Decode(&second); err != nil {
		t.Fatal(err)
	}
	if !second.IsDefault {
		t.Fatalf("expected second channel to be default, got %+v", second)
	}

	getFirstReq, err := http.NewRequest(http.MethodGet, "/notification-channels/"+strconv.Itoa(created.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	getFirstReq.Header.Set("x-api-key", apiKey)

	getFirstResp, err := app.Test(getFirstReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if getFirstResp.StatusCode != http.StatusOK {
		t.Fatalf("expected get notification channel status %d, got %d", http.StatusOK, getFirstResp.StatusCode)
	}
	defer getFirstResp.Body.Close()

	var firstAfterSecond notificationchannelroutes.NotificationChannelResponse
	if err := json.NewDecoder(getFirstResp.Body).Decode(&firstAfterSecond); err != nil {
		t.Fatal(err)
	}
	if firstAfterSecond.IsDefault {
		t.Fatal("expected first SMTP channel default flag to be cleared")
	}

	updateBody := []byte(`{"name":"Secondary SMTP Updated","is_active":false,"smtp_password":""}`)
	updateReq, err := http.NewRequest(http.MethodPatch, "/notification-channels/"+strconv.Itoa(second.ID), bytes.NewReader(updateBody))
	if err != nil {
		t.Fatal(err)
	}
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("x-api-key", apiKey)

	updateResp, err := app.Test(updateReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("expected update notification channel status %d, got %d", http.StatusOK, updateResp.StatusCode)
	}
	defer updateResp.Body.Close()

	var updated notificationchannelroutes.NotificationChannelResponse
	if err := json.NewDecoder(updateResp.Body).Decode(&updated); err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Secondary SMTP Updated" || updated.IsActive || updated.HasSMTPPassword {
		t.Fatalf("unexpected updated notification channel response: %+v", updated)
	}

	listReq, err := http.NewRequest(http.MethodGet, "/notification-channels?type=smtp", nil)
	if err != nil {
		t.Fatal(err)
	}
	listReq.Header.Set("x-api-key", apiKey)

	listResp, err := app.Test(listReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("expected list notification channels status %d, got %d", http.StatusOK, listResp.StatusCode)
	}
	defer listResp.Body.Close()

	var list notificationchannelroutes.NotificationChannelListResponse
	if err := json.NewDecoder(listResp.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list.NotificationChannels) != 2 {
		t.Fatalf("expected 2 SMTP notification channels, got %+v", list.NotificationChannels)
	}

	deleteReq, err := http.NewRequest(http.MethodDelete, "/notification-channels/"+strconv.Itoa(second.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	deleteReq.Header.Set("x-api-key", apiKey)

	deleteResp, err := app.Test(deleteReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("expected delete notification channel status %d, got %d", http.StatusOK, deleteResp.StatusCode)
	}
}

func useTestDatabase(t *testing.T) func() {
	t.Helper()

	db, err := stdsql.Open("sqlite", "file:phoenix_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatal(err)
	}

	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	if err := client.Schema.Create(context.Background()); err != nil {
		t.Fatal(err)
	}

	previous := database.Client
	database.Client = client
	return func() {
		database.Client = previous
		if err := client.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func signupAndLogin(t *testing.T, app *fiber.App, username string, email string, password string) string {
	t.Helper()

	signupBody, err := json.Marshal(map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	})
	if err != nil {
		t.Fatal(err)
	}
	signupReq, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewReader(signupBody))
	if err != nil {
		t.Fatal(err)
	}
	signupReq.Header.Set("Content-Type", "application/json")

	signupResp, err := app.Test(signupReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if signupResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected signup status %d, got %d", http.StatusCreated, signupResp.StatusCode)
	}

	loginBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		t.Fatal(err)
	}
	loginReq, err := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatal(err)
	}
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := app.Test(loginReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("expected login status %d, got %d", http.StatusOK, loginResp.StatusCode)
	}
	defer loginResp.Body.Close()

	var login authroutes.LoginResponse
	if err := json.NewDecoder(loginResp.Body).Decode(&login); err != nil {
		t.Fatal(err)
	}
	if login.ApiKey == "" {
		t.Fatal("expected login response to include api key")
	}

	return login.ApiKey
}
