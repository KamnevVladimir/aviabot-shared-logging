package usecases

import (
	"context"
	"testing"
	"time"

	"aviasales-shared-logging/domain/entities"
	"aviasales-shared-logging/domain/errors"
	"aviasales-shared-logging/domain/interfaces"
)

// TestQueryLogsUseCase_Execute тестирует поиск логов с различными фильтрами
func TestQueryLogsUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2025, 9, 1, 15, 30, 0, 0, time.UTC)

	sampleLogs := []entities.LogEntry{
		{
			ID:        "log-1",
			Level:     entities.LogLevelInfo,
			Service:   "gateway-service",
			Event:     "update_received",
			Timestamp: fixedTime,
			Message:   "Update processed successfully",
		},
		{
			ID:        "log-2",
			Level:     entities.LogLevelError,
			Service:   "search-service",
			Event:     "api_error",
			Timestamp: fixedTime.Add(time.Minute),
			Message:   "API call failed",
		},
	}

	tests := []struct {
		name           string
		request        QueryLogsRequest
		setupMocks     func(*mockLogRepository)
		expectedError  error
		expectedResult *QueryLogsResponse
	}{
		{
			name: "successful query with basic filter",
			request: QueryLogsRequest{
				Filter: interfaces.LogFilter{
					Services:  []string{"gateway-service"},
					Limit:     10,
					Offset:    0,
					SortBy:    "timestamp",
					SortOrder: "desc",
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				repo.queryFunc = func(ctx context.Context, filter interfaces.LogFilter) ([]entities.LogEntry, error) {
					return []entities.LogEntry{sampleLogs[0]}, nil
				}
				repo.countFunc = func(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
					return 1, nil
				}
			},
			expectedError: nil,
			expectedResult: &QueryLogsResponse{
				Logs:       []entities.LogEntry{sampleLogs[0]},
				TotalCount: 1,
				HasMore:    false,
			},
		},
		{
			name: "query with time range filter",
			request: QueryLogsRequest{
				Filter: interfaces.LogFilter{
					TimeFrom: &fixedTime,
					TimeTo:   timePtr(fixedTime.Add(time.Hour)),
					Levels:   []entities.LogLevel{entities.LogLevelInfo, entities.LogLevelError},
					Limit:    5,
					Offset:   0,
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				repo.queryFunc = func(ctx context.Context, filter interfaces.LogFilter) ([]entities.LogEntry, error) {
					return sampleLogs, nil
				}
				repo.countFunc = func(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
					return 2, nil
				}
			},
			expectedError: nil,
			expectedResult: &QueryLogsResponse{
				Logs:       sampleLogs,
				TotalCount: 2,
				HasMore:    false,
			},
		},
		{
			name: "query with pagination - has more results",
			request: QueryLogsRequest{
				Filter: interfaces.LogFilter{
					Limit:  1,
					Offset: 0,
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				repo.queryFunc = func(ctx context.Context, filter interfaces.LogFilter) ([]entities.LogEntry, error) {
					return []entities.LogEntry{sampleLogs[0]}, nil
				}
				repo.countFunc = func(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
					return 10, nil
				}
			},
			expectedError: nil,
			expectedResult: &QueryLogsResponse{
				Logs:       []entities.LogEntry{sampleLogs[0]},
				TotalCount: 10,
				HasMore:    true,
			},
		},
		{
			name: "query with user and chat ID filter",
			request: QueryLogsRequest{
				Filter: interfaces.LogFilter{
					UserID: int64Ptr(12345),
					ChatID: int64Ptr(67890),
					Limit:  10,
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				repo.queryFunc = func(ctx context.Context, filter interfaces.LogFilter) ([]entities.LogEntry, error) {
					return []entities.LogEntry{}, nil
				}
				repo.countFunc = func(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
					return 0, nil
				}
			},
			expectedError: nil,
			expectedResult: &QueryLogsResponse{
				Logs:       []entities.LogEntry{},
				TotalCount: 0,
				HasMore:    false,
			},
		},
		{
			name: "invalid filter - negative limit",
			request: QueryLogsRequest{
				Filter: interfaces.LogFilter{
					Limit: -1,
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				// No setup needed for validation error
			},
			expectedError:  errors.ErrInvalidFilter,
			expectedResult: nil,
		},
		{
			name: "invalid filter - negative offset",
			request: QueryLogsRequest{
				Filter: interfaces.LogFilter{
					Limit:  10,
					Offset: -1,
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				// No setup needed for validation error
			},
			expectedError:  errors.ErrInvalidFilter,
			expectedResult: nil,
		},
		{
			name: "invalid filter - limit too large",
			request: QueryLogsRequest{
				Filter: interfaces.LogFilter{
					Limit: 1001, // Assuming max is 1000
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				// No setup needed for validation error
			},
			expectedError:  errors.ErrInvalidFilter,
			expectedResult: nil,
		},
		{
			name: "repository error",
			request: QueryLogsRequest{
				Filter: interfaces.LogFilter{
					Limit: 10,
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				repo.queryFunc = func(ctx context.Context, filter interfaces.LogFilter) ([]entities.LogEntry, error) {
					return nil, errors.ErrStorageUnavailable
				}
			},
			expectedError:  errors.ErrStorageUnavailable,
			expectedResult: nil,
		},
		{
			name: "count error",
			request: QueryLogsRequest{
				Filter: interfaces.LogFilter{
					Limit: 10,
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				repo.queryFunc = func(ctx context.Context, filter interfaces.LogFilter) ([]entities.LogEntry, error) {
					return []entities.LogEntry{sampleLogs[0]}, nil
				}
				repo.countFunc = func(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
					return 0, errors.ErrStorageUnavailable
				}
			},
			expectedError:  errors.ErrStorageUnavailable,
			expectedResult: nil,
		},
		{
			name: "query with default parameters",
			request: QueryLogsRequest{
				Filter: interfaces.LogFilter{
					// Limit, SortBy, SortOrder не указаны - должны применяться defaults
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				repo.queryFunc = func(ctx context.Context, filter interfaces.LogFilter) ([]entities.LogEntry, error) {
					// Проверим что применились defaults
					if filter.Limit != 100 || filter.SortBy != "timestamp" || filter.SortOrder != "desc" {
						return nil, errors.ErrInvalidFilter
					}
					return []entities.LogEntry{}, nil
				}
				repo.countFunc = func(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
					return 0, nil
				}
			},
			expectedError: nil,
			expectedResult: &QueryLogsResponse{
				Logs:       []entities.LogEntry{},
				TotalCount: 0,
				HasMore:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &mockLogRepository{}
			tt.setupMocks(mockRepo)

			// Create use case
			useCase := NewQueryLogsUseCase(mockRepo)

			// Execute
			result, err := useCase.Execute(ctx, tt.request)

			// Assertions
			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("QueryLogsUseCase.Execute() expected error = %v, got nil", tt.expectedError)
					return
				}
				if err != tt.expectedError {
					t.Errorf("QueryLogsUseCase.Execute() error = %v, want %v", err, tt.expectedError)
					return
				}
			} else {
				if err != nil {
					t.Errorf("QueryLogsUseCase.Execute() unexpected error = %v", err)
					return
				}
			}

			if tt.expectedResult != nil {
				if result == nil {
					t.Errorf("QueryLogsUseCase.Execute() result = nil, want %v", tt.expectedResult)
					return
				}

				if len(result.Logs) != len(tt.expectedResult.Logs) {
					t.Errorf("QueryLogsUseCase.Execute() result.Logs length = %d, want %d", len(result.Logs), len(tt.expectedResult.Logs))
				}

				if result.TotalCount != tt.expectedResult.TotalCount {
					t.Errorf("QueryLogsUseCase.Execute() result.TotalCount = %d, want %d", result.TotalCount, tt.expectedResult.TotalCount)
				}

				if result.HasMore != tt.expectedResult.HasMore {
					t.Errorf("QueryLogsUseCase.Execute() result.HasMore = %v, want %v", result.HasMore, tt.expectedResult.HasMore)
				}
			}
		})
	}
}

// Helper function
func timePtr(t time.Time) *time.Time {
	return &t
}
