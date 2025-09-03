package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aviasales-shared-logging/application/usecases"
	"aviasales-shared-logging/domain/entities"
	domainerrors "aviasales-shared-logging/domain/errors"
	"aviasales-shared-logging/domain/interfaces"
)

// Mock use case implementations
type mockLogEventUseCase struct {
	executeFunc func(ctx context.Context, request usecases.LogEventRequest) (*usecases.LogEventResponse, error)
}

func (m *mockLogEventUseCase) Execute(ctx context.Context, request usecases.LogEventRequest) (*usecases.LogEventResponse, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, request)
	}
	return &usecases.LogEventResponse{
		ID:        "test-id-123",
		Timestamp: time.Now(),
		Success:   true,
	}, nil
}

type mockQueryLogsUseCase struct {
	executeFunc func(ctx context.Context, request usecases.QueryLogsRequest) (*usecases.QueryLogsResponse, error)
}

func (m *mockQueryLogsUseCase) Execute(ctx context.Context, request usecases.QueryLogsRequest) (*usecases.QueryLogsResponse, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, request)
	}
	return &usecases.QueryLogsResponse{
		Logs:       []entities.LogEntry{},
		TotalCount: 0,
		HasMore:    false,
	}, nil
}

type mockGetLogStatsUseCase struct {
	executeFunc func(ctx context.Context, request usecases.GetLogStatsRequest) (*usecases.GetLogStatsResponse, error)
}

func (m *mockGetLogStatsUseCase) Execute(ctx context.Context, request usecases.GetLogStatsRequest) (*usecases.GetLogStatsResponse, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, request)
	}
	return &usecases.GetLogStatsResponse{
		Stats: interfaces.LogStats{
			TotalCount:     0,
			CountByLevel:   map[entities.LogLevel]int64{},
			CountByService: map[string]int64{},
			CountByEvent:   map[string]int64{},
		},
	}, nil
}

// TestLogsHandler_CreateLog тестирует создание лог записи через HTTP API
func TestLogsHandler_CreateLog(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mockLogEventUseCase)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful log creation",
			requestBody: map[string]interface{}{
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
			setupMock: func(mock *mockLogEventUseCase) {
				mock.executeFunc = func(ctx context.Context, request usecases.LogEventRequest) (*usecases.LogEventResponse, error) {
					return &usecases.LogEventResponse{
						ID:        "log-123",
						Timestamp: time.Date(2025, 9, 1, 15, 30, 0, 0, time.UTC),
						Success:   true,
						AlertSent: false,
					}, nil
				}
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":         "log-123",
				"timestamp":  "2025-09-01T15:30:00Z",
				"success":    true,
				"alert_sent": false,
			},
		},
		{
			name: "successful log creation with alert",
			requestBody: map[string]interface{}{
				"level":   "ERROR",
				"service": "gateway-service",
				"event":   "update_error",
				"message": "Failed to process update",
			},
			setupMock: func(mock *mockLogEventUseCase) {
				mock.executeFunc = func(ctx context.Context, request usecases.LogEventRequest) (*usecases.LogEventResponse, error) {
					return &usecases.LogEventResponse{
						ID:        "log-error-456",
						Timestamp: time.Date(2025, 9, 1, 15, 35, 0, 0, time.UTC),
						Success:   true,
						AlertSent: true,
					}, nil
				}
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":         "log-error-456",
				"timestamp":  "2025-09-01T15:35:00Z",
				"success":    true,
				"alert_sent": true,
			},
		},
		{
			name:        "invalid JSON body",
			requestBody: "invalid json",
			setupMock: func(mock *mockLogEventUseCase) {
				// Mock не должен вызываться
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Invalid JSON format",
				"success": false,
			},
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"level": "INFO",
				// service, event, message отсутствуют
			},
			setupMock: func(mock *mockLogEventUseCase) {
				mock.executeFunc = func(ctx context.Context, request usecases.LogEventRequest) (*usecases.LogEventResponse, error) {
					return nil, domainerrors.ErrInvalidLogEntry
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Invalid log entry",
				"success": false,
			},
		},
		{
			name: "invalid log level",
			requestBody: map[string]interface{}{
				"level":   "INVALID_LEVEL",
				"service": "gateway-service",
				"event":   "update_received",
				"message": "Test message",
			},
			setupMock: func(mock *mockLogEventUseCase) {
				// Парсинг уровня произойдет в handler
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Invalid log level",
				"success": false,
			},
		},
		{
			name: "storage unavailable",
			requestBody: map[string]interface{}{
				"level":   "INFO",
				"service": "gateway-service",
				"event":   "update_received",
				"message": "Test message",
			},
			setupMock: func(mock *mockLogEventUseCase) {
				mock.executeFunc = func(ctx context.Context, request usecases.LogEventRequest) (*usecases.LogEventResponse, error) {
					return nil, domainerrors.ErrStorageUnavailable
				}
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody: map[string]interface{}{
				"error":   "Storage unavailable",
				"success": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockUseCase := &mockLogEventUseCase{}
			tt.setupMock(mockUseCase)

			handler := NewLogsHandler(mockUseCase, nil, nil)

			// Prepare request
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			// Execute
			handler.CreateLog(recorder, req)

			// Assertions
			if recorder.Code != tt.expectedStatus {
				t.Errorf("CreateLog() status code = %d, want %d", recorder.Code, tt.expectedStatus)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			for key, expectedValue := range tt.expectedBody {
				if response[key] != expectedValue {
					t.Errorf("CreateLog() response[%s] = %v, want %v", key, response[key], expectedValue)
				}
			}
		})
	}
}

// TestLogsHandler_GetLogs тестирует получение логов через HTTP API
func TestLogsHandler_GetLogs(t *testing.T) {
	fixedTime := time.Date(2025, 9, 1, 15, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*mockQueryLogsUseCase)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "successful query without filters",
			queryParams: "",
			setupMock: func(mock *mockQueryLogsUseCase) {
				mock.executeFunc = func(ctx context.Context, request usecases.QueryLogsRequest) (*usecases.QueryLogsResponse, error) {
					return &usecases.QueryLogsResponse{
						Logs: []entities.LogEntry{
							{
								ID:        "log-1",
								Level:     entities.LogLevelInfo,
								Service:   "gateway-service",
								Event:     "update_received",
								Timestamp: fixedTime,
								Message:   "Test message",
							},
						},
						TotalCount: 1,
						HasMore:    false,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"total_count": float64(1),
				"has_more":    false,
			},
		},
		{
			name:        "query with filters",
			queryParams: "service=gateway-service&level=INFO&limit=10&offset=0",
			setupMock: func(mock *mockQueryLogsUseCase) {
				mock.executeFunc = func(ctx context.Context, request usecases.QueryLogsRequest) (*usecases.QueryLogsResponse, error) {
					// Проверяем что фильтры правильно переданы
					if len(request.Filter.Services) != 1 || request.Filter.Services[0] != "gateway-service" {
						return nil, domainerrors.ErrInvalidFilter
					}
					if len(request.Filter.Levels) != 1 || request.Filter.Levels[0] != entities.LogLevelInfo {
						return nil, domainerrors.ErrInvalidFilter
					}
					if request.Filter.Limit != 10 || request.Filter.Offset != 0 {
						return nil, domainerrors.ErrInvalidFilter
					}

					return &usecases.QueryLogsResponse{
						Logs:       []entities.LogEntry{},
						TotalCount: 0,
						HasMore:    false,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"total_count": float64(0),
				"has_more":    false,
			},
		},
		{
			name:        "invalid limit parameter",
			queryParams: "limit=invalid",
			setupMock: func(mock *mockQueryLogsUseCase) {
				// Mock не должен вызываться
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Invalid limit parameter",
				"success": false,
			},
		},
		{
			name:        "storage unavailable",
			queryParams: "",
			setupMock: func(mock *mockQueryLogsUseCase) {
				mock.executeFunc = func(ctx context.Context, request usecases.QueryLogsRequest) (*usecases.QueryLogsResponse, error) {
					return nil, domainerrors.ErrStorageUnavailable
				}
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody: map[string]interface{}{
				"error":   "Storage unavailable",
				"success": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockUseCase := &mockQueryLogsUseCase{}
			tt.setupMock(mockUseCase)

			handler := NewLogsHandler(nil, mockUseCase, nil)

			// Prepare request
			req := httptest.NewRequest(http.MethodGet, "/logs?"+tt.queryParams, nil)
			recorder := httptest.NewRecorder()

			// Execute
			handler.GetLogs(recorder, req)

			// Assertions
			if recorder.Code != tt.expectedStatus {
				t.Errorf("GetLogs() status code = %d, want %d", recorder.Code, tt.expectedStatus)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			for key, expectedValue := range tt.expectedBody {
				if response[key] != expectedValue {
					t.Errorf("GetLogs() response[%s] = %v, want %v", key, response[key], expectedValue)
				}
			}
		})
	}
}

// TestLogsHandler_GetStats тестирует получение статистики через HTTP API
func TestLogsHandler_GetStats(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*mockGetLogStatsUseCase)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "successful stats retrieval",
			queryParams: "",
			setupMock: func(mock *mockGetLogStatsUseCase) {
				mock.executeFunc = func(ctx context.Context, request usecases.GetLogStatsRequest) (*usecases.GetLogStatsResponse, error) {
					return &usecases.GetLogStatsResponse{
						Stats: interfaces.LogStats{
							TotalCount: 100,
							CountByLevel: map[entities.LogLevel]int64{
								entities.LogLevelInfo:    60,
								entities.LogLevelWarning: 30,
								entities.LogLevelError:   10,
							},
							CountByService: map[string]int64{
								"gateway-service": 70,
								"search-service":  30,
							},
							CountByEvent: map[string]int64{
								"update_received": 50,
								"api_call":         50,
							},
						},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"total_count": float64(100),
			},
		},
		{
			name:        "storage unavailable",
			queryParams: "",
			setupMock: func(mock *mockGetLogStatsUseCase) {
				mock.executeFunc = func(ctx context.Context, request usecases.GetLogStatsRequest) (*usecases.GetLogStatsResponse, error) {
					return nil, domainerrors.ErrStorageUnavailable
				}
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody: map[string]interface{}{
				"error":   "Storage unavailable",
				"success": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockUseCase := &mockGetLogStatsUseCase{}
			tt.setupMock(mockUseCase)

			handler := NewLogsHandler(nil, nil, mockUseCase)

			// Prepare request
			req := httptest.NewRequest(http.MethodGet, "/logs/stats?"+tt.queryParams, nil)
			recorder := httptest.NewRecorder()

			// Execute
			handler.GetStats(recorder, req)

			// Assertions
			if recorder.Code != tt.expectedStatus {
				t.Errorf("GetStats() status code = %d, want %d", recorder.Code, tt.expectedStatus)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			for key, expectedValue := range tt.expectedBody {
				if key == "total_count" {
					// Для статистики нужно проверить вложенную структуру
					if stats, ok := response["stats"].(map[string]interface{}); ok {
						if stats["total_count"] != expectedValue {
							t.Errorf("GetStats() response.stats.total_count = %v, want %v", stats["total_count"], expectedValue)
						}
					} else {
						t.Errorf("GetStats() response missing stats object")
					}
				} else {
					if response[key] != expectedValue {
						t.Errorf("GetStats() response[%s] = %v, want %v", key, response[key], expectedValue)
					}
				}
			}
		})
	}
}

// TestLogsHandler_MethodNotAllowed тестирует обработку неподдерживаемых HTTP методов
func TestLogsHandler_MethodNotAllowed(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		endpoint string
		handler  func(*LogsHandler, http.ResponseWriter, *http.Request)
	}{
		{
			name:     "CreateLog with GET method",
			method:   http.MethodGet,
			endpoint: "/logs",
			handler:  (*LogsHandler).CreateLog,
		},
		{
			name:     "GetLogs with POST method",
			method:   http.MethodPost,
			endpoint: "/logs",
			handler:  (*LogsHandler).GetLogs,
		},
		{
			name:     "GetStats with POST method",
			method:   http.MethodPost,
			endpoint: "/logs/stats",
			handler:  (*LogsHandler).GetStats,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := &mockLogEventUseCase{}
			handler := NewLogsHandler(mockUseCase, nil, nil)

			req := httptest.NewRequest(tt.method, tt.endpoint, nil)
			recorder := httptest.NewRecorder()

			tt.handler(handler, recorder, req)

			if recorder.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s status code = %d, want %d", tt.name, recorder.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

// TestLogsHandler_ParseLogLevel тестирует парсинг уровней логирования
func TestLogsHandler_ParseLogLevel(t *testing.T) {
	tests := []struct {
		name      string
		levelStr  string
		expected  entities.LogLevel
		expectErr bool
	}{
		{"debug level", "DEBUG", entities.LogLevelDebug, false},
		{"info level", "INFO", entities.LogLevelInfo, false},
		{"warning level", "WARNING", entities.LogLevelWarning, false},
		{"warn alias", "WARN", entities.LogLevelWarning, false},
		{"error level", "ERROR", entities.LogLevelError, false},
		{"critical level", "CRITICAL", entities.LogLevelCritical, false},
		{"crit alias", "CRIT", entities.LogLevelCritical, false},
		{"lowercase debug", "debug", entities.LogLevelDebug, false},
		{"mixed case info", "Info", entities.LogLevelInfo, false},
		{"invalid level", "INVALID", 0, true},
		{"empty level", "", 0, true},
	}

	handler := &LogsHandler{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.parseLogLevel(tt.levelStr)

			if tt.expectErr {
				if err == nil {
					t.Errorf("parseLogLevel(%s) expected error but got none", tt.levelStr)
				}
			} else {
				if err != nil {
					t.Errorf("parseLogLevel(%s) unexpected error: %v", tt.levelStr, err)
				}
				if result != tt.expected {
					t.Errorf("parseLogLevel(%s) = %v, want %v", tt.levelStr, result, tt.expected)
				}
			}
		})
	}
}

// TestHealthHandler тестирует health check endpoint
func TestHealthHandler(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	handler.Check(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Health check status code = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}

	expectedFields := []string{"status", "timestamp", "version", "service"}
	for _, field := range expectedFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Health response missing field: %s", field)
		}
	}

	if response["status"] != "healthy" {
		t.Errorf("Health status = %v, want healthy", response["status"])
	}

	if response["service"] != "logging-service" {
		t.Errorf("Health service = %v, want logging-service", response["service"])
	}
}

// TestHealthHandler_MethodNotAllowed тестирует неподдерживаемые методы для health check
func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	recorder := httptest.NewRecorder()

	handler.Check(recorder, req)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("Health check with POST status code = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}
