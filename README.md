# Aviabot Shared Logging

Библиотека для структурированного логирования в микросервисах Aviabot.

## Содержимое

- `domain/entities` - Сущности логирования (LogEntry, LogLevel)
- `domain/errors` - Ошибки логирования
- `domain/interfaces` - Интерфейсы репозиториев
- `application/usecases` - Use cases для логирования
- `infrastructure/http` - HTTP handlers для логирования

## Использование

```go
import (
    "github.com/KamnevVladimir/aviabot-shared-logging/domain/entities"
    "github.com/KamnevVladimir/aviabot-shared-logging/application/usecases"
)
```

## Особенности

- Структурированные JSON логи
- Поддержка различных уровней логирования
- Метрики и алерты
- Railway-совместимый формат

## Установка

```bash
go get github.com/KamnevVladimir/aviabot-shared-logging@v1.0.0
```
