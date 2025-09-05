# AviaBot Shared Logging Client

Простой и типизированный **Go клиент** для отправки логов в `aviabot-logging-service`.

> ⚠️ **Важно**: Это **Go библиотека**, а не микросервис! Не разворачивается в Railway, а импортируется в другие сервисы через `go get`.

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

## 🏗️ Архитектура системы логирования

### Взаимодействие компонентов

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ telegram-poller │    │   gateway-service│    │  search-service │
│                 │    │                  │    │                 │
│ import logging  │    │ import logging   │    │ import logging  │
│ logger.Health() │    │ logger.Info()    │    │ logger.Error()  │
└─────────┬───────┘    └─────────┬────────┘    └─────────┬───────┘
          │                      │                       │
          │ POST /log            │ POST /log             │ POST /log
          │ {level, service,     │ {level, service,      │ {level, service,
          │  event, message}     │  event, message}      │  event, message}
          │                      │                       │
          └──────────────────────┼───────────────────────┘
                                 │
                                 ▼
                    ┌─────────────────────────┐
                    │  aviabot-logging-service│
                    │  (Railway deployment)   │
                    │                         │
                    │  ┌─────────────────────┐│
                    │  │ HTTP API            ││
                    │  │ POST /log           ││
                    │  │ GET /logs           ││
                    │  │ GET /health         ││
                    │  └─────────────────────┘│
                    │           │             │
                    │           ▼             │
                    │  ┌─────────────────────┐│
                    │  │ PostgreSQL          ││
                    │  │ (log storage)       ││
                    │  └─────────────────────┘│
                    │           │             │
                    │           ▼             │
                    │  ┌─────────────────────┐│
                    │  │ Grafana Integration ││
                    │  │ (dashboards)        ││
                    │  └─────────────────────┘│
                    └─────────────────────────┘
```

### Роли компонентов

#### **aviabot-shared-logging** (эта библиотека)
- ✅ **Go module** - импортируется как зависимость
- ✅ **HTTP клиент** - отправляет логи по HTTP
- ✅ **Типизированный API** - стандартизированные методы
- ❌ **НЕ микросервис** - не разворачивается отдельно
- ❌ **НЕ в Railway** - живет только в коде других сервисов

#### **aviabot-logging-service** (отдельный микросервис)
- ✅ **HTTP сервер** - принимает POST /log запросы
- ✅ **База данных** - сохраняет логи в PostgreSQL
- ✅ **Railway deployment** - развернут как отдельный сервис
- ✅ **Grafana integration** - создает дашборды для мониторинга

## 🤔 Зачем понадобился shared-logging?

### ❌ **Проблемы ДО внедрения:**

```go
// В каждом сервисе приходилось писать свой httpLogger
type httpLogger struct{ endpoint string }

func (l *httpLogger) Log(level string, event string, message string, metadata map[string]interface{}) {
    // 🚫 Дублирование кода в каждом сервисе!
    type payload struct {
        Level    string                 `json:"level"`
        Service  string                 `json:"service"`
        Event    string                 `json:"event"`
        Message  string                 `json:"message"`
        Metadata map[string]interface{} `json:"metadata,omitempty"`
    }
    p := payload{Level: level, Service: "service-name", Event: event, Message: message, Metadata: metadata}
    b, _ := json.Marshal(p) // 🚫 Возможные ошибки маршалинга
    http.Post(l.endpoint+"/log", "application/json", bytes.NewReader(b))
}
```

**Проблемы этого подхода:**
- 🔴 **Дублирование кода** в каждом микросервисе (telegram-poller, gateway, search, etc.)
- 🔴 **Ошибки в JSON структурах** - легко опечататься в поле
- 🔴 **Нет типизации** - можно отправить level="INFOO" и не заметить
- 🔴 **Сложно поддерживать** - изменения API нужно делать в 5+ местах
- 🔴 **Нет стандартизации** событий - каждый сервис изобретает свои event типы
- 🔴 **Нужны тесты** на логирование в каждом сервисе

### ✅ **Решение ПОСЛЕ внедрения:**

```go
// В каждом сервисе теперь просто:
import "github.com/KamnevVladimir/aviabot-shared-logging"

logger := logging.NewClient(cfg.LoggingURL, "telegram-poller")
logger.ServiceStart("v1.0.0", "service started")  // ✅ Типизированно!
logger.Health("healthy", "all ok", metadata)      // ✅ Стандартизированно!
logger.Error(err, "failed", metadata)             // ✅ Без ошибок в JSON!
```

**Преимущества:**
- ✅ **DRY принцип** - код логирования написан один раз
- ✅ **Типизация** - компилятор найдет ошибки на этапе сборки
- ✅ **Стандартизация** - все сервисы используют одинаковые event типы
- ✅ **Простота обновлений** - изменили клиент, обновили версию везде
- ✅ **Нет тестов на логирование** в каждом сервисе (клиент уже протестирован на 96%)
- ✅ **Централизованная документация** API

## 📦 Установка и развертывание

### 1️⃣ **Установка в микросервисах**

```bash
# В каждом сервисе (telegram-poller, gateway-service, search-service, etc.)
cd telegram-poller/
go get github.com/KamnevVladimir/aviabot-shared-logging
```

### 2️⃣ **В go.mod появится зависимость:**

```go
module aviasales-bot/telegram-poller

require (
    github.com/KamnevVladimir/aviabot-shared-logging v2.0.0
    github.com/redis/go-redis/v9 v9.0.0
    // другие зависимости...
)
```

### 3️⃣ **Использование в коде:**

```go
package main

import (
    "context"
    "time"
    "github.com/KamnevVladimir/aviabot-shared-logging"
)

func main() {
    // Инициализация в каждом сервисе
    logger := logging.NewClient(cfg.LoggingURL, "telegram-poller")
    
    // Service lifecycle
    logger.ServiceStart("v1.0.0", "telegram poller started")
    
    // В главном цикле
    for {
        // Health monitoring
        logger.Health("healthy", "polling active", map[string]interface{}{
            "uptime": time.Since(startTime).Seconds(),
            "last_update": lastUpdateTime,
        })
        
        // При ошибках
        if err != nil {
            logger.Error(err, "polling failed", map[string]interface{}{
                "retry_count": retryCount,
                "endpoint": telegramAPI,
            })
        }
        
        // External API calls
        duration := time.Since(apiCallStart)
        logger.ExternalAPI("telegram", "getUpdates", 200, duration, metadata)
        
        time.Sleep(pollInterval)
    }
}
```

## 🎯 **Railway развертывание**

### ❌ **shared-logging НЕ разворачивается:**
- Это Go библиотека, а не сервис
- Живет только в коде других микросервисов
- Не имеет своего main.go или HTTP сервера

### ✅ **logging-service разворачивается:**
```yaml
# railway.toml для aviabot-logging-service
[build]
builder = "dockerfile"

[deploy]
startCommand = "./main"

[env]
PORT = "8080"
DATABASE_URL = "${Postgres.DATABASE_URL}"
GRAFANA_URL = "http://grafana.railway.internal:3000"
```

## 🔄 **Жизненный цикл логирования**

```
1. Событие в микросервисе
   ↓
2. logger.Health("healthy", message, metadata)
   ↓
3. shared-logging формирует HTTP запрос
   ↓
4. POST http://logging-service:8080/log
   {
     "level": "INFO",
     "service": "telegram-poller", 
     "event": "health_check",
     "message": "service is healthy",
     "metadata": {"uptime": 3600}
   }
   ↓
5. logging-service обрабатывает запрос
   ↓
6. Сохранение в PostgreSQL
   ↓
7. Отображение в Grafana дашбордах
```

## 🚀 **Примеры интеграции**

### telegram-poller
```go
logger := logging.NewClient(cfg.LoggingURL, "telegram-poller")
logger.ServiceStart("v1.0.0", "poller started")
logger.ExternalAPI("telegram", telegramURL, 200, duration, metadata)
logger.ServiceCommunication("gateway", "send_update", true, duration, metadata)
```

### gateway-service  
```go
logger := logging.NewClient(cfg.LoggingURL, "gateway-service")
logger.HTTPRequest("POST", "/ingest/telegram", 200, duration, metadata)
logger.Info("user_action", "message processed", metadata)
```

### search-service
```go
logger := logging.NewClient(cfg.LoggingURL, "search-service")
logger.ExternalAPI("aviasales", apiEndpoint, 200, duration, metadata)
logger.Warning("slow response", map[string]interface{}{"latency": 1200})
```

## 📋 Changelog

- **v2.0.0** - Полная переработка с типизированным API
- **v1.0.3** - Legacy версия (deprecated)
