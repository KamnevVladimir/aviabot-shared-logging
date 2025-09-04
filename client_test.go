package logging

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8080", "test-service")
	
	if client == nil {
		t.Fatal("client should not be nil")
	}
	
	if client.baseURL != "http://localhost:8080" {
		t.Errorf("expected baseURL http://localhost:8080, got %s", client.baseURL)
	}
	
	if client.serviceName != "test-service" {
		t.Errorf("expected serviceName test-service, got %s", client.serviceName)
	}
}

func TestClient_ServiceStart(t *testing.T) {
	var receivedPayload LogRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/log" {
			t.Errorf("expected path /log, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()
	
	client := NewClient(server.URL, "test-service")
	err := client.ServiceStart("v1.0.0", "service started successfully")
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if receivedPayload.Level != "INFO" {
		t.Errorf("expected level INFO, got %s", receivedPayload.Level)
	}
	
	if receivedPayload.Service != "test-service" {
		t.Errorf("expected service test-service, got %s", receivedPayload.Service)
	}
	
	if receivedPayload.Event != "service_start" {
		t.Errorf("expected event service_start, got %s", receivedPayload.Event)
	}
	
	if receivedPayload.Message != "service started successfully" {
		t.Errorf("expected message 'service started successfully', got %s", receivedPayload.Message)
	}
	
	if receivedPayload.Metadata["version"] != "v1.0.0" {
		t.Errorf("expected version v1.0.0, got %v", receivedPayload.Metadata["version"])
	}
}

func TestClient_ServiceStop(t *testing.T) {
	var receivedPayload LogRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()
	
	client := NewClient(server.URL, "test-service")
	uptime := 3600 * time.Second
	err := client.ServiceStop(uptime, "graceful shutdown")
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if receivedPayload.Level != "INFO" {
		t.Errorf("expected level INFO, got %s", receivedPayload.Level)
	}
	
	if receivedPayload.Event != "service_stop" {
		t.Errorf("expected event service_stop, got %s", receivedPayload.Event)
	}
	
	if receivedPayload.Metadata["uptime_seconds"] != float64(3600) {
		t.Errorf("expected uptime_seconds 3600, got %v", receivedPayload.Metadata["uptime_seconds"])
	}
}

func TestClient_Health(t *testing.T) {
	var receivedPayload LogRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()
	
	client := NewClient(server.URL, "test-service")
	metadata := map[string]interface{}{
		"uptime": 3600,
		"memory_usage": "50MB",
	}
	err := client.Health("healthy", "service is running smoothly", metadata)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if receivedPayload.Level != "INFO" {
		t.Errorf("expected level INFO, got %s", receivedPayload.Level)
	}
	
	if receivedPayload.Event != "health_check" {
		t.Errorf("expected event health_check, got %s", receivedPayload.Event)
	}
	
	if receivedPayload.Metadata["status"] != "healthy" {
		t.Errorf("expected status healthy, got %v", receivedPayload.Metadata["status"])
	}
	
	if receivedPayload.Metadata["uptime"] != float64(3600) {
		t.Errorf("expected uptime 3600, got %v", receivedPayload.Metadata["uptime"])
	}
}

func TestClient_Error(t *testing.T) {
	var receivedPayload LogRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()
	
	client := NewClient(server.URL, "test-service")
	testError := &testErr{msg: "connection failed"}
	metadata := map[string]interface{}{"retry_count": 3}
	
	err := client.Error(testError, "failed to connect to database", metadata)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if receivedPayload.Level != "ERROR" {
		t.Errorf("expected level ERROR, got %s", receivedPayload.Level)
	}
	
	if receivedPayload.Event != "error_event" {
		t.Errorf("expected event error_event, got %s", receivedPayload.Event)
	}
	
	if receivedPayload.Metadata["error"] != "connection failed" {
		t.Errorf("expected error 'connection failed', got %v", receivedPayload.Metadata["error"])
	}
	
	if receivedPayload.Metadata["retry_count"] != float64(3) {
		t.Errorf("expected retry_count 3, got %v", receivedPayload.Metadata["retry_count"])
	}
}

func TestClient_Warning(t *testing.T) {
	var receivedPayload LogRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()
	
	client := NewClient(server.URL, "test-service")
	metadata := map[string]interface{}{"latency_ms": 1200}
	
	err := client.Warning("slow response detected", metadata)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if receivedPayload.Level != "WARNING" {
		t.Errorf("expected level WARNING, got %s", receivedPayload.Level)
	}
	
	if receivedPayload.Event != "warning_event" {
		t.Errorf("expected event warning_event, got %s", receivedPayload.Event)
	}
}

func TestClient_Info(t *testing.T) {
	var receivedPayload LogRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()
	
	client := NewClient(server.URL, "test-service")
	metadata := map[string]interface{}{"user_id": 12345}
	
	err := client.Info("user_action", "user logged in", metadata)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if receivedPayload.Level != "INFO" {
		t.Errorf("expected level INFO, got %s", receivedPayload.Level)
	}
	
	if receivedPayload.Event != "user_action" {
		t.Errorf("expected event user_action, got %s", receivedPayload.Event)
	}
}

func TestClient_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	
	client := NewClient(server.URL, "test-service")
	err := client.Info("test_event", "test message", nil)
	
	if err == nil {
		t.Fatal("expected error for HTTP 500 response")
	}
}

func TestClient_InvalidURL(t *testing.T) {
	client := NewClient("", "test-service")
	err := client.Info("test_event", "test message", nil)
	
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

// Test helper
type testErr struct {
	msg string
}

func (e *testErr) Error() string {
	return e.msg
}
