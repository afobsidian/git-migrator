package requirements

import (
	"testing"

	"github.com/adamf123git/git-migrator/internal/mapping"
)

// TestAuthorMapLoad tests loading author mappings
func TestAuthorMapLoad(t *testing.T) {
	config := map[string]string{
		"cvsuser1": "John Doe <john@example.com>",
		"cvsuser2": "Jane Smith <jane@example.com>",
	}

	authorMap := mapping.NewAuthorMap(config)

	// Test mapping exists
	name, email := authorMap.Get("cvsuser1")
	if name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %q", name)
	}
	if email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got %q", email)
	}
}

// TestAuthorMapUnmapped tests handling unmapped users
func TestAuthorMapUnmapped(t *testing.T) {
	authorMap := mapping.NewAuthorMap(nil)

	// Unmapped user should use default format
	name, email := authorMap.Get("unknownuser")
	if name != "unknownuser" {
		t.Errorf("Expected name 'unknownuser', got %q", name)
	}
	if email == "" {
		t.Error("Expected email to be set for unmapped user")
	}
}

// TestAuthorMapDefaultDomain tests default domain for unmapped users
func TestAuthorMapDefaultDomain(t *testing.T) {
	authorMap := mapping.NewAuthorMapWithDefault(nil, "example.org")

	name, email := authorMap.Get("testuser")
	if name != "testuser" {
		t.Errorf("Expected name 'testuser', got %q", name)
	}
	expectedEmail := "testuser@example.org"
	if email != expectedEmail {
		t.Errorf("Expected email %q, got %q", expectedEmail, email)
	}
}

// TestAuthorMapParse tests parsing author string
func TestAuthorMapParse(t *testing.T) {
	tests := []struct {
		input    string
		name     string
		email    string
		hasError bool
	}{
		{"John Doe <john@example.com>", "John Doe", "john@example.com", false},
		{"Jane Smith <jane@test.org>", "Jane Smith", "jane@test.org", false},
		{"No Email", "", "", true},
		{"", "", "", true},
	}

	for _, tt := range tests {
		name, email, err := mapping.ParseAuthor(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("Expected error for input %q", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", tt.input, err)
			}
			if name != tt.name {
				t.Errorf("Expected name %q, got %q", tt.name, name)
			}
			if email != tt.email {
				t.Errorf("Expected email %q, got %q", tt.email, email)
			}
		}
	}
}

// TestAuthorExtract tests extracting authors from CVS
func TestAuthorExtract(t *testing.T) {
	extractor := mapping.NewAuthorExtractor()

	// Add some authors
	extractor.Add("user1")
	extractor.Add("user2")
	extractor.Add("user1") // Duplicate

	authors := extractor.List()
	if len(authors) != 2 {
		t.Errorf("Expected 2 unique authors, got %d", len(authors))
	}

	// Check authors exist
	found := make(map[string]bool)
	for _, a := range authors {
		found[a] = true
	}
	if !found["user1"] || !found["user2"] {
		t.Error("Expected user1 and user2 in author list")
	}
}

// TestAuthorGenerateTemplate tests generating mapping template
func TestAuthorGenerateTemplate(t *testing.T) {
	extractor := mapping.NewAuthorExtractor()
	extractor.Add("jdoe")
	extractor.Add("asmith")

	template := extractor.GenerateTemplate()

	// Template should contain both authors
	if len(template) == 0 {
		t.Error("Expected non-empty template")
	}

	// Check format
	for user, format := range template {
		if format == "" {
			t.Errorf("Expected format for user %s", user)
		}
	}
}
