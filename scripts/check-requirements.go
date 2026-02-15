package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Requirements holds requirement metadata
type Requirement struct {
	ID     string
	Name   string
	Status string
}

func main() {
	fmt.Println("🔍 Checking Requirements Coverage...")
	fmt.Println()

	requirements := []Requirement{
		{ID: "REQ-007", Name: "CLI Interface"},
		{ID: "REQ-009", Name: "TDD with Regression Testing"},
		{ID: "REQ-010", Name: "Requirements Validation"},
	}

	allValid := true
	complete := 0
	inProgress := 0
	notStarted := 0

	for _, req := range requirements {
		status := checkRequirement(req.ID)
		req.Status = status

		switch status {
		case "✅ Complete":
			complete++
		case "🟡 In Progress":
			inProgress++
		case "⚪ Not Started":
			notStarted++
		}

		fmt.Printf("  %s: %s - %s\n", req.ID, status, req.Name)

		if status == "⚪ Not Started" {
			allValid = false
		}
	}

	fmt.Println()
	fmt.Printf("Summary: %d complete, %d in progress, %d not started\n",
		complete, inProgress, notStarted)

	if !allValid {
		fmt.Println()
		fmt.Println("❌ Some requirements have no tests!")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ All requirements have test coverage")
}

// findRequirementDir finds the directory for a requirement ID (e.g., REQ-007 matches REQ-007-cli-interface)
func findRequirementDir(reqID string) string {
	baseDir := "test/requirements"
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), reqID) {
			return filepath.Join(baseDir, entry.Name())
		}
	}
	return ""
}

func checkRequirement(reqID string) string {
	// Find requirement directory (handles REQ-007 -> REQ-007-cli-interface)
	reqDir := findRequirementDir(reqID)
	if reqDir == "" {
		return "⚪ Not Started"
	}

	// Check if requirement.md exists
	reqFile := filepath.Join(reqDir, "requirement.md")
	if _, err := os.Stat(reqFile); os.IsNotExist(err) {
		return "⚪ Not Started"
	}

	// Find all test files (both *_test.go and test.go patterns)
	testFiles := []string{}
	if err := filepath.WalkDir(reqDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			filename := filepath.Base(path)
			if strings.HasSuffix(path, "_test.go") || filename == "test.go" {
				testFiles = append(testFiles, path)
			}
		}
		return nil
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to walk directory %s: %v\n", reqDir, err)
	}

	if len(testFiles) == 0 {
		return "⚪ Not Started"
	}

	// Count test functions
	testCount := 0
	for _, testFile := range testFiles {
		content, err := os.ReadFile(testFile)
		if err != nil {
			continue
		}
		testCount += strings.Count(string(content), "func Test")
	}

	if testCount == 0 {
		return "⚪ Not Started"
	}

	// If tests exist, consider it in progress
	return "🟡 In Progress"
}
