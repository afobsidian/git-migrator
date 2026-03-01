package commands

import (
	"fmt"
	"os"

	"github.com/adamf123git/git-migrator/internal/core"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync changes between Git and CVS repositories",
	Long: `Synchronise changes between a Git repository and a CVS repository.

Three sync directions are supported:
  git-to-cvs    Apply new Git commits to the CVS repository
  cvs-to-git    Apply new CVS commits to the Git repository
  bidirectional Sync in both directions (default)

Use --dry-run to preview planned changes without applying them.

Example usage:
  git-migrator sync --config sync-config.yaml
  git-migrator sync --config sync-config.yaml --direction git-to-cvs
  git-migrator sync --config sync-config.yaml --dry-run`,
	RunE: runSync,
}

var (
	syncConfigFile string
	syncDryRun     bool
	syncVerbose    bool
	syncDirection  string
)

// SyncConfigFile is the YAML schema for a sync configuration file.
type SyncConfigFile struct {
	Git struct {
		Path string `yaml:"path"`
	} `yaml:"git"`

	CVS struct {
		Path    string `yaml:"path"`
		Module  string `yaml:"module"`
		WorkDir string `yaml:"workDir"`
	} `yaml:"cvs"`

	Sync struct {
		Direction string `yaml:"direction"`
		StateFile string `yaml:"stateFile"`
	} `yaml:"sync"`

	Mapping struct {
		Authors map[string]string `yaml:"authors"`
	} `yaml:"mapping"`

	Options struct {
		DryRun  bool `yaml:"dryRun"`
		Verbose bool `yaml:"verbose"`
	} `yaml:"options"`
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringVarP(&syncConfigFile, "config", "c", "", "Path to sync configuration file (required)")
	syncCmd.Flags().BoolVarP(&syncDryRun, "dry-run", "d", false, "Preview sync without making changes")
	syncCmd.Flags().BoolVarP(&syncVerbose, "verbose", "v", false, "Show detailed output")
	syncCmd.Flags().StringVar(&syncDirection, "direction", "", "Sync direction: git-to-cvs, cvs-to-git, bidirectional")

	if err := syncCmd.MarkFlagRequired("config"); err != nil {
		fmt.Fprintf(os.Stderr, "Error marking flag as required: %v\n", err)
		os.Exit(1)
	}
}

func runSync(cmd *cobra.Command, args []string) error {
	config, err := loadSyncConfigFile(syncConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load sync configuration: %w", err)
	}

	// CLI flags override the config file values
	if syncDryRun {
		config.Options.DryRun = true
	}
	if syncVerbose {
		config.Options.Verbose = true
	}
	if syncDirection != "" {
		config.Sync.Direction = syncDirection
	}

	syncConfig := &core.SyncConfig{
		GitPath:    config.Git.Path,
		CVSPath:    config.CVS.Path,
		CVSModule:  config.CVS.Module,
		CVSWorkDir: config.CVS.WorkDir,
		Direction:  core.SyncDirection(config.Sync.Direction),
		AuthorMap:  config.Mapping.Authors,
		StateFile:  config.Sync.StateFile,
		DryRun:     config.Options.DryRun,
	}

	if config.Options.Verbose || config.Options.DryRun {
		printSyncInfo(config, syncConfig)
	}

	if config.Options.DryRun {
		fmt.Println("\n🔍 DRY RUN MODE - No changes will be made")
	}

	syncer := core.NewSyncer(syncConfig)

	fmt.Printf("\nStarting %s sync...\n", syncConfig.Direction)
	if err := syncer.Run(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	if config.Options.DryRun {
		fmt.Println("\n✓ Dry run completed successfully")
		fmt.Println("Run without --dry-run to apply changes")
	} else {
		fmt.Println("\n✓ Sync completed successfully!")
	}

	return nil
}

func loadSyncConfigFile(path string) (*SyncConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config SyncConfigFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.Git.Path == "" {
		return nil, fmt.Errorf("git.path is required")
	}
	if config.CVS.Path == "" {
		return nil, fmt.Errorf("cvs.path is required")
	}
	if config.CVS.Module == "" {
		return nil, fmt.Errorf("cvs.module is required")
	}

	// Defaults
	if config.Sync.Direction == "" {
		config.Sync.Direction = string(core.SyncBidirectional)
	}
	if config.Mapping.Authors == nil {
		config.Mapping.Authors = make(map[string]string)
	}

	return &config, nil
}

func printSyncInfo(config *SyncConfigFile, syncConfig *core.SyncConfig) {
	fmt.Println("\nSync Configuration")
	fmt.Println("==================")
	fmt.Printf("Git Repository:  %s\n", config.Git.Path)
	fmt.Printf("CVS Repository:  %s\n", config.CVS.Path)
	fmt.Printf("CVS Module:      %s\n", config.CVS.Module)
	fmt.Printf("Direction:       %s\n", syncConfig.Direction)
	fmt.Printf("Dry Run:         %v\n", config.Options.DryRun)

	if len(config.Mapping.Authors) > 0 {
		fmt.Printf("\nAuthor Mappings: %d\n", len(config.Mapping.Authors))
		if config.Options.Verbose {
			for src, dst := range config.Mapping.Authors {
				fmt.Printf("  %s -> %s\n", src, dst)
			}
		}
	}
}
