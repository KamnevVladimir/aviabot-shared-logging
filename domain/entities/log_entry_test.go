package entities

import (
	"testing"
	"time"
)

// TestLogLevel_String тестирует строковое представление уровня логирования
func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
		want  string
	}{
		{
			name:  "debug level",
			level: LogLevelDebug,
			want:  "DEBUG",
		},
		{
			name:  "info level",
			level: LogLevelInfo,
			want:  "INFO",
		},
		{
			name:  "warning level",
			level: LogLevelWarning,
			want:  "WARNING",
		},
		{
			name:  "error level",
			level: LogLevelError,
			want:  "ERROR",
		},
		{
			name:  "critical level",
			level: LogLevelCritical,
			want:  "CRITICAL",
		},
		{
			name:  "unknown level",
			level: LogLevel(99),
			want:  "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.want {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLogLevel_IsValid тестирует валидацию уровня логирования
func TestLogLevel_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
		want  bool
	}{
		{
			name:  "valid debug level",
			level: LogLevelDebug,
			want:  true,
		},
		{
			name:  "valid info level",
			level: LogLevelInfo,
			want:  true,
		},
		{
			name:  "valid warning level",
			level: LogLevelWarning,
			want:  true,
		},
		{
			name:  "valid error level",
			level: LogLevelError,
			want:  true,
		},
		{
			name:  "valid critical level",
			level: LogLevelCritical,
			want:  true,
		},
		{
			name:  "invalid level",
			level: LogLevel(99),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.IsValid()
			if got != tt.want {
				t.Errorf("LogLevel.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLogEntry_IsValid тестирует валидацию лог записи
func TestLogEntry_IsValid(t *testing.T) {
	validTimestamp := time.Now()

	tests := []struct {
		name     string
		logEntry LogEntry
		want     bool
	}{
		{
			name: "valid log entry",
			logEntry: LogEntry{
				ID:        "log-123",
				Level:     LogLevelInfo,
				Service:   "gateway-service",
				Event:     "update_received",
				Timestamp: validTimestamp,
				Message:   "Update processed successfully",
			},
			want: true,
		},
		{
			name: "valid log entry with optional fields",
			logEntry: LogEntry{
				ID:        "log-124",
				Level:     LogLevelDebug,
				Service:   "search-service",
				Event:     "api_call",
				Timestamp: validTimestamp,
				UserID:    int64Ptr(12345),
				ChatID:    int64Ptr(67890),
				Message:   "API call initiated",
				Metadata: map[string]interface{}{
					"duration_ms": 150,
					"endpoint":    "/search",
				},
			},
			want: true,
		},
		{
			name: "invalid log entry - empty ID",
			logEntry: LogEntry{
				ID:        "",
				Level:     LogLevelInfo,
				Service:   "gateway-service",
				Event:     "update_received",
				Timestamp: validTimestamp,
				Message:   "Test message",
			},
			want: false,
		},
		{
			name: "invalid log entry - invalid level",
			logEntry: LogEntry{
				ID:        "log-123",
				Level:     LogLevel(99),
				Service:   "gateway-service",
				Event:     "update_received",
				Timestamp: validTimestamp,
				Message:   "Test message",
			},
			want: false,
		},
		{
			name: "invalid log entry - empty service",
			logEntry: LogEntry{
				ID:        "log-123",
				Level:     LogLevelInfo,
				Service:   "",
				Event:     "update_received",
				Timestamp: validTimestamp,
				Message:   "Test message",
			},
			want: false,
		},
		{
			name: "invalid log entry - empty event",
			logEntry: LogEntry{
				ID:        "log-123",
				Level:     LogLevelInfo,
				Service:   "gateway-service",
				Event:     "",
				Timestamp: validTimestamp,
				Message:   "Test message",
			},
			want: false,
		},
		{
			name: "invalid log entry - zero timestamp",
			logEntry: LogEntry{
				ID:        "log-123",
				Level:     LogLevelInfo,
				Service:   "gateway-service",
				Event:     "update_received",
				Timestamp: time.Time{},
				Message:   "Test message",
			},
			want: false,
		},
		{
			name: "invalid log entry - empty message",
			logEntry: LogEntry{
				ID:        "log-123",
				Level:     LogLevelInfo,
				Service:   "gateway-service",
				Event:     "update_received",
				Timestamp: validTimestamp,
				Message:   "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.logEntry.IsValid()
			if got != tt.want {
				t.Errorf("LogEntry.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLogEntry_ToJSON тестирует сериализацию в JSON
func TestLogEntry_ToJSON(t *testing.T) {
	timestamp := time.Date(2025, 9, 1, 15, 30, 0, 0, time.UTC)

	logEntry := LogEntry{
		ID:        "log-123",
		Level:     LogLevelInfo,
		Service:   "gateway-service",
		Event:     "update_received",
		Timestamp: timestamp,
		UserID:    int64Ptr(12345),
		ChatID:    int64Ptr(67890),
		Message:   "Update processed successfully",
		Metadata: map[string]interface{}{
			"duration_ms": 150,
			"status":      "success",
		},
	}

	jsonData, err := logEntry.ToJSON()
	if err != nil {
		t.Errorf("LogEntry.ToJSON() error = %v", err)
		return
	}

	// Проверяем, что JSON содержит ожидаемые поля
	expectedFields := []string{
		`"id":"log-123"`,
		`"level":"INFO"`,
		`"service":"gateway-service"`,
		`"event":"update_received"`,
		`"timestamp":"2025-09-01T15:30:00Z"`,
		`"user_id":12345`,
		`"chat_id":67890`,
		`"message":"Update processed successfully"`,
		`"duration_ms":150`,
		`"status":"success"`,
	}

	jsonString := string(jsonData)
	for _, field := range expectedFields {
		if !contains(jsonString, field) {
			t.Errorf("LogEntry.ToJSON() missing field %s in output: %s", field, jsonString)
		}
	}
}

// TestLogEntry_GetPriority тестирует получение приоритета лог записи
func TestLogEntry_GetPriority(t *testing.T) {
	tests := []struct {
		name     string
		logEntry LogEntry
		want     int
	}{
		{
			name: "critical priority",
			logEntry: LogEntry{
				Level: LogLevelCritical,
			},
			want: 5,
		},
		{
			name: "error priority",
			logEntry: LogEntry{
				Level: LogLevelError,
			},
			want: 4,
		},
		{
			name: "warning priority",
			logEntry: LogEntry{
				Level: LogLevelWarning,
			},
			want: 3,
		},
		{
			name: "info priority",
			logEntry: LogEntry{
				Level: LogLevelInfo,
			},
			want: 2,
		},
		{
			name: "debug priority",
			logEntry: LogEntry{
				Level: LogLevelDebug,
			},
			want: 1,
		},
		{
			name: "unknown level priority",
			logEntry: LogEntry{
				Level: LogLevel(99),
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.logEntry.GetPriority()
			if got != tt.want {
				t.Errorf("LogEntry.GetPriority() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLogEntry_ShouldAlert тестирует определение необходимости алерта
func TestLogEntry_ShouldAlert(t *testing.T) {
	tests := []struct {
		name     string
		logEntry LogEntry
		want     bool
	}{
		{
			name: "critical level should alert",
			logEntry: LogEntry{
				Level: LogLevelCritical,
			},
			want: true,
		},
		{
			name: "error level should alert",
			logEntry: LogEntry{
				Level: LogLevelError,
			},
			want: true,
		},
		{
			name: "warning level should not alert",
			logEntry: LogEntry{
				Level: LogLevelWarning,
			},
			want: false,
		},
		{
			name: "info level should not alert",
			logEntry: LogEntry{
				Level: LogLevelInfo,
			},
			want: false,
		},
		{
			name: "debug level should not alert",
			logEntry: LogEntry{
				Level: LogLevelDebug,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.logEntry.ShouldAlert()
			if got != tt.want {
				t.Errorf("LogEntry.ShouldAlert() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper functions for tests
func int64Ptr(v int64) *int64 {
	return &v
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
