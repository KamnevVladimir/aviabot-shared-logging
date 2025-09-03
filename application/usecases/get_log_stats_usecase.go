package usecases

import (
	"context"
	
	"aviasales-shared-logging/domain/errors"
	"aviasales-shared-logging/domain/interfaces"
)

// GetLogStatsUseCase обрабатывает получение статистики логирования
type GetLogStatsUseCase struct {
	repository interfaces.LogRepository
}

// NewGetLogStatsUseCase создает новый экземпляр GetLogStatsUseCase
func NewGetLogStatsUseCase(repository interfaces.LogRepository) *GetLogStatsUseCase {
	return &GetLogStatsUseCase{
		repository: repository,
	}
}

// Execute выполняет получение статистики логирования
func (uc *GetLogStatsUseCase) Execute(ctx context.Context, request GetLogStatsRequest) (*GetLogStatsResponse, error) {
	// Валидация фильтра
	if err := uc.validateFilter(request.Filter); err != nil {
		return nil, err
	}
	
	// Получение статистики из репозитория
	stats, err := uc.repository.GetStats(ctx, request.Filter)
	if err != nil {
		return nil, err
	}
	
	return &GetLogStatsResponse{
		Stats: *stats,
	}, nil
}

// validateFilter валидирует параметры фильтра для статистики
func (uc *GetLogStatsUseCase) validateFilter(filter interfaces.LogFilter) error {
	// Проверка временного диапазона
	if filter.TimeFrom != nil && filter.TimeTo != nil {
		if filter.TimeTo.Before(*filter.TimeFrom) {
			return errors.ErrInvalidFilter
		}
	}
	
	return nil
}