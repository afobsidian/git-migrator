package requirements_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adamf123git/git-migrator/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIHealthEndpoint tests the health check endpoint
func TestAPIHealthEndpoint(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response web.APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
}

// TestAPIListMigrations tests listing migrations
func TestAPIListMigrations(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/api/migrations", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response web.APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Data should be an array of migrations
	_, ok := response.Data.([]interface{})
	assert.True(t, ok, "Data should be an array")
}

// TestAPIStartMigration tests starting a new migration
func TestAPIStartMigration(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})
	router := server.Router()

	migrationReq := web.StartMigrationRequest{
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

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response web.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Data should contain migration ID
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok, "Data should be an object")
	assert.Contains(t, data, "id")
}

// TestAPIGetMigrationStatus tests getting migration status
func TestAPIGetMigrationStatus(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})
	router := server.Router()

	// First create a migration
	migrationReq := web.StartMigrationRequest{
		SourceType: "cvs",
		SourcePath: "/tmp/test-cvs",
		TargetPath: "/tmp/test-git",
		Options: map[string]interface{}{
			"dryRun": true,
		},
	}

	body, _ := json.Marshal(migrationReq)
	req := httptest.NewRequest(http.MethodPost, "/api/migrations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var createResp web.APIResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp), "Failed to unmarshal create response")
	createData := createResp.Data.(map[string]interface{})
	migrationID := createData["id"].(string)

	// Now get the status
	req = httptest.NewRequest(http.MethodGet, "/api/migrations/"+migrationID, nil)
	rec = httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response web.APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok, "Data should be an object")
	assert.Contains(t, data, "id")
	assert.Contains(t, data, "status")
}

// TestAPIStopMigration tests stopping a migration
func TestAPIStopMigration(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})
	router := server.Router()

	// First create a migration
	migrationReq := web.StartMigrationRequest{
		SourceType: "cvs",
		SourcePath: "/tmp/test-cvs",
		TargetPath: "/tmp/test-git",
		Options: map[string]interface{}{
			"dryRun": true,
		},
	}

	body, _ := json.Marshal(migrationReq)
	req := httptest.NewRequest(http.MethodPost, "/api/migrations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var createResp web.APIResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp), "Failed to unmarshal create response")
	createData := createResp.Data.(map[string]interface{})
	migrationID := createData["id"].(string)

	// Now stop it
	req = httptest.NewRequest(http.MethodPost, "/api/migrations/"+migrationID+"/stop", nil)
	rec = httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response web.APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
}

// TestAPIGetConfig tests getting configuration
func TestAPIGetConfig(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})
	router := server.Router()

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response web.APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
}

// TestAPIUpdateConfig tests updating configuration
func TestAPIUpdateConfig(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})
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

	assert.Equal(t, http.StatusOK, rec.Code)

	var response web.APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
}

// TestAPIAnalyzeRepository tests repository analysis
func TestAPIAnalyzeRepository(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})
	router := server.Router()

	analyzeReq := web.AnalyzeRequest{
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
}

// TestAPIErrorResponse tests error handling
func TestAPIErrorResponse(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})
	router := server.Router()

	// Request non-existent migration
	req := httptest.NewRequest(http.MethodGet, "/api/migrations/nonexistent", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response web.APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
}

// TestAPIInvalidJSON tests handling invalid JSON
func TestAPIInvalidJSON(t *testing.T) {
	server := web.NewServer(web.ServerConfig{
		Port: 8080,
	})
	router := server.Router()

	req := httptest.NewRequest(http.MethodPost, "/api/migrations", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response web.APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
}
