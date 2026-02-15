package commands

import (
	"fmt"
	"os"

	"github.com/adamf123git/git-migrator/internal/mapping"
	"github.com/adamf123git/git-migrator/internal/vcs/cvs"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a CVS or SVN repository",
	Long: `Analyze a version control repository to understand its structure,
including the number of commits, branches, tags, and unique authors.

This command is useful for understanding what will be migrated before
running the actual migration.`,
	RunE: runAnalyze,
}

var (
	analyzeSourceType string
	analyzeSource     string
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(&analyzeSourceType, "source-type", "t", "cvs", "Source VCS type (cvs or svn)")
	analyzeCmd.Flags().StringVarP(&analyzeSource, "source", "s", "", "Path to source repository")
	var err = analyzeCmd.MarkFlagRequired("source")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marking flag as required: %v\n", err)
		os.Exit(1)
	}
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	// Validate source type
	if analyzeSourceType != "cvs" && analyzeSourceType != "svn" {
		return fmt.Errorf("unsupported source type: %s (supported: cvs, svn)", analyzeSourceType)
	}

	// Currently only CVS is implemented
	if analyzeSourceType == "svn" {
		return fmt.Errorf("SVN support is not yet implemented")
	}

	// Create reader
	reader := cvs.NewReader(analyzeSource)

	// Validate repository
	fmt.Printf("Analyzing %s repository at: %s\n\n", analyzeSourceType, analyzeSource)
	if err := reader.Validate(); err != nil {
		return fmt.Errorf("repository validation failed: %w", err)
	}

	// Get branches
	branches, err := reader.GetBranches()
	if err != nil {
		return fmt.Errorf("failed to get branches: %w", err)
	}

	// Get tags
	tags, err := reader.GetTags()
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	// Get commits and extract authors
	commitIter, err := reader.GetCommits()
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	authorExtractor := mapping.NewAuthorExtractor()
	commitCount := 0

	for commitIter.Next() {
		commit := commitIter.Commit()
		commitCount++
		authorExtractor.Add(commit.Author)
	}

	if err := commitIter.Err(); err != nil {
		return fmt.Errorf("error iterating commits: %w", err)
	}

	// Close reader
	if err := reader.Close(); err != nil {
		return fmt.Errorf("failed to close reader: %w", err)
	}

	// Display results
	fmt.Println("Repository Analysis Results")
	fmt.Println("==========================")
	fmt.Printf("Type:           %s\n", analyzeSourceType)
	fmt.Printf("Path:           %s\n", analyzeSource)
	fmt.Printf("Commits:        %d\n", commitCount)
	fmt.Printf("Branches:       %d\n", len(branches))
	fmt.Printf("Tags:           %d\n", len(tags))
	fmt.Printf("Unique Authors: %d\n\n", len(authorExtractor.List()))

	if len(branches) > 0 {
		fmt.Println("Branches:")
		for _, branch := range branches {
			fmt.Printf("  - %s\n", branch)
		}
		fmt.Println()
	}

	if len(tags) > 0 {
		fmt.Println("Tags:")
		for name, rev := range tags {
			fmt.Printf("  - %s (revision: %s)\n", name, rev)
		}
		fmt.Println()
	}

	authors := authorExtractor.List()
	if len(authors) > 0 {
		fmt.Println("Authors:")
		for _, author := range authors {
			fmt.Printf("  - %s\n", author)
		}
		fmt.Println()
	}

	fmt.Println("Repository is valid and ready for migration.")

	return nil
}
