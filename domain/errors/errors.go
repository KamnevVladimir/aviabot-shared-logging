package errors

import "errors"

// Domain errors для логирования сервиса
var (
	// ErrLogNotFound возвращается когда лог запись не найдена
	ErrLogNotFound = errors.New("log entry not found")

	// ErrInvalidLogEntry возвращается при некорректной лог записи
	ErrInvalidLogEntry = errors.New("invalid log entry")

	// ErrInvalidLogLevel возвращается при некорректном уровне логирования
	ErrInvalidLogLevel = errors.New("invalid log level")

	// ErrInvalidFilter возвращается при некорректном фильтре поиска
	ErrInvalidFilter = errors.New("invalid filter parameters")

	// ErrStorageUnavailable возвращается при недоступности хранилища
	ErrStorageUnavailable = errors.New("storage unavailable")

	// ErrAlertServiceUnavailable возвращается при недоступности сервиса алертов
	ErrAlertServiceUnavailable = errors.New("alert service unavailable")

	// ErrIDGenerationFailed возвращается при ошибке генерации ID
	ErrIDGenerationFailed = errors.New("ID generation failed")

	// ErrUnauthorized возвращается при попытке неавторизованного доступа
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrRateLimitExceeded возвращается при превышении лимита запросов
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)
