package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aviasales-shared-logging/application/usecases"
	"aviasales-shared-logging/domain/entities"
	domainerrors "aviasales-shared-logging/domain/errors"
	"aviasales-shared-logging/domain/interfaces"
)

// LogsHandler обрабатывает HTTP запросы для логирования
type LogsHandler struct {
	logEventUseCase    LogEventUseCase
	queryLogsUseCase   QueryLogsUseCase
	getLogStatsUseCase GetLogStatsUseCase
}

// Use case interfaces
type LogEventUseCase interface {
	Execute(ctx context.Context, request usecases.LogEventRequest) (*usecases.LogEventResponse, error)
}

type QueryLogsUseCase interface {
	Execute(ctx context.Context, request usecases.QueryLogsRequest) (*usecases.QueryLogsResponse, error)
}

type GetLogStatsUseCase interface {
	Execute(ctx context.Context, request usecases.GetLogStatsRequest) (*usecases.GetLogStatsResponse, error)
}

// NewLogsHandler создает новый экземпляр LogsHandler
func NewLogsHandler(
	logEventUseCase LogEventUseCase,
	queryLogsUseCase QueryLogsUseCase,
	getLogStatsUseCase GetLogStatsUseCase,
) *LogsHandler {
	return &LogsHandler{
		logEventUseCase:    logEventUseCase,
		queryLogsUseCase:   queryLogsUseCase,
		getLogStatsUseCase: getLogStatsUseCase,
	}
}

// CreateLog обрабатывает POST /logs - создание лог записи
func (h *LogsHandler) CreateLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Парсинг JSON запроса
	var request struct {
		Level    string                 `json:"level"`
		Service  string                 `json:"service"`
		Event    string                 `json:"event"`
		Message  string                 `json:"message"`
		UserID   *int64                 `json:"user_id,omitempty"`
		ChatID   *int64                 `json:"chat_id,omitempty"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Парсинг уровня логирования
	level, err := h.parseLogLevel(request.Level)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid log level")
		return
	}

	// Создание запроса для use case
	useCaseRequest := usecases.LogEventRequest{
		Level:    level,
		Service:  request.Service,
		Event:    request.Event,
		Message:  request.Message,
		UserID:   request.UserID,
		ChatID:   request.ChatID,
		Metadata: request.Metadata,
	}

	// Выполнение use case
	response, err := h.logEventUseCase.Execute(r.Context(), useCaseRequest)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Отправка успешного ответа
	h.writeJSONResponse(w, http.StatusCreated, response)
}

// GetLogs обрабатывает GET /logs - получение логов с фильтрацией
func (h *LogsHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Парсинг query параметров
	filter, err := h.parseQueryFilters(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Создание запроса для use case
	useCaseRequest := usecases.QueryLogsRequest{
		Filter: filter,
	}

	// Выполнение use case
	response, err := h.queryLogsUseCase.Execute(r.Context(), useCaseRequest)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Отправка успешного ответа
	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetStats обрабатывает GET /logs/stats - получение статистики
func (h *LogsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Парсинг query параметров (используем те же фильтры)
	filter, err := h.parseQueryFilters(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Создание запроса для use case
	useCaseRequest := usecases.GetLogStatsRequest{
		Filter: filter,
	}

	// Выполнение use case
	response, err := h.getLogStatsUseCase.Execute(r.Context(), useCaseRequest)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Отправка успешного ответа
	h.writeJSONResponse(w, http.StatusOK, response)
}

// parseLogLevel преобразует строку в LogLevel
func (h *LogsHandler) parseLogLevel(levelStr string) (entities.LogLevel, error) {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return entities.LogLevelDebug, nil
	case "INFO":
		return entities.LogLevelInfo, nil
	case "WARNING", "WARN":
		return entities.LogLevelWarning, nil
	case "ERROR":
		return entities.LogLevelError, nil
	case "CRITICAL", "CRIT":
		return entities.LogLevelCritical, nil
	default:
		return 0, domainerrors.ErrInvalidLogLevel
	}
}

// parseQueryFilters парсит URL query параметры в LogFilter
func (h *LogsHandler) parseQueryFilters(r *http.Request) (interfaces.LogFilter, error) {
	query := r.URL.Query()
	filter := interfaces.LogFilter{}

	// Парсинг limit
	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return filter, errors.New("Invalid limit parameter")
		}
		filter.Limit = limit
	}

	// Парсинг offset
	if offsetStr := query.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return filter, errors.New("Invalid offset parameter")
		}
		filter.Offset = offset
	}

	// Парсинг services (может быть несколько)
	if services := query["service"]; len(services) > 0 {
		filter.Services = services
	}

	// Парсинг events (может быть несколько)
	if events := query["event"]; len(events) > 0 {
		filter.Events = events
	}

	// Парсинг levels (может быть несколько)
	if levelStrs := query["level"]; len(levelStrs) > 0 {
		levels := make([]entities.LogLevel, 0, len(levelStrs))
		for _, levelStr := range levelStrs {
			level, err := h.parseLogLevel(levelStr)
			if err != nil {
				return filter, errors.New("Invalid level parameter")
			}
			levels = append(levels, level)
		}
		filter.Levels = levels
	}

	// Парсинг user_id
	if userIDStr := query.Get("user_id"); userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			return filter, errors.New("Invalid user_id parameter")
		}
		filter.UserID = &userID
	}

	// Парсинг chat_id
	if chatIDStr := query.Get("chat_id"); chatIDStr != "" {
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			return filter, errors.New("Invalid chat_id parameter")
		}
		filter.ChatID = &chatID
	}

	// Парсинг message_contains
	if messageContains := query.Get("message_contains"); messageContains != "" {
		filter.MessageContains = messageContains
	}

	// Парсинг time_from
	if timeFromStr := query.Get("time_from"); timeFromStr != "" {
		timeFrom, err := time.Parse(time.RFC3339, timeFromStr)
		if err != nil {
			return filter, errors.New("Invalid time_from parameter (use RFC3339 format)")
		}
		filter.TimeFrom = &timeFrom
	}

	// Парсинг time_to
	if timeToStr := query.Get("time_to"); timeToStr != "" {
		timeTo, err := time.Parse(time.RFC3339, timeToStr)
		if err != nil {
			return filter, errors.New("Invalid time_to parameter (use RFC3339 format)")
		}
		filter.TimeTo = &timeTo
	}

	// Парсинг sort_by
	if sortBy := query.Get("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	}

	// Парсинг sort_order
	if sortOrder := query.Get("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}

	return filter, nil
}

// handleUseCaseError обрабатывает ошибки от use cases
func (h *LogsHandler) handleUseCaseError(w http.ResponseWriter, err error) {
	switch err {
	case domainerrors.ErrInvalidLogEntry:
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid log entry")
	case domainerrors.ErrInvalidFilter:
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid filter parameters")
	case domainerrors.ErrStorageUnavailable:
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Storage unavailable")
	case domainerrors.ErrUnauthorized:
		h.writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized access")
	case domainerrors.ErrRateLimitExceeded:
		h.writeErrorResponse(w, http.StatusTooManyRequests, "Rate limit exceeded")
	default:
		h.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
	}
}

// writeJSONResponse записывает JSON ответ
func (h *LogsHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Если не удалось закодировать ответ, отправляем ошибку
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// writeErrorResponse записывает JSON ответ с ошибкой
func (h *LogsHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := map[string]interface{}{
		"error":   message,
		"success": false,
	}

	h.writeJSONResponse(w, statusCode, errorResponse)
}

// HealthHandler обрабатывает health check запросы
type HealthHandler struct {
	version string
}

// NewHealthHandler создает новый экземпляр HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		version: "1.0.0",
	}
}

// Check обрабатывает GET /health - health check
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	healthResponse := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   h.version,
		"service":   "logging-service",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(healthResponse); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
