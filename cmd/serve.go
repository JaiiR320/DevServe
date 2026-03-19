package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jaiir320/devserve/cli"
	"github.com/jaiir320/devserve/client"
	"github.com/jaiir320/devserve/config"
	"github.com/jaiir320/devserve/protocol"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve -p PORT -c COMMAND [--name NAME] [--save]",
	Args:  cobra.NoArgs,
	Short: "Serve your dev server with tailscale",
	Example: `  devserve serve -p 3000 -c "npm run dev"
  devserve serve -n myapp -p 3000 -c "npm run dev" --save`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		command, _ := cmd.Flags().GetString("command")
		name, _ := cmd.Flags().GetString("name")
		save, _ := cmd.Flags().GetBool("save")

		return runServe(name, port, command, save)
	},
}

func init() {
	serveCmd.Flags().IntP("port", "p", 0, "port your dev server listens on")
	serveCmd.Flags().StringP("command", "c", "", "command to start your dev server")
	serveCmd.Flags().StringP("name", "n", "", "process name (defaults to directory name)")
	serveCmd.Flags().BoolP("save", "s", false, "save configuration for later use")

	_ = serveCmd.MarkFlagRequired("port")
	_ = serveCmd.MarkFlagRequired("command")

	rootCmd.AddCommand(serveCmd)
}

func runServe(name string, port int, command string, save bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Default name to slugified directory name
	if name == "" {
		name = slugify(filepath.Base(cwd))
	} else {
		name = slugify(name)
	}

	// Validate port range
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if port < 1024 {
		fmt.Fprintln(os.Stderr, cli.Info(fmt.Sprintf("port %d is privileged, may require elevated permissions", port)))
	}

	var result *protocol.ServeResult
	cli.Spin("Starting process...", func() {
		result, err = client.Serve(name, port, command, cwd)
	})
	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	fmt.Println(cli.RenderServeResult(result))

	if save {
		cfg := config.ProcessConfig{
			Name:      name,
			Port:      port,
			Command:   command,
			Directory: cwd,
		}
		if err := config.SaveConfig(config.ConfigFile, cfg); err != nil {
			fmt.Fprintln(os.Stderr, cli.Error(fmt.Sprintf("failed to save config: %s", err)))
		} else {
			fmt.Println(cli.Success(fmt.Sprintf("config '%s' saved", name)))
		}
	}

	return nil
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// slugify converts a string into a URL-friendly slug.
// e.g. "My Project.v2" -> "my-project-v2"
func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
