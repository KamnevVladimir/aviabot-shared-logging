package usecases

import (
	"time"
	
	"github.com/KamnevVladimir/aviabot-shared-logging/domain/entities"
	"github.com/KamnevVladimir/aviabot-shared-logging/domain/interfaces"
)

// LogEventRequest представляет запрос на создание лог записи
type LogEventRequest struct {
	Level    entities.LogLevel          `json:"level"`
	Service  string                     `json:"service"`
	Event    string                     `json:"event"`
	Message  string                     `json:"message"`
	UserID   *int64                     `json:"user_id,omitempty"`
	ChatID   *int64                     `json:"chat_id,omitempty"`
	Metadata map[string]interface{}     `json:"metadata,omitempty"`
}

// LogEventResponse представляет ответ на создание лог записи
type LogEventResponse struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	AlertSent bool      `json:"alert_sent"`
}

// QueryLogsRequest представляет запрос на поиск логов
type QueryLogsRequest struct {
	Filter interfaces.LogFilter `json:"filter"`
}

// QueryLogsResponse представляет ответ на поиск логов
type QueryLogsResponse struct {
	Logs       []entities.LogEntry `json:"logs"`
	TotalCount int64               `json:"total_count"`
	HasMore    bool                `json:"has_more"`
}

// GetLogStatsRequest представляет запрос на получение статистики
type GetLogStatsRequest struct {
	Filter interfaces.LogFilter `json:"filter"`
}

// GetLogStatsResponse представляет ответ со статистикой
type GetLogStatsResponse struct {
	Stats interfaces.LogStats `json:"stats"`
}