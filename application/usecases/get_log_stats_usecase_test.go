package usecases

import (
	"context"
	"testing"
	"time"
	
	"github.com/KamnevVladimir/aviabot-shared-logging/domain/entities"
	"github.com/KamnevVladimir/aviabot-shared-logging/domain/errors"
	"github.com/KamnevVladimir/aviabot-shared-logging/domain/interfaces"
)

// TestGetLogStatsUseCase_Execute тестирует получение статистики логирования
func TestGetLogStatsUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2025, 9, 1, 15, 30, 0, 0, time.UTC)
	
	sampleStats := &interfaces.LogStats{
		TotalCount: 100,
		CountByLevel: map[entities.LogLevel]int64{
			entities.LogLevelDebug:    10,
			entities.LogLevelInfo:     60,
			entities.LogLevelWarning:  20,
			entities.LogLevelError:    8,
			entities.LogLevelCritical: 2,
		},
		CountByService: map[string]int64{
			"gateway-service": 50,
			"search-service":  30,
			"ui-service":      20,
		},
		CountByEvent: map[string]int64{
			"update_received": 40,
			"api_call":         30,
			"user_action":      20,
			"error_occurred":   10,
		},
		TimeRange: interfaces.TimeRange{
			From: fixedTime.Add(-24 * time.Hour),
			To:   fixedTime,
		},
	}
	
	tests := []struct {
		name           string
		request        GetLogStatsRequest
		setupMocks     func(*mockLogRepository)
		expectedError  error
		expectedResult *GetLogStatsResponse
	}{
		{
			name: "successful stats retrieval with time filter",
			request: GetLogStatsRequest{
				Filter: interfaces.LogFilter{
					TimeFrom: timePtr(fixedTime.Add(-24 * time.Hour)),
					TimeTo:   &fixedTime,
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				repo.getStatsFunc = func(ctx context.Context, filter interfaces.LogFilter) (*interfaces.LogStats, error) {
					return sampleStats, nil
				}
			},
			expectedError: nil,
			expectedResult: &GetLogStatsResponse{
				Stats: *sampleStats,
			},
		},
		{
			name: "stats with service filter",
			request: GetLogStatsRequest{
				Filter: interfaces.LogFilter{
					Services: []string{"gateway-service", "search-service"},
					TimeFrom: timePtr(fixedTime.Add(-time.Hour)),
					TimeTo:   &fixedTime,
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				filteredStats := &interfaces.LogStats{
					TotalCount: 80,
					CountByLevel: map[entities.LogLevel]int64{
						entities.LogLevelInfo:     50,
						entities.LogLevelWarning:  20,
						entities.LogLevelError:    8,
						entities.LogLevelCritical: 2,
					},
					CountByService: map[string]int64{
						"gateway-service": 50,
						"search-service":  30,
					},
					CountByEvent: map[string]int64{
						"update_received": 40,
						"api_call":         30,
						"error_occurred":   10,
					},
					TimeRange: interfaces.TimeRange{
						From: fixedTime.Add(-time.Hour),
						To:   fixedTime,
					},
				}
				repo.getStatsFunc = func(ctx context.Context, filter interfaces.LogFilter) (*interfaces.LogStats, error) {
					return filteredStats, nil
				}
			},
			expectedError: nil,
			expectedResult: &GetLogStatsResponse{
				Stats: interfaces.LogStats{
					TotalCount: 80,
					CountByLevel: map[entities.LogLevel]int64{
						entities.LogLevelInfo:     50,
						entities.LogLevelWarning:  20,
						entities.LogLevelError:    8,
						entities.LogLevelCritical: 2,
					},
					CountByService: map[string]int64{
						"gateway-service": 50,
						"search-service":  30,
					},
					CountByEvent: map[string]int64{
						"update_received": 40,
						"api_call":         30,
						"error_occurred":   10,
					},
					TimeRange: interfaces.TimeRange{
						From: fixedTime.Add(-time.Hour),
						To:   fixedTime,
					},
				},
			},
		},
		{
			name: "stats with level filter",
			request: GetLogStatsRequest{
				Filter: interfaces.LogFilter{
					Levels: []entities.LogLevel{entities.LogLevelError, entities.LogLevelCritical},
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				errorStats := &interfaces.LogStats{
					TotalCount: 10,
					CountByLevel: map[entities.LogLevel]int64{
						entities.LogLevelError:    8,
						entities.LogLevelCritical: 2,
					},
					CountByService: map[string]int64{
						"gateway-service": 5,
						"search-service":  3,
						"ui-service":      2,
					},
					CountByEvent: map[string]int64{
						"error_occurred":  8,
						"critical_error":  2,
					},
					TimeRange: interfaces.TimeRange{
						From: fixedTime.Add(-24 * time.Hour),
						To:   fixedTime,
					},
				}
				repo.getStatsFunc = func(ctx context.Context, filter interfaces.LogFilter) (*interfaces.LogStats, error) {
					return errorStats, nil
				}
			},
			expectedError: nil,
			expectedResult: &GetLogStatsResponse{
				Stats: interfaces.LogStats{
					TotalCount: 10,
					CountByLevel: map[entities.LogLevel]int64{
						entities.LogLevelError:    8,
						entities.LogLevelCritical: 2,
					},
					CountByService: map[string]int64{
						"gateway-service": 5,
						"search-service":  3,
						"ui-service":      2,
					},
					CountByEvent: map[string]int64{
						"error_occurred":  8,
						"critical_error":  2,
					},
					TimeRange: interfaces.TimeRange{
						From: fixedTime.Add(-24 * time.Hour),
						To:   fixedTime,
					},
				},
			},
		},
		{
			name: "empty stats",
			request: GetLogStatsRequest{
				Filter: interfaces.LogFilter{
					Services: []string{"non-existent-service"},
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				emptyStats := &interfaces.LogStats{
					TotalCount:     0,
					CountByLevel:   map[entities.LogLevel]int64{},
					CountByService: map[string]int64{},
					CountByEvent:   map[string]int64{},
					TimeRange: interfaces.TimeRange{
						From: fixedTime,
						To:   fixedTime,
					},
				}
				repo.getStatsFunc = func(ctx context.Context, filter interfaces.LogFilter) (*interfaces.LogStats, error) {
					return emptyStats, nil
				}
			},
			expectedError: nil,
			expectedResult: &GetLogStatsResponse{
				Stats: interfaces.LogStats{
					TotalCount:     0,
					CountByLevel:   map[entities.LogLevel]int64{},
					CountByService: map[string]int64{},
					CountByEvent:   map[string]int64{},
					TimeRange: interfaces.TimeRange{
						From: fixedTime,
						To:   fixedTime,
					},
				},
			},
		},
		{
			name: "invalid time range - TimeTo before TimeFrom",
			request: GetLogStatsRequest{
				Filter: interfaces.LogFilter{
					TimeFrom: &fixedTime,
					TimeTo:   timePtr(fixedTime.Add(-time.Hour)),
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				// No setup needed for validation error
			},
			expectedError: errors.ErrInvalidFilter,
			expectedResult: nil,
		},
		{
			name: "repository error",
			request: GetLogStatsRequest{
				Filter: interfaces.LogFilter{
					TimeFrom: timePtr(fixedTime.Add(-time.Hour)),
					TimeTo:   &fixedTime,
				},
			},
			setupMocks: func(repo *mockLogRepository) {
				repo.getStatsFunc = func(ctx context.Context, filter interfaces.LogFilter) (*interfaces.LogStats, error) {
					return nil, errors.ErrStorageUnavailable
				}
			},
			expectedError: errors.ErrStorageUnavailable,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &mockLogRepository{}
			tt.setupMocks(mockRepo)
			
			// Create use case
			useCase := NewGetLogStatsUseCase(mockRepo)
			
			// Execute
			result, err := useCase.Execute(ctx, tt.request)
			
			// Assertions
			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("GetLogStatsUseCase.Execute() expected error = %v, got nil", tt.expectedError)
					return
				}
				if err != tt.expectedError {
					t.Errorf("GetLogStatsUseCase.Execute() error = %v, want %v", err, tt.expectedError)
					return
				}
			} else {
				if err != nil {
					t.Errorf("GetLogStatsUseCase.Execute() unexpected error = %v", err)
					return
				}
			}
			
			if tt.expectedResult != nil {
				if result == nil {
					t.Errorf("GetLogStatsUseCase.Execute() result = nil, want %v", tt.expectedResult)
					return
				}
				
				if result.Stats.TotalCount != tt.expectedResult.Stats.TotalCount {
					t.Errorf("GetLogStatsUseCase.Execute() result.Stats.TotalCount = %d, want %d", result.Stats.TotalCount, tt.expectedResult.Stats.TotalCount)
				}
				
				// Compare CountByLevel maps
				if len(result.Stats.CountByLevel) != len(tt.expectedResult.Stats.CountByLevel) {
					t.Errorf("GetLogStatsUseCase.Execute() result.Stats.CountByLevel length = %d, want %d", len(result.Stats.CountByLevel), len(tt.expectedResult.Stats.CountByLevel))
				}
				
				for level, count := range tt.expectedResult.Stats.CountByLevel {
					if result.Stats.CountByLevel[level] != count {
						t.Errorf("GetLogStatsUseCase.Execute() result.Stats.CountByLevel[%v] = %d, want %d", level, result.Stats.CountByLevel[level], count)
					}
				}
				
				// Compare CountByService maps
				if len(result.Stats.CountByService) != len(tt.expectedResult.Stats.CountByService) {
					t.Errorf("GetLogStatsUseCase.Execute() result.Stats.CountByService length = %d, want %d", len(result.Stats.CountByService), len(tt.expectedResult.Stats.CountByService))
				}
				
				for service, count := range tt.expectedResult.Stats.CountByService {
					if result.Stats.CountByService[service] != count {
						t.Errorf("GetLogStatsUseCase.Execute() result.Stats.CountByService[%s] = %d, want %d", service, result.Stats.CountByService[service], count)
					}
				}
			}
		})
	}
}