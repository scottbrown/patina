package main

import (
	"fmt"
	"time"

	"github.com/scottbrown/patina"
	"github.com/spf13/cobra"
)

var scanRefresh bool

var scanCmd = &cobra.Command{
	Use:   "scan <organization>",
	Short: "Scan a GitHub organization for stale repositories",
	Long: `Scan retrieves all repositories for a GitHub organization and displays
a freshness summary showing how many repositories fall into each category:

  🟢 Green:  Updated within last 2 months (active)
  🟡 Yellow: Updated between 2-6 months ago (aging)
  🔴 Red:    Not updated in over 6 months (stale)

The scan also lists the top 10 most stale repositories.

Repository data is cached for 30 days to speed up subsequent commands.
Use --refresh to force a fresh fetch from GitHub.`,
	Args: cobra.ExactArgs(1),
	RunE: runScan,
}

func init() {
	scanCmd.Flags().BoolVarP(&scanRefresh, "refresh", "r", false, "Force refresh from GitHub API")
}

func runScan(cmd *cobra.Command, args []string) error {
	org := args[0]

	scanner, err := patina.NewScanner()
	if err != nil {
		return fmt.Errorf("failed to initialize scanner: %w", err)
	}

	fmt.Printf("Scanning organization: %s\n", org)
	if scanRefresh {
		fmt.Println("(forcing refresh from GitHub API)")
	}
	fmt.Println()

	result, err := scanner.Scan(org, patina.ScanOptions{Refresh: scanRefresh})
	if err != nil {
		return fmt.Errorf("failed to scan organization: %w", err)
	}

	now := time.Now()

	if result.FromCache {
		fmt.Printf("Using cached data from %s\n\n", result.FetchedAt.Format("2006-01-02 15:04:05"))
	}

	// Calculate and display summary
	summary := patina.CalculateSummary(result.Repositories, now)
	printSummary(summary)

	// Display top stale repositories
	fmt.Println()
	printTopStale(result.Repositories, now, 10)

	return nil
}

func printSummary(summary patina.FreshnessSummary) {
	fmt.Println("Repository Freshness Summary")
	fmt.Println("============================")
	fmt.Println()
	fmt.Printf("Total repositories: %d\n\n", summary.Total)

	green := patina.FreshnessGreen
	yellow := patina.FreshnessYellow
	red := patina.FreshnessRed

	fmt.Printf("%s %sGreen%s  (≤2 months):  %d\n",
		green.Emoji(), green.Colour(), patina.ColourReset(), summary.Green)
	fmt.Printf("%s %sYellow%s (2-6 months): %d\n",
		yellow.Emoji(), yellow.Colour(), patina.ColourReset(), summary.Yellow)
	fmt.Printf("%s %sRed%s    (>6 months):  %d\n",
		red.Emoji(), red.Colour(), patina.ColourReset(), summary.Red)
}

func printTopStale(repos []patina.Repository, now time.Time, n int) {
	topStale := patina.GetTopStale(repos, n)

	if len(topStale) == 0 {
		fmt.Println("No repositories found.")
		return
	}

	fmt.Printf("Top %d Most Stale Repositories\n", len(topStale))
	fmt.Println("==============================")
	fmt.Println()

	maxNameLen := 0
	for _, repo := range topStale {
		if len(repo.Name) > maxNameLen {
			maxNameLen = len(repo.Name)
		}
	}

	for i, repo := range topStale {
		freshness := patina.CalculateFreshness(repo.LastUpdated, now)
		age := patina.Age(repo.LastUpdated, now)

		fmt.Printf("%2d. %s %s%-*s%s  %s\n",
			i+1,
			freshness.Emoji(),
			freshness.Colour(),
			maxNameLen,
			repo.Name,
			patina.ColourReset(),
			age,
		)
	}
}
