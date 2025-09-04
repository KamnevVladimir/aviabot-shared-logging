# Dockerfile для тестирования shared logging библиотеки
FROM golang:1.21-alpine AS test

WORKDIR /app

# Копируем файлы модуля
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY *.go ./

# Запускаем тесты с покрытием
RUN go test -v -cover -coverprofile=coverage.out ./...

# Проверяем, что покрытие >= 90%
RUN go tool cover -func=coverage.out | grep total | awk '{if($3+0 < 90.0) exit 1}'

# Production образ (если потребуется для библиотеки)
FROM alpine:latest AS production

# Минимальная установка для Go модуля
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Копируем только необходимые файлы
COPY --from=test /app/go.mod /app/go.sum ./
COPY --from=test /app/*.go ./

LABEL version="2.0.0"
LABEL description="AviaBot Shared Logging Client"
LABEL maintainer="aviabot-team"
