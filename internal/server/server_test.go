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
	alertruleroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/alert_rules"
	authroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/auth"
	incidentroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/incidents"
	maintenancewindowroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/maintenance_windows"
	monitorroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/monitors"
	notificationchannelroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/notification_channels"
	statuspageroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/status_pages"
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

func TestAlertRuleRoutesCRUD(t *testing.T) {
	cleanup := useTestDatabase(t)
	defer cleanup()

	app := New()
	apiKey := signupAndLogin(t, app, "alertuser", "alert@example.com", "examplepassword")

	tagID := createTestTag(t, app, apiKey, "production")
	monitorID := createTestMonitor(t, app, apiKey, "Alert Website")
	channelID := createTestNotificationChannel(t, app, apiKey)

	createBody, err := json.Marshal(map[string]any{
		"name":                     "Production down",
		"description":              "Notify when production tag goes down",
		"event":                    "down",
		"scope":                    "tags",
		"tag_ids":                  []int{tagID},
		"notification_channel_ids": []int{channelID},
	})
	if err != nil {
		t.Fatal(err)
	}
	createReq, err := http.NewRequest(http.MethodPost, "/alert-rules", bytes.NewReader(createBody))
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
		t.Fatalf("expected create alert rule status %d, got %d", http.StatusCreated, createResp.StatusCode)
	}
	defer createResp.Body.Close()

	var created alertruleroutes.AlertRuleResponse
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.Scope != "tags" || len(created.TagIDs) != 1 || len(created.NotificationChannelIDs) != 1 {
		t.Fatalf("unexpected created alert rule response: %+v", created)
	}

	listReq, err := http.NewRequest(http.MethodGet, "/alert-rules?scope=tags", nil)
	if err != nil {
		t.Fatal(err)
	}
	listReq.Header.Set("x-api-key", apiKey)

	listResp, err := app.Test(listReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("expected list alert rules status %d, got %d", http.StatusOK, listResp.StatusCode)
	}
	defer listResp.Body.Close()

	var list alertruleroutes.AlertRuleListResponse
	if err := json.NewDecoder(listResp.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list.AlertRules) != 1 || list.AlertRules[0].ID != created.ID {
		t.Fatalf("expected created alert rule in list, got %+v", list.AlertRules)
	}

	updateBody, err := json.Marshal(map[string]any{
		"name":                    "Website recovered",
		"event":                   "recovered",
		"scope":                   "monitors",
		"monitor_ids":             []int{monitorID},
		"resend_interval_seconds": 300,
	})
	if err != nil {
		t.Fatal(err)
	}
	updateReq, err := http.NewRequest(http.MethodPatch, "/alert-rules/"+strconv.Itoa(created.ID), bytes.NewReader(updateBody))
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
		t.Fatalf("expected update alert rule status %d, got %d", http.StatusOK, updateResp.StatusCode)
	}
	defer updateResp.Body.Close()

	var updated alertruleroutes.AlertRuleResponse
	if err := json.NewDecoder(updateResp.Body).Decode(&updated); err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Website recovered" || updated.Event != "recovered" || updated.Scope != "monitors" || len(updated.MonitorIDs) != 1 || len(updated.TagIDs) != 0 {
		t.Fatalf("unexpected updated alert rule response: %+v", updated)
	}
	if updated.ResendIntervalSeconds == nil || *updated.ResendIntervalSeconds != 300 {
		t.Fatalf("expected resend interval to be set, got %+v", updated.ResendIntervalSeconds)
	}

	deleteReq, err := http.NewRequest(http.MethodDelete, "/alert-rules/"+strconv.Itoa(created.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	deleteReq.Header.Set("x-api-key", apiKey)

	deleteResp, err := app.Test(deleteReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("expected delete alert rule status %d, got %d", http.StatusOK, deleteResp.StatusCode)
	}
}

func TestIncidentRoutesCRUDAndState(t *testing.T) {
	cleanup := useTestDatabase(t)
	defer cleanup()

	app := New()
	apiKey := signupAndLogin(t, app, "incidentuser", "incident@example.com", "examplepassword")
	monitorID := createTestMonitor(t, app, apiKey, "Incident Website")

	createBody, err := json.Marshal(map[string]any{
		"monitor_id": monitorID,
		"title":      "Website outage",
		"content":    "The website is returning 500s",
		"severity":   "critical",
	})
	if err != nil {
		t.Fatal(err)
	}
	createReq, err := http.NewRequest(http.MethodPost, "/incidents", bytes.NewReader(createBody))
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
		t.Fatalf("expected create incident status %d, got %d", http.StatusCreated, createResp.StatusCode)
	}
	defer createResp.Body.Close()

	var created incidentroutes.IncidentResponse
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.MonitorID != monitorID || created.Status != "open" || created.Severity != "critical" {
		t.Fatalf("unexpected created incident response: %+v", created)
	}

	listReq, err := http.NewRequest(http.MethodGet, "/incidents?status=open&monitor_id="+strconv.Itoa(monitorID), nil)
	if err != nil {
		t.Fatal(err)
	}
	listReq.Header.Set("x-api-key", apiKey)

	listResp, err := app.Test(listReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("expected list incidents status %d, got %d", http.StatusOK, listResp.StatusCode)
	}
	defer listResp.Body.Close()

	var list incidentroutes.IncidentListResponse
	if err := json.NewDecoder(listResp.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list.Incidents) != 1 || list.Incidents[0].ID != created.ID {
		t.Fatalf("expected created incident in list, got %+v", list.Incidents)
	}

	updateBody := []byte(`{"title":"Website outage updated","severity":"warning","is_pinned":false}`)
	updateReq, err := http.NewRequest(http.MethodPatch, "/incidents/"+strconv.Itoa(created.ID), bytes.NewReader(updateBody))
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
		t.Fatalf("expected update incident status %d, got %d", http.StatusOK, updateResp.StatusCode)
	}
	defer updateResp.Body.Close()

	var updated incidentroutes.IncidentResponse
	if err := json.NewDecoder(updateResp.Body).Decode(&updated); err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Website outage updated" || updated.Severity != "warning" || updated.IsPinned {
		t.Fatalf("unexpected updated incident response: %+v", updated)
	}

	ackReq, err := http.NewRequest(http.MethodPost, "/incidents/"+strconv.Itoa(created.ID)+"/acknowledge", nil)
	if err != nil {
		t.Fatal(err)
	}
	ackReq.Header.Set("x-api-key", apiKey)

	ackResp, err := app.Test(ackReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if ackResp.StatusCode != http.StatusOK {
		t.Fatalf("expected acknowledge incident status %d, got %d", http.StatusOK, ackResp.StatusCode)
	}
	defer ackResp.Body.Close()

	var acknowledged incidentroutes.IncidentResponse
	if err := json.NewDecoder(ackResp.Body).Decode(&acknowledged); err != nil {
		t.Fatal(err)
	}
	if acknowledged.Status != "acknowledged" || acknowledged.EndedAt != nil || acknowledged.ResolvedByID != nil {
		t.Fatalf("unexpected acknowledged incident response: %+v", acknowledged)
	}

	resolveReq, err := http.NewRequest(http.MethodPost, "/incidents/"+strconv.Itoa(created.ID)+"/resolve", nil)
	if err != nil {
		t.Fatal(err)
	}
	resolveReq.Header.Set("x-api-key", apiKey)

	resolveResp, err := app.Test(resolveReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if resolveResp.StatusCode != http.StatusOK {
		t.Fatalf("expected resolve incident status %d, got %d", http.StatusOK, resolveResp.StatusCode)
	}
	defer resolveResp.Body.Close()

	var resolved incidentroutes.IncidentResponse
	if err := json.NewDecoder(resolveResp.Body).Decode(&resolved); err != nil {
		t.Fatal(err)
	}
	if resolved.Status != "resolved" || resolved.EndedAt == nil || resolved.ResolvedByID == nil {
		t.Fatalf("unexpected resolved incident response: %+v", resolved)
	}

	deleteReq, err := http.NewRequest(http.MethodDelete, "/incidents/"+strconv.Itoa(created.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	deleteReq.Header.Set("x-api-key", apiKey)

	deleteResp, err := app.Test(deleteReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("expected delete incident status %d, got %d", http.StatusOK, deleteResp.StatusCode)
	}
}

func TestMaintenanceWindowRoutesCRUDAndState(t *testing.T) {
	cleanup := useTestDatabase(t)
	defer cleanup()

	app := New()
	apiKey := signupAndLogin(t, app, "maintenanceuser", "maintenance@example.com", "examplepassword")
	monitorID := createTestMonitor(t, app, apiKey, "Maintenance Website")
	startAt := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	endAt := startAt.Add(time.Hour)

	createBody, err := json.Marshal(map[string]any{
		"title":       "Planned maintenance",
		"description": "Database upgrade",
		"strategy":    "single",
		"start_at":    startAt.Format(time.RFC3339),
		"end_at":      endAt.Format(time.RFC3339),
		"monitor_ids": []int{monitorID},
	})
	if err != nil {
		t.Fatal(err)
	}
	createReq, err := http.NewRequest(http.MethodPost, "/maintenance-windows", bytes.NewReader(createBody))
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
		t.Fatalf("expected create maintenance window status %d, got %d", http.StatusCreated, createResp.StatusCode)
	}
	defer createResp.Body.Close()

	var created maintenancewindowroutes.MaintenanceWindowResponse
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.Strategy != "single" || len(created.MonitorIDs) != 1 || created.MonitorIDs[0] != monitorID {
		t.Fatalf("unexpected created maintenance window response: %+v", created)
	}

	listReq, err := http.NewRequest(http.MethodGet, "/maintenance-windows?strategy=single&monitor_id="+strconv.Itoa(monitorID), nil)
	if err != nil {
		t.Fatal(err)
	}
	listReq.Header.Set("x-api-key", apiKey)

	listResp, err := app.Test(listReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("expected list maintenance windows status %d, got %d", http.StatusOK, listResp.StatusCode)
	}
	defer listResp.Body.Close()

	var list maintenancewindowroutes.MaintenanceWindowListResponse
	if err := json.NewDecoder(listResp.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list.MaintenanceWindows) != 1 || list.MaintenanceWindows[0].ID != created.ID {
		t.Fatalf("expected created maintenance window in list, got %+v", list.MaintenanceWindows)
	}

	updateBody, err := json.Marshal(map[string]any{
		"title":            "Recurring maintenance",
		"description":      "",
		"strategy":         "recurring",
		"duration_seconds": 1800,
		"monitor_ids":      []int{},
	})
	if err != nil {
		t.Fatal(err)
	}
	updateReq, err := http.NewRequest(http.MethodPatch, "/maintenance-windows/"+strconv.Itoa(created.ID), bytes.NewReader(updateBody))
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
		t.Fatalf("expected update maintenance window status %d, got %d", http.StatusOK, updateResp.StatusCode)
	}
	defer updateResp.Body.Close()

	var updated maintenancewindowroutes.MaintenanceWindowResponse
	if err := json.NewDecoder(updateResp.Body).Decode(&updated); err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Recurring maintenance" || updated.Strategy != "recurring" || updated.Description != nil || len(updated.MonitorIDs) != 0 {
		t.Fatalf("unexpected updated maintenance window response: %+v", updated)
	}
	if updated.DurationSeconds == nil || *updated.DurationSeconds != 1800 {
		t.Fatalf("expected duration seconds to be updated, got %+v", updated.DurationSeconds)
	}

	deactivateReq, err := http.NewRequest(http.MethodPost, "/maintenance-windows/"+strconv.Itoa(created.ID)+"/deactivate", nil)
	if err != nil {
		t.Fatal(err)
	}
	deactivateReq.Header.Set("x-api-key", apiKey)

	deactivateResp, err := app.Test(deactivateReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if deactivateResp.StatusCode != http.StatusOK {
		t.Fatalf("expected deactivate maintenance window status %d, got %d", http.StatusOK, deactivateResp.StatusCode)
	}
	defer deactivateResp.Body.Close()

	var deactivated maintenancewindowroutes.MaintenanceWindowResponse
	if err := json.NewDecoder(deactivateResp.Body).Decode(&deactivated); err != nil {
		t.Fatal(err)
	}
	if deactivated.IsActive {
		t.Fatalf("expected deactivated maintenance window, got %+v", deactivated)
	}

	activateReq, err := http.NewRequest(http.MethodPost, "/maintenance-windows/"+strconv.Itoa(created.ID)+"/activate", nil)
	if err != nil {
		t.Fatal(err)
	}
	activateReq.Header.Set("x-api-key", apiKey)

	activateResp, err := app.Test(activateReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if activateResp.StatusCode != http.StatusOK {
		t.Fatalf("expected activate maintenance window status %d, got %d", http.StatusOK, activateResp.StatusCode)
	}
	defer activateResp.Body.Close()

	var activated maintenancewindowroutes.MaintenanceWindowResponse
	if err := json.NewDecoder(activateResp.Body).Decode(&activated); err != nil {
		t.Fatal(err)
	}
	if !activated.IsActive {
		t.Fatalf("expected activated maintenance window, got %+v", activated)
	}

	deleteReq, err := http.NewRequest(http.MethodDelete, "/maintenance-windows/"+strconv.Itoa(created.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	deleteReq.Header.Set("x-api-key", apiKey)

	deleteResp, err := app.Test(deleteReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("expected delete maintenance window status %d, got %d", http.StatusOK, deleteResp.StatusCode)
	}
}

func TestStatusPageRoutesCRUDAndMonitorAssignment(t *testing.T) {
	cleanup := useTestDatabase(t)
	defer cleanup()

	app := New()
	apiKey := signupAndLogin(t, app, "statuspageuser", "statuspage@example.com", "examplepassword")
	monitorID := createTestMonitor(t, app, apiKey, "Status Page Website")

	createBody, err := json.Marshal(map[string]any{
		"slug":                   "main-status",
		"name":                   "Main Status",
		"description":            "Public service status",
		"is_public":              true,
		"password":               "secret",
		"show_tags":              true,
		"show_charts":            false,
		"show_uptime_percentage": true,
		"show_powered_by":        false,
		"auto_refresh_interval":  120,
	})
	if err != nil {
		t.Fatal(err)
	}
	createReq, err := http.NewRequest(http.MethodPost, "/status-pages", bytes.NewReader(createBody))
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
		t.Fatalf("expected create status page status %d, got %d", http.StatusCreated, createResp.StatusCode)
	}
	defer createResp.Body.Close()

	var created statuspageroutes.StatusPageResponse
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.Slug != "main-status" || created.Name != "Main Status" || !created.HasPassword || !created.ShowTags || created.ShowCharts || created.ShowPoweredBy {
		t.Fatalf("unexpected created status page response: %+v", created)
	}

	replaceBody, err := json.Marshal(map[string]any{
		"monitors": []map[string]any{
			{
				"monitor_id":   monitorID,
				"display_name": "Website",
				"weight":       10,
				"send_url":     true,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	replaceReq, err := http.NewRequest(http.MethodPut, "/status-pages/"+strconv.Itoa(created.ID)+"/monitors", bytes.NewReader(replaceBody))
	if err != nil {
		t.Fatal(err)
	}
	replaceReq.Header.Set("Content-Type", "application/json")
	replaceReq.Header.Set("x-api-key", apiKey)

	replaceResp, err := app.Test(replaceReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if replaceResp.StatusCode != http.StatusOK {
		t.Fatalf("expected replace status page monitors status %d, got %d", http.StatusOK, replaceResp.StatusCode)
	}
	defer replaceResp.Body.Close()

	var replaced statuspageroutes.StatusPageMonitorsResponse
	if err := json.NewDecoder(replaceResp.Body).Decode(&replaced); err != nil {
		t.Fatal(err)
	}
	if len(replaced.Monitors) != 1 || replaced.Monitors[0].MonitorID != monitorID || replaced.Monitors[0].DisplayName == nil || *replaced.Monitors[0].DisplayName != "Website" || replaced.Monitors[0].Weight != 10 || !replaced.Monitors[0].SendURL {
		t.Fatalf("unexpected replaced status page monitors response: %+v", replaced.Monitors)
	}

	monitorsReq, err := http.NewRequest(http.MethodGet, "/status-pages/"+strconv.Itoa(created.ID)+"/monitors", nil)
	if err != nil {
		t.Fatal(err)
	}
	monitorsReq.Header.Set("x-api-key", apiKey)

	monitorsResp, err := app.Test(monitorsReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if monitorsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected list status page monitors status %d, got %d", http.StatusOK, monitorsResp.StatusCode)
	}
	defer monitorsResp.Body.Close()

	var monitors statuspageroutes.StatusPageMonitorsResponse
	if err := json.NewDecoder(monitorsResp.Body).Decode(&monitors); err != nil {
		t.Fatal(err)
	}
	if len(monitors.Monitors) != 1 || monitors.Monitors[0].MonitorID != monitorID {
		t.Fatalf("expected assigned monitor, got %+v", monitors.Monitors)
	}

	listReq, err := http.NewRequest(http.MethodGet, "/status-pages?is_public=true", nil)
	if err != nil {
		t.Fatal(err)
	}
	listReq.Header.Set("x-api-key", apiKey)

	listResp, err := app.Test(listReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("expected list status pages status %d, got %d", http.StatusOK, listResp.StatusCode)
	}
	defer listResp.Body.Close()

	var list statuspageroutes.StatusPageListResponse
	if err := json.NewDecoder(listResp.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list.StatusPages) != 1 || list.StatusPages[0].ID != created.ID || len(list.StatusPages[0].Monitors) != 1 {
		t.Fatalf("expected created status page in list, got %+v", list.StatusPages)
	}

	updateBody := []byte(`{"name":"Public Status","password":"","footer_text":"Operational details"}`)
	updateReq, err := http.NewRequest(http.MethodPatch, "/status-pages/"+strconv.Itoa(created.ID), bytes.NewReader(updateBody))
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
		t.Fatalf("expected update status page status %d, got %d", http.StatusOK, updateResp.StatusCode)
	}
	defer updateResp.Body.Close()

	var updated statuspageroutes.StatusPageResponse
	if err := json.NewDecoder(updateResp.Body).Decode(&updated); err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Public Status" || updated.HasPassword || updated.FooterText == nil || *updated.FooterText != "Operational details" {
		t.Fatalf("unexpected updated status page response: %+v", updated)
	}

	deleteReq, err := http.NewRequest(http.MethodDelete, "/status-pages/"+strconv.Itoa(created.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	deleteReq.Header.Set("x-api-key", apiKey)

	deleteResp, err := app.Test(deleteReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("expected delete status page status %d, got %d", http.StatusOK, deleteResp.StatusCode)
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

func createTestTag(t *testing.T, app *fiber.App, apiKey string, name string) int {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"name":  name,
		"color": "#ff0000",
	})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/tags", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create tag status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	defer resp.Body.Close()

	var created tagroutes.TagResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	return created.ID
}

func createTestMonitor(t *testing.T, app *fiber.App, apiKey string, name string) int {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"name": name,
		"type": "http",
		"url":  "https://example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/monitors", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create monitor status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	defer resp.Body.Close()

	var created monitorroutes.MonitorResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	return created.ID
}

func createTestNotificationChannel(t *testing.T, app *fiber.App, apiKey string) int {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"name":               "Alert Webhook",
		"type":               "webhook",
		"webhook_url":        "https://example.com/webhook",
		"webhook_method":     "POST",
		"notification_token": "ignored",
	})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/notification-channels", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create notification channel status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	defer resp.Body.Close()

	var created notificationchannelroutes.NotificationChannelResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	return created.ID
}
