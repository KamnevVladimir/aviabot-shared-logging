# Shared Logging Library

Библиотека для централизованного логирования.

## Назначение

- HTTP клиент для отправки логов
- Структуры данных для логов
- Use cases для работы с логами
- HTTP handlers для logging-service

## Содержимое

- `domain/entities/` - структуры логов
- `application/usecases/` - бизнес-логика
- `infrastructure/http/` - HTTP handlers

## Использование

```go
import "github.com/KamnevVladimir/aviabot-shared-logging/infrastructure/http"

// Создание HTTP клиента
client := http.NewLogClient("http://logging-service:8080")
client.Log("INFO", "service-name", "event", "message", metadata)
```

## Версионирование

- v1.0.3 - текущая версия
- Используется во всех микросервисах
