package commands

import (
	"fmt"
	"os"

	"github.com/adamf123git/git-migrator/internal/mapping"
	"github.com/adamf123git/git-migrator/internal/vcs/cvs"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var authorsCmd = &cobra.Command{
	Use:   "authors",
	Short: "Manage author mappings",
	Long: `Commands for extracting and managing author information from
version control repositories.

Use the extract subcommand to get a list of all unique authors from a
repository, which can then be used to create author mappings for migration.`,
}

var authorsExtractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract unique authors from a repository",
	Long: `Extract all unique author usernames from a version control repository.
This is useful for creating author mappings before migration.

The output can be in plain text (one author per line) or YAML format
(ready to be included in a migration config file).`,
	RunE: runAuthorsExtract,
}

var (
	authorsSource string
	authorsFormat string
)

func init() {
	rootCmd.AddCommand(authorsCmd)
	authorsCmd.AddCommand(authorsExtractCmd)

	authorsExtractCmd.Flags().StringVarP(&authorsSource, "source", "s", "", "Path to source repository")
	authorsExtractCmd.Flags().StringVarP(&authorsFormat, "format", "f", "text", "Output format (text or yaml)")
	var err = authorsExtractCmd.MarkFlagRequired("source")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marking flag as required: %v\n", err)
		os.Exit(1)
	}
}

func runAuthorsExtract(cmd *cobra.Command, args []string) error {
	// Validate format
	if authorsFormat != "text" && authorsFormat != "yaml" {
		return fmt.Errorf("unsupported format: %s (supported: text, yaml)", authorsFormat)
	}

	// Create reader (currently only CVS is supported)
	reader := cvs.NewReader(authorsSource)

	// Validate repository
	if err := reader.Validate(); err != nil {
		return fmt.Errorf("repository validation failed: %w", err)
	}

	// Get commits and extract authors
	commitIter, err := reader.GetCommits()
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	authorExtractor := mapping.NewAuthorExtractor()

	for commitIter.Next() {
		commit := commitIter.Commit()
		authorExtractor.Add(commit.Author)
	}

	if err := commitIter.Err(); err != nil {
		return fmt.Errorf("error iterating commits: %w", err)
	}

	// Close reader
	if err := reader.Close(); err != nil {
		return fmt.Errorf("failed to close reader: %w", err)
	}

	// Output based on format
	switch authorsFormat {
	case "text":
		// Simple text format - one author per line
		for _, author := range authorExtractor.List() {
			fmt.Println(author)
		}
	case "yaml":
		// YAML format - ready to be included in config
		template := authorExtractor.GenerateTemplate()
		output, err := yaml.Marshal(template)
		if err != nil {
			return fmt.Errorf("failed to generate YAML: %w", err)
		}
		fmt.Print(string(output))
	}

	return nil
}
