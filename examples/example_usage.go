package main

import (
	"errors"
	"time"

	"github.com/KamnevVladimir/aviabot-shared-logging"
)

// Пример интеграции logging клиента в микросервис
func main() {
	// Инициализация клиента
	logger := logging.NewClient("http://logging-service:8080", "example-service")

	// Service lifecycle
	logger.ServiceStart("v1.2.3", "Service started successfully")

	// Health monitoring
	healthMetadata := map[string]interface{}{
		"uptime":       3600,
		"memory_usage": "128MB",
		"cpu_usage":    "15%",
	}
	logger.Health("healthy", "All systems operational", healthMetadata)

	// Error handling
	err := errors.New("database connection timeout")
	errorMetadata := map[string]interface{}{
		"retry_count": 3,
		"timeout_ms":  5000,
	}
	logger.Error(err, "Failed to connect to database", errorMetadata)

	// HTTP requests logging
	duration := 150 * time.Millisecond
	httpMetadata := map[string]interface{}{
		"user_agent": "curl/7.68.0",
		"ip_address": "192.168.1.100",
	}
	logger.HTTPRequest("POST", "/api/users", 201, duration, httpMetadata)

	// External API calls
	apiDuration := 800 * time.Millisecond
	apiMetadata := map[string]interface{}{
		"request_id": "req-12345",
		"api_key":    "***",
	}
	logger.ExternalAPI("telegram", "https://api.telegram.org/getUpdates", 200, apiDuration, apiMetadata)

	// Service communication
	commDuration := 75 * time.Millisecond
	commMetadata := map[string]interface{}{
		"request_id": "comm-67890",
		"payload_size": "2.3KB",
	}
	logger.ServiceCommunication("gateway-service", "send_update", true, commDuration, commMetadata)

	// Warning example
	warningMetadata := map[string]interface{}{
		"latency_ms": 1200,
		"threshold":  1000,
	}
	logger.Warning("Slow response detected", warningMetadata)

	// Custom info event
	customMetadata := map[string]interface{}{
		"user_id": 12345,
		"action":  "login",
	}
	logger.Info("user_action", "User logged in successfully", customMetadata)

	// Service shutdown
	uptime := 7200 * time.Second // 2 hours
	logger.ServiceStop(uptime, "Graceful shutdown completed")
}
