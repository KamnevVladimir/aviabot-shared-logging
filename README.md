# AviaBot Shared Logging Client

Простой и типизированный клиент для отправки логов в `aviabot-logging-service`.

## 🚀 Основные преимущества

- **Простота использования**: `logger.Health("healthy", "all systems ok", metadata)`
- **Типизация**: Никаких ручных JSON структур
- **Переиспользование**: Один раз подключили, везде используем
- **Нет необходимости в тестах**: Клиент протестирован на 96%

## 📦 Установка

```bash
go get github.com/KamnevVladimir/aviabot-shared-logging
```

## 🔧 Использование

### Базовая инициализация

```go
import "github.com/KamnevVladimir/aviabot-shared-logging"

// Создание клиента
logger := logging.NewClient("http://logging-service:8080", "my-service")
```

### Service Lifecycle Events

```go
// Запуск сервиса
logger.ServiceStart("v1.2.3", "service started successfully")

// Остановка сервиса
uptime := time.Since(startTime)
logger.ServiceStop(uptime, "graceful shutdown completed")
```

### Health Monitoring

```go
// Отчет о состоянии здоровья
metadata := map[string]interface{}{
    "uptime": 3600,
    "memory_usage": "128MB",
    "cpu_usage": "15%",
}
logger.Health("healthy", "all systems operational", metadata)
```

### Error & Warning Logging

```go
// Логирование ошибок
err := errors.New("connection timeout")
metadata := map[string]interface{}{"retry_count": 3}
logger.Error(err, "failed to connect to database", metadata)

// Предупреждения
metadata := map[string]interface{}{"latency_ms": 1200}
logger.Warning("slow response detected", metadata)
```

### HTTP Requests

```go
// Логирование HTTP запросов
duration := 150 * time.Millisecond
metadata := map[string]interface{}{"user_agent": "curl/7.68.0"}
logger.HTTPRequest("POST", "/api/users", 201, duration, metadata)
```

### External API Calls

```go
// Вызовы внешних API
duration := 800 * time.Millisecond
metadata := map[string]interface{}{"request_id": "req-123"}
logger.ExternalAPI("telegram", "https://api.telegram.org/getUpdates", 200, duration, metadata)
```

### Service Communication

```go
// Взаимодействие между сервисами
duration := 75 * time.Millisecond
metadata := map[string]interface{}{"request_id": "req-456"}
logger.ServiceCommunication("gateway-service", "send_update", true, duration, metadata)
```

### Общие методы

```go
// Информационные сообщения
logger.Info("user_action", "user logged in", metadata)

// Отладочная информация
logger.Debug("variable state", metadata)

// Критические события
logger.Critical("system failure", metadata)
```

## 📊 API Reference

### Client Methods

| Метод | Уровень | Event Type | Описание |
|-------|---------|------------|----------|
| `ServiceStart(version, message)` | INFO | service_start | Запуск сервиса |
| `ServiceStop(uptime, message)` | INFO | service_stop | Остановка сервиса |
| `Health(status, message, metadata)` | INFO | health_check | Состояние здоровья |
| `Error(err, message, metadata)` | ERROR | error_event | Ошибки |
| `Warning(message, metadata)` | WARNING | warning_event | Предупреждения |
| `Info(event, message, metadata)` | INFO | custom | Информационные события |
| `Critical(message, metadata)` | CRITICAL | critical_event | Критические события |
| `Debug(message, metadata)` | DEBUG | debug_event | Отладочная информация |
| `HTTPRequest(method, path, status, duration, metadata)` | INFO | http_request | HTTP запросы |
| `ExternalAPI(api, endpoint, status, duration, metadata)` | INFO | external_api | Внешние API |
| `ServiceCommunication(service, op, success, duration, metadata)` | INFO/ERROR | service_communication | Связь сервисов |

## ✅ Тестирование

```bash
# Запуск тестов
go test -v

# Проверка покрытия
go test -cover
# coverage: 96.0% of statements
```

## 🏗️ Архитектура

```
aviabot-shared-logging/
├── client.go          # HTTP клиент и core функции
├── events.go          # Типизированные методы для событий
├── client_test.go     # Тесты клиента
├── events_test.go     # Тесты событий
└── go.mod            # Module definition
```

## 🔄 Интеграция

### В telegram-poller

```go
import "github.com/KamnevVladimir/aviabot-shared-logging"

logger := logging.NewClient(cfg.LoggingURL, "telegram-poller")
logger.ServiceStart("v1.0.0", "poller started")
logger.Health("healthy", "polling active", map[string]interface{}{"uptime": 3600})
```

### В других микросервисах

```go
// gateway-service
logger := logging.NewClient(cfg.LoggingURL, "gateway-service")
logger.Info("user_action", "user search request", metadata)

// search-service
logger := logging.NewClient(cfg.LoggingURL, "search-service")
logger.ExternalAPI("aviasales", endpoint, 200, duration, metadata)
```

## ⚡ Performance

- **HTTP timeout**: 10 секунд
- **Async logging**: Не блокирует основной поток
- **Error handling**: Graceful fallback при недоступности logging-service

## 📋 Changelog

- **v2.0.0** - Полная переработка с типизированным API
- **v1.0.3** - Legacy версия (deprecated)
