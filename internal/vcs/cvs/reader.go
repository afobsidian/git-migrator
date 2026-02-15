package cvs

import (
	"os"
	"path/filepath"
)

// ValidationMessage represents a validation message
type ValidationMessage struct {
	Field   string
	Message string
}

// ValidationResult represents the result of CVS repository validation
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationMessage
	Warnings []ValidationMessage
	Infos    []ValidationMessage
}

// Validator validates CVS repositories
type Validator struct{}

// NewValidator creates a new CVS repository validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a CVS repository at the given path
func (v *Validator) Validate(path string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Check path exists
	info, err := os.Stat(path)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationMessage{
			Field:   "path",
			Message: "Path does not exist: " + path,
		})
		return result
	}

	// Check is directory
	if !info.IsDir() {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationMessage{
			Field:   "path",
			Message: "Path is not a directory: " + path,
		})
		return result
	}

	// Check for CVSROOT
	cvsroot := filepath.Join(path, "CVSROOT")
	if _, err := os.Stat(cvsroot); os.IsNotExist(err) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationMessage{
			Field:   "CVSROOT",
			Message: "CVSROOT directory not found",
		})
		return result
	}

	result.Infos = append(result.Infos, ValidationMessage{
		Field:   "repository",
		Message: "Repository structure is valid",
	})

	// Check for common CVSROOT files (warnings if missing)
	requiredFiles := []string{"history", "val-tags"}
	for _, file := range requiredFiles {
		filePath := filepath.Join(cvsroot, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			result.Warnings = append(result.Warnings, ValidationMessage{
				Field:   "CVSROOT/" + file,
				Message: "Optional file not found",
			})
		}
	}

	return result
}
