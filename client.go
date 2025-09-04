package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client HTTP клиент для отправки логов в logging-service
type Client struct {
	baseURL     string
	serviceName string
	httpClient  *http.Client
}

// LogRequest структура запроса для отправки логов
type LogRequest struct {
	Level    string                 `json:"level"`
	Service  string                 `json:"service"`
	Event    string                 `json:"event"`
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewClient создает новый клиент для отправки логов
func NewClient(baseURL, serviceName string) *Client {
	return &Client{
		baseURL:     baseURL,
		serviceName: serviceName,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// sendLog отправляет лог в logging-service
func (c *Client) sendLog(level, event, message string, metadata map[string]interface{}) error {
	if c.baseURL == "" {
		return fmt.Errorf("logging client baseURL is empty")
	}

	payload := LogRequest{
		Level:    level,
		Service:  c.serviceName,
		Event:    event,
		Message:  message,
		Metadata: metadata,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal log payload: %w", err)
	}

	url := c.baseURL + "/log"
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send log to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("logging service returned status %d", resp.StatusCode)
	}

	return nil
}

// mergeMetadata объединяет метаданные
func (c *Client) mergeMetadata(base, additional map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	for k, v := range base {
		result[k] = v
	}
	
	for k, v := range additional {
		result[k] = v
	}
	
	return result
}
