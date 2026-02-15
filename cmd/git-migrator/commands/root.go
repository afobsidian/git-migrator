package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information (set at build time)
	Version   = "dev"
	GitCommit = "none"
	BuildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "git-migrator",
	Short: "Migrate version control repositories with full history preservation",
	Long: `Git-Migrator is an open-source tool for migrating repositories from legacy
version control systems (CVS, SVN) to Git while preserving complete history,
including commits, branches, tags, and author information.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Version = Version
}

// handleError provides centralized error handling for commands.
// This function is reserved for future use when implementing additional
// subcommands that require consistent error handling and exit behavior.
// It will be used to standardize error reporting across all command handlers.
//
//nolint:unused // Reserved for future command implementations
func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
