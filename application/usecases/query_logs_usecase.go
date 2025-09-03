package usecases

import (
	"context"
	
	"aviasales-shared-logging/domain/errors"
	"aviasales-shared-logging/domain/interfaces"
)

// QueryLogsUseCase обрабатывает поиск лог записей
type QueryLogsUseCase struct {
	repository interfaces.LogRepository
}

// NewQueryLogsUseCase создает новый экземпляр QueryLogsUseCase
func NewQueryLogsUseCase(repository interfaces.LogRepository) *QueryLogsUseCase {
	return &QueryLogsUseCase{
		repository: repository,
	}
}

// Execute выполняет поиск лог записей
func (uc *QueryLogsUseCase) Execute(ctx context.Context, request QueryLogsRequest) (*QueryLogsResponse, error) {
	// Валидация фильтра
	if err := uc.validateFilter(request.Filter); err != nil {
		return nil, err
	}
	
	// Применение значений по умолчанию
	filter := uc.applyDefaults(request.Filter)
	
	// Получение логов
	logs, err := uc.repository.Query(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	// Получение общего количества
	totalCount, err := uc.repository.Count(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	// Определение наличия дополнительных записей
	hasMore := int64(filter.Offset+len(logs)) < totalCount
	
	return &QueryLogsResponse{
		Logs:       logs,
		TotalCount: totalCount,
		HasMore:    hasMore,
	}, nil
}

// validateFilter валидирует параметры фильтра
func (uc *QueryLogsUseCase) validateFilter(filter interfaces.LogFilter) error {
	// Проверка лимита
	if filter.Limit < 0 {
		return errors.ErrInvalidFilter
	}
	
	if filter.Limit > 1000 { // Максимальный лимит
		return errors.ErrInvalidFilter
	}
	
	// Проверка смещения
	if filter.Offset < 0 {
		return errors.ErrInvalidFilter
	}
	
	// Проверка временного диапазона
	if filter.TimeFrom != nil && filter.TimeTo != nil {
		if filter.TimeTo.Before(*filter.TimeFrom) {
			return errors.ErrInvalidFilter
		}
	}
	
	return nil
}

// applyDefaults применяет значения по умолчанию к фильтру
func (uc *QueryLogsUseCase) applyDefaults(filter interfaces.LogFilter) interfaces.LogFilter {
	// Применяем лимит по умолчанию
	if filter.Limit == 0 {
		filter.Limit = 100
	}
	
	// Применяем сортировку по умолчанию
	if filter.SortBy == "" {
		filter.SortBy = "timestamp"
	}
	
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}
	
	return filter
}