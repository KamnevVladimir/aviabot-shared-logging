package logging

import (
	"fmt"
	"time"
)

// ServiceStart логирует запуск сервиса
func (c *Client) ServiceStart(version, message string) error {
	metadata := map[string]interface{}{
		"version": version,
	}
	return c.sendLog("INFO", "service_start", message, metadata)
}

// ServiceStop логирует остановку сервиса
func (c *Client) ServiceStop(uptime time.Duration, message string) error {
	metadata := map[string]interface{}{
		"uptime_seconds": uptime.Seconds(),
	}
	return c.sendLog("INFO", "service_stop", message, metadata)
}

// Health логирует состояние здоровья сервиса
func (c *Client) Health(status, message string, metadata map[string]interface{}) error {
	baseMetadata := map[string]interface{}{
		"status": status,
	}
	finalMetadata := c.mergeMetadata(baseMetadata, metadata)
	return c.sendLog("INFO", "health_check", message, finalMetadata)
}

// Error логирует ошибки
func (c *Client) Error(err error, message string, metadata map[string]interface{}) error {
	baseMetadata := map[string]interface{}{
		"error": err.Error(),
	}
	finalMetadata := c.mergeMetadata(baseMetadata, metadata)
	return c.sendLog("ERROR", "error_event", message, finalMetadata)
}

// Warning логирует предупреждения
func (c *Client) Warning(message string, metadata map[string]interface{}) error {
	return c.sendLog("WARNING", "warning_event", message, metadata)
}

// Info логирует информационные события
func (c *Client) Info(event, message string, metadata map[string]interface{}) error {
	return c.sendLog("INFO", event, message, metadata)
}

// Critical логирует критические события
func (c *Client) Critical(message string, metadata map[string]interface{}) error {
	return c.sendLog("CRITICAL", "critical_event", message, metadata)
}

// Debug логирует отладочную информацию
func (c *Client) Debug(message string, metadata map[string]interface{}) error {
	return c.sendLog("DEBUG", "debug_event", message, metadata)
}

// HTTPRequest логирует HTTP запросы
func (c *Client) HTTPRequest(method, path string, statusCode int, duration time.Duration, metadata map[string]interface{}) error {
	baseMetadata := map[string]interface{}{
		"method":       method,
		"path":         path,
		"status_code":  statusCode,
		"duration_ms":  duration.Milliseconds(),
	}
	finalMetadata := c.mergeMetadata(baseMetadata, metadata)
	message := fmt.Sprintf("%s %s - %d", method, path, statusCode)
	return c.sendLog("INFO", "http_request", message, finalMetadata)
}

// ExternalAPI логирует вызовы внешних API
func (c *Client) ExternalAPI(apiName, endpoint string, statusCode int, duration time.Duration, metadata map[string]interface{}) error {
	baseMetadata := map[string]interface{}{
		"api_name":     apiName,
		"endpoint":     endpoint,
		"status_code":  statusCode,
		"duration_ms":  duration.Milliseconds(),
	}
	finalMetadata := c.mergeMetadata(baseMetadata, metadata)
	message := fmt.Sprintf("API call to %s", apiName)
	return c.sendLog("INFO", "external_api", message, finalMetadata)
}

// ServiceCommunication логирует взаимодействие между сервисами
func (c *Client) ServiceCommunication(targetService, operation string, success bool, duration time.Duration, metadata map[string]interface{}) error {
	baseMetadata := map[string]interface{}{
		"target_service": targetService,
		"operation":      operation,
		"success":        success,
		"duration_ms":    duration.Milliseconds(),
	}
	finalMetadata := c.mergeMetadata(baseMetadata, metadata)
	message := fmt.Sprintf("Communication with %s: %s", targetService, operation)
	
	level := "INFO"
	if !success {
		level = "ERROR"
	}
	
	return c.sendLog(level, "service_communication", message, finalMetadata)
}
