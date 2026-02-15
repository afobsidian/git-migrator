package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// Server is the web server
type Server struct {
	config     ServerConfig
	router     *chi.Mux
	migrations map[string]*MigrationStatus
	mu         sync.RWMutex
}

// NewServer creates a new web server
func NewServer(config ServerConfig) *Server {
	s := &Server{
		config:     config,
		migrations: make(map[string]*MigrationStatus),
	}

	s.setupRouter()
	return s
}

// Router returns the HTTP router
func (s *Server) Router() *chi.Mux {
	return s.router
}

// setupRouter configures the HTTP router
func (s *Server) setupRouter() {
	s.router = chi.NewRouter()

	// Middleware
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.RequestID)

	// Static files
	s.router.Get("/static/*", s.serveStatic)

	// Web UI routes
	s.router.Get("/", s.serveIndex)
	s.router.Get("/new", s.serveNewMigration)
	s.router.Get("/config", s.serveConfig)
	s.router.Get("/migration/{id}", s.serveMigration)

	// API routes
	s.router.Get("/api/health", s.handleHealth)
	s.router.Get("/api/migrations", s.handleListMigrations)
	s.router.Post("/api/migrations", s.handleStartMigration)
	s.router.Get("/api/migrations/{id}", s.handleGetMigration)
	s.router.Post("/api/migrations/{id}/stop", s.handleStopMigration)
	s.router.Get("/api/config", s.handleGetConfig)
	s.router.Post("/api/config", s.handleUpdateConfig)
	s.router.Post("/api/repos/analyze", s.handleAnalyzeRepo)

	// WebSocket
	s.router.Get("/ws/progress/{id}", s.handleWebSocket)
}

// serveStatic serves static files
func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	// Use embedded static files
	fs := http.FileServer(getStaticFS())
	http.StripPrefix("/static/", fs).ServeHTTP(w, r)
}

// serveIndex serves the main page
func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(indexHTML)); err != nil {
		log.Printf("Warning: failed to write index HTML response: %v", err)
	}
}

// serveNewMigration serves the new migration page
func (s *Server) serveNewMigration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(newMigrationHTML)); err != nil {
		log.Printf("Warning: failed to write new migration HTML response: %v", err)
	}
}

// serveConfig serves the config page
func (s *Server) serveConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(configHTML)); err != nil {
		log.Printf("Warning: failed to write config HTML response: %v", err)
	}
}

// serveMigration serves the migration detail page
func (s *Server) serveMigration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(migrationHTML)); err != nil {
		log.Printf("Warning: failed to write migration HTML response: %v", err)
	}
}

// handleHealth handles GET /api/health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(SuccessResponse(HealthStatus{
		Status:  "ok",
		Version: "0.1.0",
	})); err != nil {
		log.Printf("Warning: failed to encode health response: %v", err)
	}
}

// handleListMigrations handles GET /api/migrations
func (s *Server) handleListMigrations(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	migrations := make([]interface{}, 0, len(s.migrations))
	for _, m := range s.migrations {
		migrations = append(migrations, m)
	}
	s.mu.RUnlock()

	if err := json.NewEncoder(w).Encode(SuccessResponse(migrations)); err != nil {
		log.Printf("Warning: failed to encode migrations list response: %v", err)
	}
}

// handleStartMigration handles POST /api/migrations
func (s *Server) handleStartMigration(w http.ResponseWriter, r *http.Request) {
	var req StartMigrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(ErrorResponse("INVALID_JSON", "Invalid JSON body")); encodeErr != nil {
			log.Printf("Warning: failed to encode error response: %v", encodeErr)
		}
		return
	}

	// Validate required fields
	if req.SourcePath == "" || req.TargetPath == "" || req.SourceType == "" {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(ErrorResponse("VALIDATION_ERROR", "Missing required fields")); err != nil {
			log.Printf("Warning: failed to encode validation error response: %v", err)
		}
		return
	}

	// Create migration
	id := uuid.New().String()
	now := time.Now()
	migration := &MigrationStatus{
		ID:               id,
		Status:           "pending",
		Percentage:       0,
		CurrentStep:      "Initializing",
		TotalCommits:     0,
		ProcessedCommits: 0,
		Errors:           []string{},
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	s.mu.Lock()
	s.migrations[id] = migration
	s.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(SuccessResponse(map[string]interface{}{
		"id":      id,
		"status":  migration.Status,
		"message": "Migration started",
	})); err != nil {
		log.Printf("Warning: failed to encode start migration response: %v", err)
	}
}

// handleGetMigration handles GET /api/migrations/:id
func (s *Server) handleGetMigration(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.mu.RLock()
	migration, exists := s.migrations[id]
	s.mu.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(ErrorResponse("NOT_FOUND", "Migration not found")); err != nil {
			log.Printf("Warning: failed to encode not found error response: %v", err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(SuccessResponse(migration)); err != nil {
		log.Printf("Warning: failed to encode migration response: %v", err)
	}
}

// handleStopMigration handles POST /api/migrations/:id/stop
func (s *Server) handleStopMigration(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.mu.Lock()
	migration, exists := s.migrations[id]
	if exists {
		migration.Status = "stopped"
		migration.UpdatedAt = time.Now()
	}
	s.mu.Unlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(ErrorResponse("NOT_FOUND", "Migration not found")); err != nil {
			log.Printf("Warning: failed to encode not found error response: %v", err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(SuccessResponse(map[string]string{
		"id":      id,
		"status":  "stopped",
		"message": "Migration stopped",
	})); err != nil {
		log.Printf("Warning: failed to encode stop migration response: %v", err)
	}
}

// handleGetConfig handles GET /api/config
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(SuccessResponse(ConfigData{
		ChunkSize: 100,
		Verbose:   false,
		DryRun:    false,
	})); err != nil {
		log.Printf("Warning: failed to encode config response: %v", err)
	}
}

// handleUpdateConfig handles POST /api/config
func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(ErrorResponse("INVALID_JSON", "Invalid JSON body")); encodeErr != nil {
			log.Printf("Warning: failed to encode config error response: %v", encodeErr)
		}
		return
	}

	// In a real implementation, this would update the config file
	// For now, just return success
	if err := json.NewEncoder(w).Encode(SuccessResponse(map[string]string{
		"message": "Configuration updated",
	})); err != nil {
		log.Printf("Warning: failed to encode config update response: %v", err)
	}
}

// handleAnalyzeRepo handles POST /api/repos/analyze
func (s *Server) handleAnalyzeRepo(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(ErrorResponse("INVALID_JSON", "Invalid JSON body")); encodeErr != nil {
			log.Printf("Warning: failed to encode analyze error response: %v", encodeErr)
		}
		return
	}

	// Validate
	if req.SourceType == "" || req.SourcePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(ErrorResponse("VALIDATION_ERROR", "Missing required fields")); encodeErr != nil {
			log.Printf("Warning: failed to encode validation error response: %v", encodeErr)
		}
		return
	}

	// In a real implementation, this would analyze the repository
	// For now, return a mock response
	if err := json.NewEncoder(w).Encode(SuccessResponse(map[string]interface{}{
		"type":        req.SourceType,
		"path":        req.SourcePath,
		"commitCount": 0,
		"branchCount": 0,
		"tagCount":    0,
		"authors":     []string{},
		"valid":       true,
	})); err != nil {
		log.Printf("Warning: failed to encode analyze response: %v", err)
	}
}

// Start starts the web server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	fmt.Printf("Starting web server on %s\n", addr)
	return http.ListenAndServe(addr, s.router)
}
