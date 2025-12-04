package main

import (
	"fmt"
	"time"

	"github.com/scottbrown/patina"
	"github.com/spf13/cobra"
)

var (
	listFreshness string
	listRefresh   bool
)

var listCmd = &cobra.Command{
	Use:   "list <organization>",
	Short: "List all repositories with their freshness status",
	Long: `List displays all repositories in a GitHub organization along with
their age and freshness indicator (green, yellow, red).

Use the --freshness flag to filter by status:
  --freshness green   Show only active repos (updated ≤2 months)
  --freshness yellow  Show only aging repos (updated 2-6 months ago)
  --freshness red     Show only stale repos (not updated in >6 months)

Repository data is cached for 30 days. Use --refresh to force a fresh fetch.`,
	Args: cobra.ExactArgs(1),
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVarP(&listFreshness, "freshness", "f", "", "Filter by freshness (green, yellow, red)")
	listCmd.Flags().BoolVarP(&listRefresh, "refresh", "r", false, "Force refresh from GitHub API")
}

func runList(cmd *cobra.Command, args []string) error {
	org := args[0]

	// Validate freshness filter if provided
	var filterFreshness patina.Freshness
	if listFreshness != "" {
		f, ok := patina.ParseFreshness(listFreshness)
		if !ok {
			return fmt.Errorf("invalid freshness value: %q (must be green, yellow, or red)", listFreshness)
		}
		filterFreshness = f
	}

	scanner, err := patina.NewScanner()
	if err != nil {
		return fmt.Errorf("failed to initialize scanner: %w", err)
	}

	result, err := scanner.Scan(org, patina.ScanOptions{Refresh: listRefresh})
	if err != nil {
		return fmt.Errorf("failed to scan organization: %w", err)
	}

	now := time.Now()

	repos := result.Repositories

	// Apply freshness filter if specified
	if filterFreshness != "" {
		repos = patina.FilterByFreshness(repos, filterFreshness, now)
	}

	// Sort by age (oldest first)
	patina.SortByAge(repos)

	// Print header
	if result.FromCache {
		fmt.Printf("Using cached data from %s\n\n", result.FetchedAt.Format("2006-01-02 15:04:05"))
	}

	if filterFreshness != "" {
		fmt.Printf("Repositories in %s (%s): %d\n\n", org, filterFreshness, len(repos))
	} else {
		fmt.Printf("All repositories in %s: %d\n\n", org, len(repos))
	}

	if len(repos) == 0 {
		fmt.Println("No repositories found matching the criteria.")
		return nil
	}

	// Calculate max name length for alignment
	maxNameLen := 0
	for _, repo := range repos {
		if len(repo.Name) > maxNameLen {
			maxNameLen = len(repo.Name)
		}
	}

	// Print each repository
	for _, repo := range repos {
		freshness := patina.CalculateFreshness(repo.LastUpdated, now)
		age := patina.Age(repo.LastUpdated, now)

		fmt.Printf("%s %s%-*s%s  %s\n",
			freshness.Emoji(),
			freshness.Colour(),
			maxNameLen,
			repo.Name,
			patina.ColourReset(),
			age,
		)
	}

	return nil
}
