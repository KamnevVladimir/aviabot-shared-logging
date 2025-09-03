package usecases

import (
	"context"
	"strings"
	
	"github.com/KamnevVladimir/aviabot-shared-logging/domain/entities"
	"github.com/KamnevVladimir/aviabot-shared-logging/domain/errors"
	"github.com/KamnevVladimir/aviabot-shared-logging/domain/interfaces"
)

// LogEventUseCase обрабатывает создание новых лог записей
type LogEventUseCase struct {
	repository    interfaces.LogRepository
	alertService  interfaces.AlertService
	idGenerator   interfaces.LogIDGenerator
	timeProvider  interfaces.TimeProvider
}

// NewLogEventUseCase создает новый экземпляр LogEventUseCase
func NewLogEventUseCase(
	repository interfaces.LogRepository,
	alertService interfaces.AlertService,
	idGenerator interfaces.LogIDGenerator,
	timeProvider interfaces.TimeProvider,
) *LogEventUseCase {
	return &LogEventUseCase{
		repository:    repository,
		alertService:  alertService,
		idGenerator:   idGenerator,
		timeProvider:  timeProvider,
	}
}

// Execute выполняет создание лог записи
func (uc *LogEventUseCase) Execute(ctx context.Context, request LogEventRequest) (*LogEventResponse, error) {
	// Валидация запроса
	if err := uc.validateRequest(request); err != nil {
		return nil, err
	}
	
	// Создание лог записи
	logEntry := entities.LogEntry{
		ID:        uc.idGenerator.Generate(),
		Level:     request.Level,
		Service:   request.Service,
		Event:     request.Event,
		Timestamp: uc.timeProvider.Now(),
		UserID:    request.UserID,
		ChatID:    request.ChatID,
		Message:   request.Message,
		Metadata:  request.Metadata,
	}
	
	// Валидация лог записи
	if !logEntry.IsValid() {
		return nil, errors.ErrInvalidLogEntry
	}
	
	// Сохранение в репозиторий
	if err := uc.repository.Store(ctx, logEntry); err != nil {
		return nil, err
	}
	
	// Попытка отправки алерта (если нужен)
	alertSent := false
	if logEntry.ShouldAlert() {
		if err := uc.alertService.SendAlert(ctx, logEntry); err == nil {
			alertSent = true
		}
		// Не возвращаем ошибку, если алерт не отправился - лог уже сохранен
	}
	
	return &LogEventResponse{
		ID:        logEntry.ID,
		Timestamp: logEntry.Timestamp,
		Success:   true,
		AlertSent: alertSent,
	}, nil
}

// validateRequest валидирует входящий запрос
func (uc *LogEventUseCase) validateRequest(request LogEventRequest) error {
	if !request.Level.IsValid() {
		return errors.ErrInvalidLogEntry
	}
	
	if strings.TrimSpace(request.Service) == "" {
		return errors.ErrInvalidLogEntry
	}
	
	if strings.TrimSpace(request.Event) == "" {
		return errors.ErrInvalidLogEntry
	}
	
	if strings.TrimSpace(request.Message) == "" {
		return errors.ErrInvalidLogEntry
	}
	
	return nil
}