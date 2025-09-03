package interfaces

import (
	"context"
	"time"

	"github.com/KamnevVladimir/aviabot-shared-logging/domain/entities"
)

// LogRepository определяет контракт для хранения и получения логов
type LogRepository interface {
	// Store сохраняет лог запись в хранилище
	Store(ctx context.Context, logEntry entities.LogEntry) error

	// GetByID получает лог запись по ID
	GetByID(ctx context.Context, id string) (*entities.LogEntry, error)

	// Query получает логи по фильтрам с пагинацией
	Query(ctx context.Context, filter LogFilter) ([]entities.LogEntry, error)

	// Count возвращает количество записей по фильтру
	Count(ctx context.Context, filter LogFilter) (int64, error)

	// GetStats возвращает статистику логирования
	GetStats(ctx context.Context, filter LogFilter) (*LogStats, error)

	// Delete удаляет лог записи по фильтру (для очистки старых логов)
	Delete(ctx context.Context, filter LogFilter) (int64, error)
}

// LogFilter определяет параметры фильтрации логов
type LogFilter struct {
	// Временные рамки
	TimeFrom *time.Time
	TimeTo   *time.Time

	// Фильтры по полям
	Services []string            // Список сервисов
	Events   []string            // Список событий
	Levels   []entities.LogLevel // Уровни логирования
	UserID   *int64              // ID пользователя
	ChatID   *int64              // ID чата

	// Поиск по тексту
	MessageContains string // Поиск в сообщении

	// Пагинация
	Limit  int // Лимит записей
	Offset int // Смещение

	// Сортировка
	SortBy    string // Поле для сортировки (timestamp, level, service)
	SortOrder string // Порядок сортировки (asc, desc)
}

// LogStats представляет статистику логирования
type LogStats struct {
	TotalCount     int64                       `json:"total_count"`
	CountByLevel   map[entities.LogLevel]int64 `json:"count_by_level"`
	CountByService map[string]int64            `json:"count_by_service"`
	CountByEvent   map[string]int64            `json:"count_by_event"`
	TimeRange      TimeRange                   `json:"time_range"`
}

// TimeRange представляет временной диапазон
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// AlertService определяет контракт для отправки алертов
type AlertService interface {
	// SendAlert отправляет алерт о критической ошибке
	SendAlert(ctx context.Context, logEntry entities.LogEntry) error

	// SendBatchAlert отправляет сводный алерт о множественных проблемах
	SendBatchAlert(ctx context.Context, entries []entities.LogEntry) error

	// IsHealthy проверяет работоспособность сервиса алертов
	IsHealthy(ctx context.Context) bool
}

// LogIDGenerator определяет контракт для генерации уникальных ID логов
type LogIDGenerator interface {
	// Generate создает уникальный ID для лог записи
	Generate() string
}

// TimeProvider определяет контракт для получения текущего времени (для тестирования)
type TimeProvider interface {
	// Now возвращает текущее время
	Now() time.Time
}
