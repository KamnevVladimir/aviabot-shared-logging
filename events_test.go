package logging

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_Critical(t *testing.T) {
	var receivedPayload LogRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-service")
	metadata := map[string]interface{}{"severity": "high"}

	err := client.Critical("system critical failure", metadata)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedPayload.Level != "CRITICAL" {
		t.Errorf("expected level CRITICAL, got %s", receivedPayload.Level)
	}

	if receivedPayload.Event != "critical_event" {
		t.Errorf("expected event critical_event, got %s", receivedPayload.Event)
	}
}

func TestClient_Debug(t *testing.T) {
	var receivedPayload LogRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-service")
	metadata := map[string]interface{}{"variable": "test_value"}

	err := client.Debug("debug information", metadata)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedPayload.Level != "DEBUG" {
		t.Errorf("expected level DEBUG, got %s", receivedPayload.Level)
	}

	if receivedPayload.Event != "debug_event" {
		t.Errorf("expected event debug_event, got %s", receivedPayload.Event)
	}
}

func TestClient_HTTPRequest(t *testing.T) {
	var receivedPayload LogRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-service")
	duration := 150 * time.Millisecond
	metadata := map[string]interface{}{"user_agent": "test-agent"}

	err := client.HTTPRequest("POST", "/api/test", 201, duration, metadata)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedPayload.Level != "INFO" {
		t.Errorf("expected level INFO, got %s", receivedPayload.Level)
	}

	if receivedPayload.Event != "http_request" {
		t.Errorf("expected event http_request, got %s", receivedPayload.Event)
	}

	if receivedPayload.Message != "POST /api/test - 201" {
		t.Errorf("expected message 'POST /api/test - 201', got %s", receivedPayload.Message)
	}

	if receivedPayload.Metadata["method"] != "POST" {
		t.Errorf("expected method POST, got %v", receivedPayload.Metadata["method"])
	}

	if receivedPayload.Metadata["duration_ms"] != float64(150) {
		t.Errorf("expected duration_ms 150, got %v", receivedPayload.Metadata["duration_ms"])
	}

	if receivedPayload.Metadata["user_agent"] != "test-agent" {
		t.Errorf("expected user_agent test-agent, got %v", receivedPayload.Metadata["user_agent"])
	}
}

func TestClient_ExternalAPI(t *testing.T) {
	var receivedPayload LogRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-service")
	duration := 1200 * time.Millisecond
	metadata := map[string]interface{}{"request_id": "req-123"}

	err := client.ExternalAPI("telegram", "https://api.telegram.org/getUpdates", 200, duration, metadata)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedPayload.Level != "INFO" {
		t.Errorf("expected level INFO, got %s", receivedPayload.Level)
	}

	if receivedPayload.Event != "external_api" {
		t.Errorf("expected event external_api, got %s", receivedPayload.Event)
	}

	if receivedPayload.Message != "API call to telegram" {
		t.Errorf("expected message 'API call to telegram', got %s", receivedPayload.Message)
	}

	if receivedPayload.Metadata["api_name"] != "telegram" {
		t.Errorf("expected api_name telegram, got %v", receivedPayload.Metadata["api_name"])
	}

	if receivedPayload.Metadata["duration_ms"] != float64(1200) {
		t.Errorf("expected duration_ms 1200, got %v", receivedPayload.Metadata["duration_ms"])
	}
}

func TestClient_ServiceCommunication_Success(t *testing.T) {
	var receivedPayload LogRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-service")
	duration := 75 * time.Millisecond
	metadata := map[string]interface{}{"request_id": "req-456"}

	err := client.ServiceCommunication("gateway-service", "send_update", true, duration, metadata)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedPayload.Level != "INFO" {
		t.Errorf("expected level INFO, got %s", receivedPayload.Level)
	}

	if receivedPayload.Event != "service_communication" {
		t.Errorf("expected event service_communication, got %s", receivedPayload.Event)
	}

	if receivedPayload.Message != "Communication with gateway-service: send_update" {
		t.Errorf("expected specific message, got %s", receivedPayload.Message)
	}

	if receivedPayload.Metadata["success"] != true {
		t.Errorf("expected success true, got %v", receivedPayload.Metadata["success"])
	}
}

func TestClient_ServiceCommunication_Failure(t *testing.T) {
	var receivedPayload LogRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-service")
	duration := 2000 * time.Millisecond
	metadata := map[string]interface{}{"error": "timeout"}

	err := client.ServiceCommunication("gateway-service", "send_update", false, duration, metadata)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedPayload.Level != "ERROR" {
		t.Errorf("expected level ERROR for failed communication, got %s", receivedPayload.Level)
	}

	if receivedPayload.Event != "service_communication" {
		t.Errorf("expected event service_communication, got %s", receivedPayload.Event)
	}

	if receivedPayload.Metadata["success"] != false {
		t.Errorf("expected success false, got %v", receivedPayload.Metadata["success"])
	}
}

func TestClient_MergeMetadata(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-service")

	base := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	additional := map[string]interface{}{
		"key2": "new_value2", // Should override
		"key3": "value3",     // Should add
	}

	result := client.mergeMetadata(base, additional)

	if result["key1"] != "value1" {
		t.Errorf("expected key1 to be value1, got %v", result["key1"])
	}

	if result["key2"] != "new_value2" {
		t.Errorf("expected key2 to be overridden to new_value2, got %v", result["key2"])
	}

	if result["key3"] != "value3" {
		t.Errorf("expected key3 to be value3, got %v", result["key3"])
	}
}

func TestClient_MergeMetadata_NilBase(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-service")

	additional := map[string]interface{}{
		"key1": "value1",
	}

	result := client.mergeMetadata(nil, additional)

	if result["key1"] != "value1" {
		t.Errorf("expected key1 to be value1, got %v", result["key1"])
	}
}

func TestClient_MergeMetadata_NilAdditional(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-service")

	base := map[string]interface{}{
		"key1": "value1",
	}

	result := client.mergeMetadata(base, nil)

	if result["key1"] != "value1" {
		t.Errorf("expected key1 to be value1, got %v", result["key1"])
	}
}
