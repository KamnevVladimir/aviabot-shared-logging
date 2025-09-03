package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KamnevVladimir/aviabot-shared-logging/application/usecases"
	"github.com/KamnevVladimir/aviabot-shared-logging/domain/entities"
	domainerrors "github.com/KamnevVladimir/aviabot-shared-logging/domain/errors"
	"github.com/KamnevVladimir/aviabot-shared-logging/domain/interfaces"
	infrahttp "github.com/KamnevVladimir/aviabot-shared-logging/infrastructure/http"
)

// Mock Repository для Integration тестов
type mockLogRepository struct {
	logs     []entities.LogEntry
	statsMap map[string]*interfaces.LogStats
}

func newMockLogRepository() *mockLogRepository {
	return &mockLogRepository{
		logs:     make([]entities.LogEntry, 0),
		statsMap: make(map[string]*interfaces.LogStats),
	}
}

func (m *mockLogRepository) Store(ctx context.Context, logEntry entities.LogEntry) error {
	if !logEntry.IsValid() {
		return domainerrors.ErrInvalidLogEntry
	}
	m.logs = append(m.logs, logEntry)
	return nil
}

func (m *mockLogRepository) GetByID(ctx context.Context, id string) (*entities.LogEntry, error) {
	for _, log := range m.logs {
		if log.ID == id {
			return &log, nil
		}
	}
	return nil, domainerrors.ErrLogNotFound
}

func (m *mockLogRepository) Query(ctx context.Context, filter interfaces.LogFilter) ([]entities.LogEntry, error) {
	result := make([]entities.LogEntry, 0)

	for _, log := range m.logs {
		if m.matchesFilter(log, filter) {
			result = append(result, log)
		}
	}

	// Применяем limit и offset
	start := filter.Offset
	if start > len(result) {
		start = len(result)
	}

	end := start + filter.Limit
	if filter.Limit == 0 {
		end = len(result)
	}
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], nil
}

func (m *mockLogRepository) Count(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
	count := int64(0)
	for _, log := range m.logs {
		if m.matchesFilter(log, filter) {
			count++
		}
	}
	return count, nil
}

func (m *mockLogRepository) GetStats(ctx context.Context, filter interfaces.LogFilter) (*interfaces.LogStats, error) {
	stats := &interfaces.LogStats{
		TotalCount:     0,
		CountByLevel:   make(map[entities.LogLevel]int64),
		CountByService: make(map[string]int64),
		CountByEvent:   make(map[string]int64),
	}

	for _, log := range m.logs {
		if m.matchesFilter(log, filter) {
			stats.TotalCount++
			stats.CountByLevel[log.Level]++
			stats.CountByService[log.Service]++
			stats.CountByEvent[log.Event]++
		}
	}

	return stats, nil
}

func (m *mockLogRepository) Delete(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
	deletedCount := int64(0)
	newLogs := make([]entities.LogEntry, 0)

	for _, log := range m.logs {
		if m.matchesFilter(log, filter) {
			deletedCount++
		} else {
			newLogs = append(newLogs, log)
		}
	}

	m.logs = newLogs
	return deletedCount, nil
}

func (m *mockLogRepository) matchesFilter(log entities.LogEntry, filter interfaces.LogFilter) bool {
	// Проверка сервисов
	if len(filter.Services) > 0 {
		found := false
		for _, service := range filter.Services {
			if log.Service == service {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Проверка событий
	if len(filter.Events) > 0 {
		found := false
		for _, event := range filter.Events {
			if log.Event == event {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Проверка уровней
	if len(filter.Levels) > 0 {
		found := false
		for _, level := range filter.Levels {
			if log.Level == level {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Проверка временного диапазона
	if filter.TimeFrom != nil && log.Timestamp.Before(*filter.TimeFrom) {
		return false
	}
	if filter.TimeTo != nil && log.Timestamp.After(*filter.TimeTo) {
		return false
	}

	// Проверка пользователя
	if filter.UserID != nil {
		if log.UserID == nil || *log.UserID != *filter.UserID {
			return false
		}
	}

	// Проверка чата
	if filter.ChatID != nil {
		if log.ChatID == nil || *log.ChatID != *filter.ChatID {
			return false
		}
	}

	return true
}

// Mock AlertService для Integration тестов
type mockAlertService struct {
	sentAlerts []entities.LogEntry
	healthy    bool
}

func newMockAlertService() *mockAlertService {
	return &mockAlertService{
		sentAlerts: make([]entities.LogEntry, 0),
		healthy:    true,
	}
}

func (m *mockAlertService) SendAlert(ctx context.Context, logEntry entities.LogEntry) error {
	if !m.healthy {
		return domainerrors.ErrAlertServiceUnavailable
	}
	m.sentAlerts = append(m.sentAlerts, logEntry)
	return nil
}

func (m *mockAlertService) SendBatchAlert(ctx context.Context, entries []entities.LogEntry) error {
	if !m.healthy {
		return domainerrors.ErrAlertServiceUnavailable
	}
	m.sentAlerts = append(m.sentAlerts, entries...)
	return nil
}

func (m *mockAlertService) IsHealthy(ctx context.Context) bool {
	return m.healthy
}

// Mock ID Generator для Integration тестов
type mockIDGenerator struct {
	counter int
}

func newMockIDGenerator() *mockIDGenerator {
	return &mockIDGenerator{counter: 0}
}

func (m *mockIDGenerator) Generate() string {
	m.counter++
	return fmt.Sprintf("log-%d", m.counter)
}

// Mock Time Provider для Integration тестов
type mockTimeProvider struct {
	fixedTime time.Time
}

func newMockTimeProvider() *mockTimeProvider {
	return &mockTimeProvider{
		fixedTime: time.Date(2025, 9, 1, 15, 30, 0, 0, time.UTC),
	}
}

func (m *mockTimeProvider) Now() time.Time {
	return m.fixedTime
}

// TestFullIntegration_LogEventFlow тестирует полный поток создания лог записи
func TestFullIntegration_LogEventFlow(t *testing.T) {
	// Setup dependencies
	repo := newMockLogRepository()
	alertService := newMockAlertService()
	idGenerator := newMockIDGenerator()
	timeProvider := newMockTimeProvider()

	// Create use cases
	logEventUseCase := usecases.NewLogEventUseCase(repo, alertService, idGenerator, timeProvider)
	queryLogsUseCase := usecases.NewQueryLogsUseCase(repo)
	getLogStatsUseCase := usecases.NewGetLogStatsUseCase(repo)

	// Create HTTP handlers
	logsHandler := infrahttp.NewLogsHandler(logEventUseCase, queryLogsUseCase, getLogStatsUseCase)

	tests := []struct {
		name           string
		logData        map[string]interface{}
		expectedStatus int
		shouldAlert    bool
	}{
		{
			name: "successful info log creation",
			logData: map[string]interface{}{
				"level":   "INFO",
				"service": "gateway-service",
				"event":   "update_received",
				"message": "Update processed successfully",
				"user_id": 12345,
				"chat_id": 67890,
				"metadata": map[string]interface{}{
					"duration_ms": 150,
					"status":      "success",
				},
			},
			expectedStatus: http.StatusCreated,
			shouldAlert:    false,
		},
		{
			name: "successful error log with alert",
			logData: map[string]interface{}{
				"level":   "ERROR",
				"service": "search-service",
				"event":   "api_error",
				"message": "Failed to process search request",
				"metadata": map[string]interface{}{
					"error_code": 500,
					"endpoint":   "/search",
				},
			},
			expectedStatus: http.StatusCreated,
			shouldAlert:    true,
		},
		{
			name: "successful critical log with alert",
			logData: map[string]interface{}{
				"level":   "CRITICAL",
				"service": "database-service",
				"event":   "connection_failure",
				"message": "Database connection lost",
			},
			expectedStatus: http.StatusCreated,
			shouldAlert:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare HTTP request
			jsonData, _ := json.Marshal(tt.logData)
			req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewReader(jsonData))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			// Execute HTTP handler
			logsHandler.CreateLog(recorder, req)

			// Verify HTTP response
			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				// Parse response
				var response map[string]interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}

				// Verify response structure
				if response["success"] != true {
					t.Errorf("Expected success=true, got %v", response["success"])
				}

				// Verify alert was sent if expected
				if tt.shouldAlert {
					if response["alert_sent"] != true {
						t.Errorf("Expected alert_sent=true for %s level", tt.logData["level"])
					}
					if len(alertService.sentAlerts) == 0 {
						t.Errorf("Expected alert to be sent, but none were sent")
					}
				} else {
					if response["alert_sent"] == true {
						t.Errorf("Expected alert_sent=false for %s level", tt.logData["level"])
					}
				}

				// Verify log was stored in repository
				if len(repo.logs) == 0 {
					t.Errorf("Expected log to be stored in repository")
				} else {
					storedLog := repo.logs[len(repo.logs)-1]
					if storedLog.Service != tt.logData["service"] {
						t.Errorf("Expected service %v, got %v", tt.logData["service"], storedLog.Service)
					}
					if storedLog.Event != tt.logData["event"] {
						t.Errorf("Expected event %v, got %v", tt.logData["event"], storedLog.Event)
					}
					if storedLog.Message != tt.logData["message"] {
						t.Errorf("Expected message %v, got %v", tt.logData["message"], storedLog.Message)
					}
				}
			}
		})
	}
}

// TestFullIntegration_QueryLogsFlow тестирует полный поток запроса логов
func TestFullIntegration_QueryLogsFlow(t *testing.T) {
	// Setup dependencies with pre-populated data
	repo := newMockLogRepository()
	alertService := newMockAlertService()
	idGenerator := newMockIDGenerator()
	timeProvider := newMockTimeProvider()

	// Pre-populate repository with test data
	testLogs := []entities.LogEntry{
		{
			ID:        "log-1",
			Level:     entities.LogLevelInfo,
			Service:   "gateway-service",
			Event:     "update_received",
			Timestamp: timeProvider.Now(),
			Message:   "Update processed",
			UserID:    int64Ptr(12345),
		},
		{
			ID:        "log-2",
			Level:     entities.LogLevelError,
			Service:   "search-service",
			Event:     "api_error",
			Timestamp: timeProvider.Now(),
			Message:   "Search failed",
			UserID:    int64Ptr(67890),
		},
		{
			ID:        "log-3",
			Level:     entities.LogLevelInfo,
			Service:   "gateway-service",
			Event:     "user_action",
			Timestamp: timeProvider.Now(),
			Message:   "User interaction",
			UserID:    int64Ptr(12345),
		},
	}

	for _, log := range testLogs {
		repo.Store(context.Background(), log)
	}

	// Create use cases
	logEventUseCase := usecases.NewLogEventUseCase(repo, alertService, idGenerator, timeProvider)
	queryLogsUseCase := usecases.NewQueryLogsUseCase(repo)
	getLogStatsUseCase := usecases.NewGetLogStatsUseCase(repo)

	// Create HTTP handlers
	logsHandler := infrahttp.NewLogsHandler(logEventUseCase, queryLogsUseCase, getLogStatsUseCase)

	tests := []struct {
		name           string
		queryParams    string
		expectedCount  int
		expectedStatus int
	}{
		{
			name:           "get all logs",
			queryParams:    "",
			expectedCount:  3,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "filter by service",
			queryParams:    "service=gateway-service",
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "filter by level",
			queryParams:    "level=ERROR",
			expectedCount:  1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "filter by user_id",
			queryParams:    "user_id=12345",
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "filter with limit",
			queryParams:    "limit=2",
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "filter with offset",
			queryParams:    "offset=1&limit=2",
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare HTTP request
			req := httptest.NewRequest(http.MethodGet, "/logs?"+tt.queryParams, nil)
			recorder := httptest.NewRecorder()

			// Execute HTTP handler
			logsHandler.GetLogs(recorder, req)

			// Verify HTTP response
			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", recorder.Code, tt.expectedStatus)
			}

			// Parse response
			var response map[string]interface{}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			// Verify logs count
			logs, ok := response["logs"].([]interface{})
			if !ok {
				t.Fatalf("Expected logs array in response")
			}

			if len(logs) != tt.expectedCount {
				t.Errorf("Expected %d logs, got %d", tt.expectedCount, len(logs))
			}

			// Verify total_count and has_more fields exist
			if _, exists := response["total_count"]; !exists {
				t.Errorf("Expected total_count field in response")
			}

			if _, exists := response["has_more"]; !exists {
				t.Errorf("Expected has_more field in response")
			}
		})
	}
}

// TestFullIntegration_StatsFlow тестирует полный поток получения статистики
func TestFullIntegration_StatsFlow(t *testing.T) {
	// Setup dependencies with pre-populated data
	repo := newMockLogRepository()
	alertService := newMockAlertService()
	idGenerator := newMockIDGenerator()
	timeProvider := newMockTimeProvider()

	// Pre-populate repository with varied test data
	testLogs := []entities.LogEntry{
		{ID: "log-1", Level: entities.LogLevelInfo, Service: "gateway-service", Event: "update_received", Timestamp: timeProvider.Now(), Message: "Test 1"},
		{ID: "log-2", Level: entities.LogLevelError, Service: "search-service", Event: "api_error", Timestamp: timeProvider.Now(), Message: "Test 2"},
		{ID: "log-3", Level: entities.LogLevelInfo, Service: "gateway-service", Event: "user_action", Timestamp: timeProvider.Now(), Message: "Test 3"},
		{ID: "log-4", Level: entities.LogLevelWarning, Service: "ui-service", Event: "slow_response", Timestamp: timeProvider.Now(), Message: "Test 4"},
		{ID: "log-5", Level: entities.LogLevelCritical, Service: "database-service", Event: "connection_failure", Timestamp: timeProvider.Now(), Message: "Test 5"},
	}

	for _, log := range testLogs {
		repo.Store(context.Background(), log)
	}

	// Create use cases
	logEventUseCase := usecases.NewLogEventUseCase(repo, alertService, idGenerator, timeProvider)
	queryLogsUseCase := usecases.NewQueryLogsUseCase(repo)
	getLogStatsUseCase := usecases.NewGetLogStatsUseCase(repo)

	// Create HTTP handlers
	logsHandler := infrahttp.NewLogsHandler(logEventUseCase, queryLogsUseCase, getLogStatsUseCase)

	// Prepare HTTP request
	req := httptest.NewRequest(http.MethodGet, "/logs/stats", nil)
	recorder := httptest.NewRecorder()

	// Execute HTTP handler
	logsHandler.GetStats(recorder, req)

	// Verify HTTP response
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify stats structure
	stats, ok := response["stats"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected stats object in response")
	}

	// Verify total count
	totalCount, ok := stats["total_count"].(float64)
	if !ok || totalCount != 5 {
		t.Errorf("Expected total_count=5, got %v", totalCount)
	}

	// Verify count_by_level exists and has correct structure
	if _, exists := stats["count_by_level"]; !exists {
		t.Errorf("Expected count_by_level field in stats")
	}

	// Verify count_by_service exists and has correct structure
	if _, exists := stats["count_by_service"]; !exists {
		t.Errorf("Expected count_by_service field in stats")
	}

	// Verify count_by_event exists and has correct structure
	if _, exists := stats["count_by_event"]; !exists {
		t.Errorf("Expected count_by_event field in stats")
	}
}

// TestFullIntegration_ErrorHandling тестирует обработку ошибок на всех слоях
func TestFullIntegration_ErrorHandling(t *testing.T) {
	// Setup dependencies
	repo := newMockLogRepository()
	alertService := newMockAlertService()
	idGenerator := newMockIDGenerator()
	timeProvider := newMockTimeProvider()

	// Create use cases
	logEventUseCase := usecases.NewLogEventUseCase(repo, alertService, idGenerator, timeProvider)
	queryLogsUseCase := usecases.NewQueryLogsUseCase(repo)
	getLogStatsUseCase := usecases.NewGetLogStatsUseCase(repo)

	// Create HTTP handlers
	logsHandler := infrahttp.NewLogsHandler(logEventUseCase, queryLogsUseCase, getLogStatsUseCase)

	tests := []struct {
		name           string
		method         string
		url            string
		body           interface{}
		expectedStatus int
		setupError     func()
	}{
		{
			name:           "invalid JSON body",
			method:         http.MethodPost,
			url:            "/logs",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid log level",
			method: http.MethodPost,
			url:    "/logs",
			body: map[string]interface{}{
				"level":   "INVALID_LEVEL",
				"service": "test-service",
				"event":   "test-event",
				"message": "test message",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "missing required fields",
			method: http.MethodPost,
			url:    "/logs",
			body: map[string]interface{}{
				"level": "INFO",
				// service, event, message отсутствуют
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid query parameter",
			method:         http.MethodGet,
			url:            "/logs?limit=invalid",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "method not allowed for create",
			method:         http.MethodPut,
			url:            "/logs",
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "method not allowed for query",
			method:         http.MethodPost,
			url:            "/logs?service=test",
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupError != nil {
				tt.setupError()
			}

			// Prepare request body
			var requestBody []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					requestBody = []byte(str)
				} else {
					requestBody, _ = json.Marshal(tt.body)
				}
			}

			// Create request
			var req *http.Request
			if requestBody != nil {
				req = httptest.NewRequest(tt.method, tt.url, bytes.NewReader(requestBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.url, nil)
			}

			recorder := httptest.NewRecorder()

			// Route to appropriate handler based on URL
			if tt.url == "/logs" {
				if tt.method == http.MethodPost {
					logsHandler.CreateLog(recorder, req)
				} else if tt.method == http.MethodGet {
					logsHandler.GetLogs(recorder, req)
				} else {
					// Для неподдерживаемых методов (PUT, DELETE, etc.) вызываем CreateLog
					// который корректно вернет 405 Method Not Allowed
					logsHandler.CreateLog(recorder, req)
				}
			} else if strings.Contains(tt.url, "/logs?") {
				logsHandler.GetLogs(recorder, req)
			} else if strings.Contains(tt.url, "/logs/stats") {
				logsHandler.GetStats(recorder, req)
			}

			// Verify response
			if recorder.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectedStatus, recorder.Code)
			}

			// For error responses, verify error structure
			if tt.expectedStatus >= 400 {
				var response map[string]interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Errorf("%s: failed to parse error response: %v", tt.name, err)
				} else {
					if response["success"] != false {
						t.Errorf("%s: expected success=false in error response", tt.name)
					}
					if _, exists := response["error"]; !exists {
						t.Errorf("%s: expected error field in error response", tt.name)
					}
				}
			}
		})
	}
}

// TestFullIntegration_AlertServiceUnavailable тестирует поведение при недоступности сервиса алертов
func TestFullIntegration_AlertServiceUnavailable(t *testing.T) {
	// Setup dependencies with unhealthy alert service
	repo := newMockLogRepository()
	alertService := newMockAlertService()
	alertService.healthy = false // Делаем сервис недоступным
	idGenerator := newMockIDGenerator()
	timeProvider := newMockTimeProvider()

	// Create use cases
	logEventUseCase := usecases.NewLogEventUseCase(repo, alertService, idGenerator, timeProvider)
	queryLogsUseCase := usecases.NewQueryLogsUseCase(repo)
	getLogStatsUseCase := usecases.NewGetLogStatsUseCase(repo)

	// Create HTTP handlers
	logsHandler := infrahttp.NewLogsHandler(logEventUseCase, queryLogsUseCase, getLogStatsUseCase)

	// Create a critical log that should trigger an alert
	logData := map[string]interface{}{
		"level":   "CRITICAL",
		"service": "test-service",
		"event":   "system_failure",
		"message": "Critical system failure occurred",
	}

	jsonData, _ := json.Marshal(logData)
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	// Execute
	logsHandler.CreateLog(recorder, req)

	// Verify that log was still created despite alert failure
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, recorder.Code)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify that log was created successfully
	if response["success"] != true {
		t.Errorf("Expected success=true despite alert failure")
	}

	// Verify that alert was not sent
	if response["alert_sent"] == true {
		t.Errorf("Expected alert_sent=false when alert service is unavailable")
	}

	// Verify that log was still stored in repository
	if len(repo.logs) != 1 {
		t.Errorf("Expected 1 log in repository, got %d", len(repo.logs))
	}
}

// Helper function
func int64Ptr(v int64) *int64 {
	return &v
}
