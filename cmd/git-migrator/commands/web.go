package commands

import (
	"fmt"

	"github.com/adamf123git/git-migrator/internal/web"
	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web UI server",
	Long: `Start the Git-Migrator web interface for managing migrations
through a browser-based UI.

The web interface provides:
- Dashboard for monitoring all migrations
- Wizard for creating new migrations
- Real-time progress updates via WebSocket
- Configuration editor
- Migration history and logs

By default, the server starts on port 8080, but this can be
customized with the --port flag.`,
	RunE: runWeb,
}

var (
	webPort int
)

func init() {
	rootCmd.AddCommand(webCmd)

	webCmd.Flags().IntVarP(&webPort, "port", "p", 8080, "Port to run the web server on")
}

func runWeb(cmd *cobra.Command, args []string) error {
	// Create server configuration
	config := web.ServerConfig{
		Port:         webPort,
		ConfigPath:   "", // Use default
		DatabasePath: "", // Use default
	}

	// Create server
	server := web.NewServer(config)

	// Display startup message
	fmt.Printf("Starting Git-Migrator web interface...\n")
	fmt.Printf("Open http://localhost:%d in your browser\n\n", webPort)

	// Start server (this blocks until server stops)
	if err := server.Start(); err != nil {
		return fmt.Errorf("failed to start web server: %w", err)
	}

	return nil
}
