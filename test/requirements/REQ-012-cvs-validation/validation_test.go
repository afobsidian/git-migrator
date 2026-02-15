package requirements

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adamf123git/git-migrator/internal/vcs/cvs"
)

// TestCVSValidatorValidRepo tests validating a valid CVS repository
func TestCVSValidatorValidRepo(t *testing.T) {
	// Create a temporary CVS repository structure
	tmpDir, err := os.MkdirTemp("", "cvs-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create CVSROOT
	cvsroot := filepath.Join(tmpDir, "CVSROOT")
	if err := os.MkdirAll(cvsroot, 0755); err != nil {
		t.Fatal(err)
	}

	// Create required CVSROOT files
	for _, file := range []string{"history", "val-tags"} {
		if err := os.WriteFile(filepath.Join(cvsroot, file), []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Validate
	validator := cvs.NewValidator()
	result := validator.Validate(tmpDir)

	if !result.Valid {
		t.Errorf("Expected valid repository, got errors: %v", result.Errors)
	}
}

// TestCVSValidatorMissingCVSRoot tests validating repo without CVSROOT
func TestCVSValidatorMissingCVSRoot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cvs-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	validator := cvs.NewValidator()
	result := validator.Validate(tmpDir)

	if result.Valid {
		t.Error("Expected invalid repository (missing CVSROOT)")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected validation errors")
	}

	foundCVSRootError := false
	for _, e := range result.Errors {
		if e.Field == "CVSROOT" {
			foundCVSRootError = true
			break
		}
	}

	if !foundCVSRootError {
		t.Error("Expected error about missing CVSROOT")
	}
}

// TestCVSValidatorNotDirectory tests validating non-directory path
func TestCVSValidatorNotDirectory(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cvs-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	validator := cvs.NewValidator()
	result := validator.Validate(tmpFile.Name())

	if result.Valid {
		t.Error("Expected invalid repository (not a directory)")
	}
}

// TestCVSValidatorNonExistent tests validating non-existent path
func TestCVSValidatorNonExistent(t *testing.T) {
	validator := cvs.NewValidator()
	result := validator.Validate("/nonexistent/path")

	if result.Valid {
		t.Error("Expected invalid repository (non-existent)")
	}
}

// TestCVSValidatorModule tests validating a module within the repository
func TestCVSValidatorModule(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cvs-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create CVSROOT
	cvsroot := filepath.Join(tmpDir, "CVSROOT")
	if err := os.MkdirAll(cvsroot, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a module
	moduleDir := filepath.Join(tmpDir, "mymodule")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create CVS subdir in module
	cvsSubdir := filepath.Join(moduleDir, "CVS")
	if err := os.MkdirAll(cvsSubdir, 0755); err != nil {
		t.Fatal(err)
	}

	validator := cvs.NewValidator()
	result := validator.Validate(tmpDir)

	if !result.Valid {
		t.Errorf("Expected valid repository, got errors: %v", result.Errors)
	}
}

// TestCVSValidationResult tests ValidationResult structure
func TestCVSValidationResult(t *testing.T) {
	result := &cvs.ValidationResult{
		Valid: true,
		Infos: []cvs.ValidationMessage{
			{Field: "repository", Message: "Repository structure is valid"},
		},
	}

	if !result.Valid {
		t.Error("Expected Valid to be true")
	}

	if len(result.Infos) != 1 {
		t.Errorf("Expected 1 info message, got %d", len(result.Infos))
	}
}

// TestCVSValidatorWarnings tests that warnings don't make repo invalid
func TestCVSValidatorWarnings(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cvs-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create CVSROOT (minimal)
	cvsroot := filepath.Join(tmpDir, "CVSROOT")
	if err := os.MkdirAll(cvsroot, 0755); err != nil {
		t.Fatal(err)
	}

	validator := cvs.NewValidator()
	result := validator.Validate(tmpDir)

	// Warnings don't make repo invalid
	if len(result.Warnings) > 0 && !result.Valid {
		t.Error("Warnings should not make repository invalid")
	}
}
