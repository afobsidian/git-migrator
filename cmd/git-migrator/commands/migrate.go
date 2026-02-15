package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adamf123git/git-migrator/internal/core"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run a migration from CVS/SVN to Git",
	Long: `Execute a repository migration using a configuration file.

The migration preserves complete history including:
- All commits with authors, dates, and messages
- Branch structure and history
- Tags and their references
- File changes and content

Use --dry-run to preview the migration without making changes.
Use --resume to continue an interrupted migration.

Example usage:
  git-migrator migrate --config migration-config.yaml
  git-migrator migrate --config config.yaml --dry-run --verbose
  git-migrator migrate --config config.yaml --resume`,
	RunE: runMigrate,
}

var (
	migrateConfigFile string
	migrateDryRun     bool
	migrateVerbose    bool
	migrateResume     bool
)

// ConfigFile represents the YAML configuration file structure
type ConfigFile struct {
	Source struct {
		Type   string `yaml:"type"`
		Path   string `yaml:"path"`
		Module string `yaml:"module"`
	} `yaml:"source"`

	Target struct {
		Type   string `yaml:"type"`
		Path   string `yaml:"path"`
		Remote string `yaml:"remote"`
	} `yaml:"target"`

	Mapping struct {
		Authors  map[string]string `yaml:"authors"`
		Branches map[string]string `yaml:"branches"`
		Tags     map[string]string `yaml:"tags"`
	} `yaml:"mapping"`

	Options struct {
		DryRun    bool `yaml:"dryRun"`
		Verbose   bool `yaml:"verbose"`
		ChunkSize int  `yaml:"chunkSize"`
		Resume    bool `yaml:"resume"`
	} `yaml:"options"`
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().StringVarP(&migrateConfigFile, "config", "c", "", "Path to configuration file (required)")
	migrateCmd.Flags().BoolVarP(&migrateDryRun, "dry-run", "d", false, "Preview migration without making changes")
	migrateCmd.Flags().BoolVarP(&migrateVerbose, "verbose", "v", false, "Show detailed progress information")
	migrateCmd.Flags().BoolVarP(&migrateResume, "resume", "r", false, "Resume an interrupted migration")

	var err = migrateCmd.MarkFlagRequired("config")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marking flag as required: %v\n", err)
		os.Exit(1)
	}
}

func runMigrate(cmd *cobra.Command, args []string) error {
	// Load configuration file
	config, err := loadConfigFile(migrateConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Command-line flags override config file settings
	if migrateDryRun {
		config.Options.DryRun = true
	}
	if migrateVerbose {
		config.Options.Verbose = true
	}
	if migrateResume {
		config.Options.Resume = true
	}

	// Convert config file to migration config
	migrationConfig := &core.MigrationConfig{
		SourceType: config.Source.Type,
		SourcePath: config.Source.Path,
		TargetPath: config.Target.Path,
		AuthorMap:  config.Mapping.Authors,
		BranchMap:  config.Mapping.Branches,
		TagMap:     config.Mapping.Tags,
		DryRun:     config.Options.DryRun,
		Resume:     config.Options.Resume,
		ChunkSize:  config.Options.ChunkSize,
	}

	// Set default chunk size if not specified
	if migrationConfig.ChunkSize == 0 {
		migrationConfig.ChunkSize = 100
	}

	// Set state file path
	stateFile := filepath.Join(
		filepath.Dir(migrationConfig.TargetPath),
		".git-migrator-state.db",
	)
	migrationConfig.StateFile = stateFile

	// Display migration information
	if config.Options.Verbose || config.Options.DryRun {
		printMigrationInfo(config, migrationConfig)
	}

	if config.Options.DryRun {
		fmt.Println("\n🔍 DRY RUN MODE - No changes will be made")
	}

	// Create migrator
	migrator := core.NewMigrator(migrationConfig)

	// Run migration
	fmt.Println("\nStarting migration...")
	if err := migrator.Run(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if config.Options.DryRun {
		fmt.Println("\n✓ Dry run completed successfully")
		fmt.Println("Run without --dry-run to perform actual migration")
	} else {
		fmt.Println("\n✓ Migration completed successfully!")
	}

	return nil
}

func loadConfigFile(path string) (*ConfigFile, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config ConfigFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate required fields
	if config.Source.Type == "" {
		return nil, fmt.Errorf("source.type is required")
	}
	if config.Source.Path == "" {
		return nil, fmt.Errorf("source.path is required")
	}
	if config.Target.Path == "" {
		return nil, fmt.Errorf("target.path is required")
	}

	// Set defaults
	if config.Target.Type == "" {
		config.Target.Type = "git"
	}
	if config.Mapping.Authors == nil {
		config.Mapping.Authors = make(map[string]string)
	}
	if config.Mapping.Branches == nil {
		config.Mapping.Branches = make(map[string]string)
	}
	if config.Mapping.Tags == nil {
		config.Mapping.Tags = make(map[string]string)
	}

	return &config, nil
}

func printMigrationInfo(config *ConfigFile, migrationConfig *core.MigrationConfig) {
	fmt.Println("\nMigration Configuration")
	fmt.Println("======================")
	fmt.Printf("Source Type:    %s\n", config.Source.Type)
	fmt.Printf("Source Path:    %s\n", config.Source.Path)
	if config.Source.Module != "" {
		fmt.Printf("Source Module:  %s\n", config.Source.Module)
	}
	fmt.Printf("Target Path:    %s\n", config.Target.Path)
	if config.Target.Remote != "" {
		fmt.Printf("Target Remote:  %s\n", config.Target.Remote)
	}
	fmt.Printf("Dry Run:        %v\n", config.Options.DryRun)
	fmt.Printf("Resume:         %v\n", config.Options.Resume)
	fmt.Printf("Chunk Size:     %d\n", config.Options.ChunkSize)

	if len(config.Mapping.Authors) > 0 {
		fmt.Printf("\nAuthor Mappings: %d\n", len(config.Mapping.Authors))
		if config.Options.Verbose {
			for cvs, git := range config.Mapping.Authors {
				fmt.Printf("  %s -> %s\n", cvs, git)
			}
		}
	}

	if len(config.Mapping.Branches) > 0 {
		fmt.Printf("\nBranch Mappings: %d\n", len(config.Mapping.Branches))
		if config.Options.Verbose {
			for cvs, git := range config.Mapping.Branches {
				fmt.Printf("  %s -> %s\n", cvs, git)
			}
		}
	}

	if len(config.Mapping.Tags) > 0 {
		fmt.Printf("\nTag Mappings: %d\n", len(config.Mapping.Tags))
		if config.Options.Verbose {
			for cvs, git := range config.Mapping.Tags {
				fmt.Printf("  %s -> %s\n", cvs, git)
			}
		}
	}
}
