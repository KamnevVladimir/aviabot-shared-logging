package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/KamnevVladimir/aviabot-shared-logging/domain/entities"
	"github.com/KamnevVladimir/aviabot-shared-logging/domain/errors"
	"github.com/KamnevVladimir/aviabot-shared-logging/domain/interfaces"
)

// Mock implementations for testing
type mockLogRepository struct {
	storeFunc    func(ctx context.Context, logEntry entities.LogEntry) error
	getByIDFunc  func(ctx context.Context, id string) (*entities.LogEntry, error)
	queryFunc    func(ctx context.Context, filter interfaces.LogFilter) ([]entities.LogEntry, error)
	countFunc    func(ctx context.Context, filter interfaces.LogFilter) (int64, error)
	getStatsFunc func(ctx context.Context, filter interfaces.LogFilter) (*interfaces.LogStats, error)
	deleteFunc   func(ctx context.Context, filter interfaces.LogFilter) (int64, error)
}

func (m *mockLogRepository) Store(ctx context.Context, logEntry entities.LogEntry) error {
	if m.storeFunc != nil {
		return m.storeFunc(ctx, logEntry)
	}
	return nil
}

func (m *mockLogRepository) GetByID(ctx context.Context, id string) (*entities.LogEntry, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.ErrLogNotFound
}

func (m *mockLogRepository) Query(ctx context.Context, filter interfaces.LogFilter) ([]entities.LogEntry, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, filter)
	}
	return []entities.LogEntry{}, nil
}

func (m *mockLogRepository) Count(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
	if m.countFunc != nil {
		return m.countFunc(ctx, filter)
	}
	return 0, nil
}

func (m *mockLogRepository) GetStats(ctx context.Context, filter interfaces.LogFilter) (*interfaces.LogStats, error) {
	if m.getStatsFunc != nil {
		return m.getStatsFunc(ctx, filter)
	}
	return &interfaces.LogStats{}, nil
}

func (m *mockLogRepository) Delete(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, filter)
	}
	return 0, nil
}

type mockAlertService struct {
	sendAlertFunc      func(ctx context.Context, logEntry entities.LogEntry) error
	sendBatchAlertFunc func(ctx context.Context, entries []entities.LogEntry) error
	isHealthyFunc      func(ctx context.Context) bool
}

func (m *mockAlertService) SendAlert(ctx context.Context, logEntry entities.LogEntry) error {
	if m.sendAlertFunc != nil {
		return m.sendAlertFunc(ctx, logEntry)
	}
	return nil
}

func (m *mockAlertService) SendBatchAlert(ctx context.Context, entries []entities.LogEntry) error {
	if m.sendBatchAlertFunc != nil {
		return m.sendBatchAlertFunc(ctx, entries)
	}
	return nil
}

func (m *mockAlertService) IsHealthy(ctx context.Context) bool {
	if m.isHealthyFunc != nil {
		return m.isHealthyFunc(ctx)
	}
	return true
}

type mockIDGenerator struct {
	generateFunc func() string
}

func (m *mockIDGenerator) Generate() string {
	if m.generateFunc != nil {
		return m.generateFunc()
	}
	return "test-id-123"
}

type mockTimeProvider struct {
	nowFunc func() time.Time
}

func (m *mockTimeProvider) Now() time.Time {
	if m.nowFunc != nil {
		return m.nowFunc()
	}
	return time.Date(2025, 9, 1, 15, 30, 0, 0, time.UTC)
}

// TestLogEventUseCase_Execute тестирует основной сценарий логирования события
func TestLogEventUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2025, 9, 1, 15, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		request        LogEventRequest
		setupMocks     func(*mockLogRepository, *mockAlertService, *mockIDGenerator, *mockTimeProvider)
		expectedError  error
		expectedResult *LogEventResponse
	}{
		{
			name: "successful log event creation",
			request: LogEventRequest{
				Level:   entities.LogLevelInfo,
				Service: "gateway-service",
				Event:   "update_received",
				Message: "Update processed successfully",
				UserID:  int64Ptr(12345),
				ChatID:  int64Ptr(67890),
				Metadata: map[string]interface{}{
					"duration_ms": 150,
					"status":      "success",
				},
			},
			setupMocks: func(repo *mockLogRepository, alert *mockAlertService, idGen *mockIDGenerator, timeProvider *mockTimeProvider) {
				idGen.generateFunc = func() string { return "log-123" }
				timeProvider.nowFunc = func() time.Time { return fixedTime }
				repo.storeFunc = func(ctx context.Context, logEntry entities.LogEntry) error {
					return nil
				}
			},
			expectedError: nil,
			expectedResult: &LogEventResponse{
				ID:        "log-123",
				Timestamp: fixedTime,
				Success:   true,
			},
		},
		{
			name: "successful log event with alert",
			request: LogEventRequest{
				Level:   entities.LogLevelError,
				Service: "gateway-service",
				Event:   "update_error",
				Message: "Failed to process update",
			},
			setupMocks: func(repo *mockLogRepository, alert *mockAlertService, idGen *mockIDGenerator, timeProvider *mockTimeProvider) {
				idGen.generateFunc = func() string { return "log-error-456" }
				timeProvider.nowFunc = func() time.Time { return fixedTime }
				repo.storeFunc = func(ctx context.Context, logEntry entities.LogEntry) error {
					return nil
				}
				alert.sendAlertFunc = func(ctx context.Context, logEntry entities.LogEntry) error {
					return nil
				}
			},
			expectedError: nil,
			expectedResult: &LogEventResponse{
				ID:        "log-error-456",
				Timestamp: fixedTime,
				Success:   true,
				AlertSent: true,
			},
		},
		{
			name: "invalid log level",
			request: LogEventRequest{
				Level:   entities.LogLevel(99),
				Service: "gateway-service",
				Event:   "update_received",
				Message: "Test message",
			},
			setupMocks: func(repo *mockLogRepository, alert *mockAlertService, idGen *mockIDGenerator, timeProvider *mockTimeProvider) {
				// No setup needed for validation error
			},
			expectedError:  errors.ErrInvalidLogEntry,
			expectedResult: nil,
		},
		{
			name: "empty service name",
			request: LogEventRequest{
				Level:   entities.LogLevelInfo,
				Service: "",
				Event:   "update_received",
				Message: "Test message",
			},
			setupMocks: func(repo *mockLogRepository, alert *mockAlertService, idGen *mockIDGenerator, timeProvider *mockTimeProvider) {
				// No setup needed for validation error
			},
			expectedError:  errors.ErrInvalidLogEntry,
			expectedResult: nil,
		},
		{
			name: "empty event name",
			request: LogEventRequest{
				Level:   entities.LogLevelInfo,
				Service: "gateway-service",
				Event:   "",
				Message: "Test message",
			},
			setupMocks: func(repo *mockLogRepository, alert *mockAlertService, idGen *mockIDGenerator, timeProvider *mockTimeProvider) {
				// No setup needed for validation error
			},
			expectedError:  errors.ErrInvalidLogEntry,
			expectedResult: nil,
		},
		{
			name: "empty message",
			request: LogEventRequest{
				Level:   entities.LogLevelInfo,
				Service: "gateway-service",
				Event:   "update_received",
				Message: "",
			},
			setupMocks: func(repo *mockLogRepository, alert *mockAlertService, idGen *mockIDGenerator, timeProvider *mockTimeProvider) {
				// No setup needed for validation error
			},
			expectedError:  errors.ErrInvalidLogEntry,
			expectedResult: nil,
		},
		{
			name: "repository storage error",
			request: LogEventRequest{
				Level:   entities.LogLevelInfo,
				Service: "gateway-service",
				Event:   "update_received",
				Message: "Test message",
			},
			setupMocks: func(repo *mockLogRepository, alert *mockAlertService, idGen *mockIDGenerator, timeProvider *mockTimeProvider) {
				idGen.generateFunc = func() string { return "log-789" }
				timeProvider.nowFunc = func() time.Time { return fixedTime }
				repo.storeFunc = func(ctx context.Context, logEntry entities.LogEntry) error {
					return errors.ErrStorageUnavailable
				}
			},
			expectedError:  errors.ErrStorageUnavailable,
			expectedResult: nil,
		},
		{
			name: "alert service error - log still saved",
			request: LogEventRequest{
				Level:   entities.LogLevelCritical,
				Service: "gateway-service",
				Event:   "critical_error",
				Message: "Critical system failure",
			},
			setupMocks: func(repo *mockLogRepository, alert *mockAlertService, idGen *mockIDGenerator, timeProvider *mockTimeProvider) {
				idGen.generateFunc = func() string { return "log-critical-111" }
				timeProvider.nowFunc = func() time.Time { return fixedTime }
				repo.storeFunc = func(ctx context.Context, logEntry entities.LogEntry) error {
					return nil
				}
				alert.sendAlertFunc = func(ctx context.Context, logEntry entities.LogEntry) error {
					return errors.ErrAlertServiceUnavailable
				}
			},
			expectedError: nil,
			expectedResult: &LogEventResponse{
				ID:        "log-critical-111",
				Timestamp: fixedTime,
				Success:   true,
				AlertSent: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &mockLogRepository{}
			mockAlert := &mockAlertService{}
			mockIDGen := &mockIDGenerator{}
			mockTimeProvider := &mockTimeProvider{}

			tt.setupMocks(mockRepo, mockAlert, mockIDGen, mockTimeProvider)

			// Create use case
			useCase := NewLogEventUseCase(mockRepo, mockAlert, mockIDGen, mockTimeProvider)

			// Execute
			result, err := useCase.Execute(ctx, tt.request)

			// Assertions
			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("LogEventUseCase.Execute() expected error = %v, got nil", tt.expectedError)
					return
				}
				if err != tt.expectedError {
					t.Errorf("LogEventUseCase.Execute() error = %v, want %v", err, tt.expectedError)
					return
				}
			} else {
				if err != nil {
					t.Errorf("LogEventUseCase.Execute() unexpected error = %v", err)
					return
				}
			}

			if tt.expectedResult != nil {
				if result == nil {
					t.Errorf("LogEventUseCase.Execute() result = nil, want %v", tt.expectedResult)
					return
				}

				if result.ID != tt.expectedResult.ID {
					t.Errorf("LogEventUseCase.Execute() result.ID = %v, want %v", result.ID, tt.expectedResult.ID)
				}

				if !result.Timestamp.Equal(tt.expectedResult.Timestamp) {
					t.Errorf("LogEventUseCase.Execute() result.Timestamp = %v, want %v", result.Timestamp, tt.expectedResult.Timestamp)
				}

				if result.Success != tt.expectedResult.Success {
					t.Errorf("LogEventUseCase.Execute() result.Success = %v, want %v", result.Success, tt.expectedResult.Success)
				}

				if result.AlertSent != tt.expectedResult.AlertSent {
					t.Errorf("LogEventUseCase.Execute() result.AlertSent = %v, want %v", result.AlertSent, tt.expectedResult.AlertSent)
				}
			}
		})
	}
}

// Helper function
func int64Ptr(v int64) *int64 {
	return &v
}
