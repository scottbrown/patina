package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "patina",
	Short: "Scan GitHub organizations for repository freshness",
	Long: `patina is a CLI tool that scans GitHub organizations to identify
repositories that haven't been updated recently.

It categorizes repositories by freshness:
  🟢 Green:  Updated within last 2 months (active)
  🟡 Yellow: Updated between 2-6 months ago (aging)
  🔴 Red:    Not updated in over 6 months (stale)

Repository data is cached for 30 days to speed up subsequent commands.

Authentication:
  Set GITHUB_TOKEN environment variable, or use 'gh auth login'.`,
	Version: version,
}

func init() {
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(reportCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
