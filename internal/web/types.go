package web

import (
	"time"
)

// APIResponse is the standard response format for all API endpoints
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an error in an API response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// StartMigrationRequest is the request body for starting a migration
type StartMigrationRequest struct {
	SourceType string                 `json:"sourceType"`
	SourcePath string                 `json:"sourcePath"`
	TargetPath string                 `json:"targetPath"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

// AnalyzeRequest is the request body for repository analysis
type AnalyzeRequest struct {
	SourceType string `json:"sourceType"`
	SourcePath string `json:"sourcePath"`
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	ID               string    `json:"id"`
	Status           string    `json:"status"`
	Percentage       int       `json:"percentage"`
	CurrentStep      string    `json:"currentStep"`
	TotalCommits     int       `json:"totalCommits"`
	ProcessedCommits int       `json:"processedCommits"`
	Errors           []string  `json:"errors"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// ProgressEvent is a WebSocket event for progress updates
type ProgressEvent struct {
	Type string       `json:"type"`
	Data ProgressData `json:"data"`
}

// ProgressData contains the progress details
type ProgressData struct {
	MigrationID      string   `json:"migrationId"`
	Status           string   `json:"status"`
	Percentage       int      `json:"percentage"`
	CurrentStep      string   `json:"currentStep"`
	TotalCommits     int      `json:"totalCommits"`
	ProcessedCommits int      `json:"processedCommits"`
	Errors           []string `json:"errors"`
}

// ServerConfig is the configuration for the web server
type ServerConfig struct {
	Port         int
	ConfigPath   string
	DatabasePath string
}

// HealthStatus represents the health check response
type HealthStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// ConfigData represents the configuration response
type ConfigData struct {
	ChunkSize int  `json:"chunkSize"`
	Verbose   bool `json:"verbose"`
	DryRun    bool `json:"dryRun"`
}

// ErrorResponse creates an error API response
func ErrorResponse(code, message string) APIResponse {
	return APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	}
}

// SuccessResponse creates a success API response
func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
	}
}
