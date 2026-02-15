// Package mapping provides author mapping for migrations.
package mapping

import (
	"fmt"
	"regexp"
	"strings"
)

// AuthorMap maps CVS usernames to Git author info
type AuthorMap struct {
	mapping      map[string]string
	defaultEmail string
}

// NewAuthorMap creates a new author map
func NewAuthorMap(config map[string]string) *AuthorMap {
	return &AuthorMap{
		mapping:      config,
		defaultEmail: "users.noreply.cvs.example.org",
	}
}

// NewAuthorMapWithDefault creates an author map with custom default domain
func NewAuthorMapWithDefault(config map[string]string, defaultDomain string) *AuthorMap {
	return &AuthorMap{
		mapping:      config,
		defaultEmail: defaultDomain,
	}
}

// Get returns the Git author name and email for a CVS username
func (am *AuthorMap) Get(username string) (string, string) {
	if format, ok := am.mapping[username]; ok {
		name, email, err := ParseAuthor(format)
		if err == nil {
			return name, email
		}
	}

	// Default format: username <username@default>
	return username, fmt.Sprintf("%s@%s", username, am.defaultEmail)
}

// ParseAuthor parses a "Name <email>" string
func ParseAuthor(format string) (string, string, error) {
	// Pattern: "Name <email>"
	re := regexp.MustCompile(`^(.+?)\s*<(.+?)>$`)
	matches := re.FindStringSubmatch(format)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid author format: %s", format)
	}

	name := strings.TrimSpace(matches[1])
	email := strings.TrimSpace(matches[2])

	if name == "" || email == "" {
		return "", "", fmt.Errorf("invalid author format: %s", format)
	}

	return name, email, nil
}

// AuthorExtractor extracts unique authors from a repository
type AuthorExtractor struct {
	authors map[string]bool
}

// NewAuthorExtractor creates a new author extractor
func NewAuthorExtractor() *AuthorExtractor {
	return &AuthorExtractor{
		authors: make(map[string]bool),
	}
}

// Add adds an author to the extractor
func (ae *AuthorExtractor) Add(username string) {
	ae.authors[username] = true
}

// List returns all unique authors
func (ae *AuthorExtractor) List() []string {
	var result []string
	for author := range ae.authors {
		result = append(result, author)
	}
	return result
}

// GenerateTemplate generates a mapping template for all authors
func (ae *AuthorExtractor) GenerateTemplate() map[string]string {
	template := make(map[string]string)
	for author := range ae.authors {
		template[author] = fmt.Sprintf("%s <%s@example.com>", author, author)
	}
	return template
}
