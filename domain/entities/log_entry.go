package entities

import (
	"encoding/json"
	"strings"
	"time"
)

// LogLevel представляет уровень логирования
type LogLevel int

const (
	LogLevelDebug LogLevel = iota + 1
	LogLevelInfo
	LogLevelWarning
	LogLevelError
	LogLevelCritical
)

// String возвращает строковое представление уровня логирования
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarning:
		return "WARNING"
	case LogLevelError:
		return "ERROR"
	case LogLevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// IsValid проверяет валидность уровня логирования
func (l LogLevel) IsValid() bool {
	return l >= LogLevelDebug && l <= LogLevelCritical
}

// LogEntry представляет запись в логе
type LogEntry struct {
	ID        string                 `json:"id"`
	Level     LogLevel               `json:"level"`
	Service   string                 `json:"service"`
	Event     string                 `json:"event"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    *int64                 `json:"user_id,omitempty"`
	ChatID    *int64                 `json:"chat_id,omitempty"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// IsValid проверяет валидность лог записи
func (l LogEntry) IsValid() bool {
	// Проверяем обязательные поля
	if strings.TrimSpace(l.ID) == "" {
		return false
	}
	
	if !l.Level.IsValid() {
		return false
	}
	
	if strings.TrimSpace(l.Service) == "" {
		return false
	}
	
	if strings.TrimSpace(l.Event) == "" {
		return false
	}
	
	if l.Timestamp.IsZero() {
		return false
	}
	
	if strings.TrimSpace(l.Message) == "" {
		return false
	}
	
	return true
}

// ToJSON сериализует лог запись в JSON с правильным форматированием уровня
func (l LogEntry) ToJSON() ([]byte, error) {
	// Создаем вспомогательную структуру для сериализации
	type logEntryJSON struct {
		ID        string                 `json:"id"`
		Level     string                 `json:"level"`
		Service   string                 `json:"service"`
		Event     string                 `json:"event"`
		Timestamp time.Time              `json:"timestamp"`
		UserID    *int64                 `json:"user_id,omitempty"`
		ChatID    *int64                 `json:"chat_id,omitempty"`
		Message   string                 `json:"message"`
		Metadata  map[string]interface{} `json:"metadata,omitempty"`
	}
	
	jsonEntry := logEntryJSON{
		ID:        l.ID,
		Level:     l.Level.String(),
		Service:   l.Service,
		Event:     l.Event,
		Timestamp: l.Timestamp,
		UserID:    l.UserID,
		ChatID:    l.ChatID,
		Message:   l.Message,
		Metadata:  l.Metadata,
	}
	
	return json.Marshal(jsonEntry)
}

// GetPriority возвращает численный приоритет лог записи
func (l LogEntry) GetPriority() int {
	switch l.Level {
	case LogLevelDebug:
		return 1
	case LogLevelInfo:
		return 2
	case LogLevelWarning:
		return 3
	case LogLevelError:
		return 4
	case LogLevelCritical:
		return 5
	default:
		return 0
	}
}

// ShouldAlert определяет, нужно ли отправлять алерт для данной записи
func (l LogEntry) ShouldAlert() bool {
	return l.Level == LogLevelError || l.Level == LogLevelCritical
}