package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	config := ServerConfig{
		Port: 8080,
	}

	server := NewServer(config)
	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.config.Port != 8080 {
		t.Errorf("Port = %d, want 8080", server.config.Port)
	}

	if server.migrations == nil {
		t.Error("migrations map should be initialized")
	}
}

func TestNewServerEmptyConfig(t *testing.T) {
	server := NewServer(ServerConfig{})
	if server == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestServerRouter(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})

	router := server.Router()
	if router == nil {
		t.Fatal("Router returned nil")
	}
}

func TestServerSetupRouter(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})

	// Router should be set up after NewServer
	router := server.Router()
	if router == nil {
		t.Fatal("Router should be initialized")
	}

	// Test that routes are registered by making test requests
	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/new"},
		{"GET", "/config"},
		{"GET", "/api/health"},
		{"GET", "/api/migrations"},
		{"GET", "/api/config"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			// Should not return 404 (route not found)
			if rec.Code == http.StatusNotFound {
				t.Errorf("Route %s %s returned 404", tt.method, tt.path)
			}
		})
	}
}

func TestServerServeIndex(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Header().Get("Content-Type") != "text/html" {
		t.Errorf("Content-Type = %q, want %q", rec.Header().Get("Content-Type"), "text/html")
	}
}

func TestServerServeNewMigration(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/new", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Header().Get("Content-Type") != "text/html" {
		t.Errorf("Content-Type = %q, want %q", rec.Header().Get("Content-Type"), "text/html")
	}
}

func TestServerServeConfig(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Header().Get("Content-Type") != "text/html" {
		t.Errorf("Content-Type = %q, want %q", rec.Header().Get("Content-Type"), "text/html")
	}
}

func TestServerServeMigration(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/migration/test-id", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Header().Get("Content-Type") != "text/html" {
		t.Errorf("Content-Type = %q, want %q", rec.Header().Get("Content-Type"), "text/html")
	}
}

func TestServerHandleHealth(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Success should be true")
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Data should be a map")
	}

	if data["status"] != "ok" {
		t.Errorf("status = %v, want ok", data["status"])
	}
}

func TestServerHandleListMigrations(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/api/migrations", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Success should be true")
	}

	// Data should be an array
	_, ok := response.Data.([]interface{})
	if !ok {
		t.Error("Data should be an array")
	}
}

func TestServerHandleListMigrationsWithData(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	// Add a migration manually
	server.mu.Lock()
	server.migrations["test-id"] = &MigrationStatus{
		ID:     "test-id",
		Status: "completed",
	}
	server.mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/migrations", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data, ok := response.Data.([]interface{})
	if !ok {
		t.Fatal("Data should be an array")
	}

	if len(data) != 1 {
		t.Errorf("Data length = %d, want 1", len(data))
	}
}

func TestServerHandleStartMigration(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	migrationReq := StartMigrationRequest{
		SourceType: "cvs",
		SourcePath: "/tmp/test-cvs",
		TargetPath: "/tmp/test-git",
		Options: map[string]interface{}{
			"dryRun": true,
		},
	}

	body, err := json.Marshal(migrationReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/migrations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Success should be true")
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Data should be a map")
	}

	if _, ok := data["id"]; !ok {
		t.Error("Response should contain id")
	}
}

func TestServerHandleStartMigrationInvalidJSON(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	req := httptest.NewRequest(http.MethodPost, "/api/migrations", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Success {
		t.Error("Success should be false")
	}

	if response.Error == nil {
		t.Error("Error should not be nil")
	}
}

func TestServerHandleStartMigrationMissingFields(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	tests := []struct {
		name string
		req  StartMigrationRequest
	}{
		{
			name: "missing source type",
			req: StartMigrationRequest{
				SourcePath: "/tmp/test",
				TargetPath: "/tmp/test",
			},
		},
		{
			name: "missing source path",
			req: StartMigrationRequest{
				SourceType: "cvs",
				TargetPath: "/tmp/test",
			},
		},
		{
			name: "missing target path",
			req: StartMigrationRequest{
				SourceType: "cvs",
				SourcePath: "/tmp/test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.req)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/migrations", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("Status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestServerHandleGetMigration(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	// Create a migration first
	server.mu.Lock()
	server.migrations["test-id-123"] = &MigrationStatus{
		ID:     "test-id-123",
		Status: "running",
	}
	server.mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/migrations/test-id-123", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Success should be true")
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Data should be a map")
	}

	if data["id"] != "test-id-123" {
		t.Errorf("id = %v, want test-id-123", data["id"])
	}
}

func TestServerHandleGetMigrationNotFound(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/api/migrations/nonexistent", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Success {
		t.Error("Success should be false")
	}

	if response.Error == nil {
		t.Error("Error should not be nil")
	}

	if response.Error.Code != "NOT_FOUND" {
		t.Errorf("Error code = %s, want NOT_FOUND", response.Error.Code)
	}
}

func TestServerHandleStopMigration(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	// Create a migration first
	server.mu.Lock()
	server.migrations["stop-test-id"] = &MigrationStatus{
		ID:     "stop-test-id",
		Status: "running",
	}
	server.mu.Unlock()

	req := httptest.NewRequest(http.MethodPost, "/api/migrations/stop-test-id/stop", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Success should be true")
	}

	// Verify status was updated
	server.mu.RLock()
	migration := server.migrations["stop-test-id"]
	server.mu.RUnlock()

	if migration.Status != "stopped" {
		t.Errorf("Migration status = %s, want stopped", migration.Status)
	}
}

func TestServerHandleStopMigrationNotFound(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	req := httptest.NewRequest(http.MethodPost, "/api/migrations/nonexistent/stop", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestServerHandleGetConfig(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Success should be true")
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Data should be a map")
	}

	if _, ok := data["chunkSize"]; !ok {
		t.Error("Response should contain chunkSize")
	}
}

func TestServerHandleUpdateConfig(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	configReq := map[string]interface{}{
		"chunkSize": 200,
		"verbose":   true,
	}

	body, err := json.Marshal(configReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Success should be true")
	}
}

func TestServerHandleUpdateConfigInvalidJSON(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	req := httptest.NewRequest(http.MethodPost, "/api/config", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServerHandleAnalyzeRepo(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	analyzeReq := AnalyzeRequest{
		SourceType: "cvs",
		SourcePath: "/tmp/test-cvs",
	}

	body, err := json.Marshal(analyzeReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/repos/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Accept 200 or 400 (if repo doesn't exist)
	assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest}, rec.Code)

	if rec.Code == http.StatusOK {
		var response APIResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if !response.Success {
			t.Error("Success should be true")
		}
	}
}

func TestServerHandleAnalyzeRepoInvalidJSON(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	req := httptest.NewRequest(http.MethodPost, "/api/repos/analyze", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServerHandleAnalyzeRepoMissingFields(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	analyzeReq := AnalyzeRequest{
		SourceType: "cvs",
		// Missing SourcePath
	}

	body, err := json.Marshal(analyzeReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/repos/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServerConcurrentMigrationAccess(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	// Start multiple migrations concurrently
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			migrationReq := StartMigrationRequest{
				SourceType: "cvs",
				SourcePath: "/tmp/test",
				TargetPath: "/tmp/test",
			}

			body, _ := json.Marshal(migrationReq)
			req := httptest.NewRequest(http.MethodPost, "/api/migrations", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)
			done <- true
		}(i)
	}

	// Wait for all requests
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all migrations were created
	server.mu.RLock()
	count := len(server.migrations)
	server.mu.RUnlock()

	if count != 10 {
		t.Errorf("Expected 10 migrations, got %d", count)
	}
}

func TestServeStatic(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	// Try to get a static file
	req := httptest.NewRequest(http.MethodGet, "/static/style.css", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Just verify the route exists (may return 404 if no static files)
	// The important thing is it doesn't panic
}

func TestServerMigrationStatusFields(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	migrationReq := StartMigrationRequest{
		SourceType: "cvs",
		SourcePath: "/tmp/test-cvs",
		TargetPath: "/tmp/test-git",
	}

	body, _ := json.Marshal(migrationReq)
	req := httptest.NewRequest(http.MethodPost, "/api/migrations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var response APIResponse
	json.Unmarshal(rec.Body.Bytes(), &response)
	data := response.Data.(map[string]interface{})
	migrationID := data["id"].(string)

	// Get the migration and check fields
	req = httptest.NewRequest(http.MethodGet, "/api/migrations/"+migrationID, nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var getResponse APIResponse
	json.Unmarshal(rec.Body.Bytes(), &getResponse)

	migrationData := getResponse.Data.(map[string]interface{})

	// Check expected fields exist
	expectedFields := []string{"id", "status", "percentage", "currentStep", "totalCommits", "processedCommits", "errors"}
	for _, field := range expectedFields {
		if _, ok := migrationData[field]; !ok {
			t.Errorf("Missing field: %s", field)
		}
	}
}

func TestServerErrorResponseType(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/api/migrations/nonexistent", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	var response APIResponse
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response.Error == nil {
		t.Fatal("Error should not be nil")
	}

	if response.Error.Code == "" {
		t.Error("Error code should not be empty")
	}

	if response.Error.Message == "" {
		t.Error("Error message should not be empty")
	}
}

func TestServerServeStaticNonExistent(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	// Request a non-existent static file
	req := httptest.NewRequest(http.MethodGet, "/static/nonexistent.xyz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// The handler should not panic, status depends on embedded files
	// This test mainly verifies no panic occurs
}

func TestServerRoutesNotExist(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/nonexistent"},
		{"POST", "/api/nonexistent"},
		{"GET", "/api/migrations/invalid/invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			// Should return 404 for non-existent routes
			if rec.Code != http.StatusNotFound {
				t.Errorf("Status = %d, want %d for %s %s", rec.Code, http.StatusNotFound, tt.method, tt.path)
			}
		})
	}
}

func TestServerStartMigrationWithEmptyOptions(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	migrationReq := StartMigrationRequest{
		SourceType: "cvs",
		SourcePath: "/tmp/test-cvs",
		TargetPath: "/tmp/test-git",
		Options:    nil, // nil options
	}

	body, err := json.Marshal(migrationReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/migrations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestServerMultipleMigrationStopStart(t *testing.T) {
	server := NewServer(ServerConfig{Port: 8080})
	router := server.Router()

	// Create first migration
	migrationReq := StartMigrationRequest{
		SourceType: "cvs",
		SourcePath: "/tmp/test1",
		TargetPath: "/tmp/test1",
	}

	body, _ := json.Marshal(migrationReq)
	req := httptest.NewRequest(http.MethodPost, "/api/migrations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var response1 APIResponse
	json.Unmarshal(rec.Body.Bytes(), &response1)
	id1 := response1.Data.(map[string]interface{})["id"].(string)

	// Create second migration
	req = httptest.NewRequest(http.MethodPost, "/api/migrations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var response2 APIResponse
	json.Unmarshal(rec.Body.Bytes(), &response2)
	id2 := response2.Data.(map[string]interface{})["id"].(string)

	// Stop first migration
	req = httptest.NewRequest(http.MethodPost, "/api/migrations/"+id1+"/stop", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Verify first is stopped, second is not
	server.mu.RLock()
	m1 := server.migrations[id1]
	m2 := server.migrations[id2]
	server.mu.RUnlock()

	if m1.Status != "stopped" {
		t.Errorf("Migration 1 status = %s, want stopped", m1.Status)
	}
	if m2.Status == "stopped" {
		t.Error("Migration 2 should not be stopped")
	}
}

func TestServerStart(t *testing.T) {
	// Use a high port number to avoid conflicts
	config := ServerConfig{Port: 54321}
	server := NewServer(config)

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Make a test request to verify server is running
	resp, err := http.Get("http://localhost:54321/api/health")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health check status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// The server will continue running, but we've verified it started successfully
	// In a real test, we'd have a way to shut it down gracefully
}
