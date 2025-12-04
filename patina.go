package patina

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cli/go-gh/v2"
)

const (
	githubAPIBaseURL = "https://api.github.com"
	githubTokenEnv   = "GITHUB_TOKEN"
)

// GitHubClient provides methods for fetching GitHub data.
type GitHubClient interface {
	FetchRepositories(org string) ([]Repository, error)
}

// ghRepo represents the repository data returned by the GitHub API.
type ghRepo struct {
	Name     string    `json:"name"`
	FullName string    `json:"full_name"`
	HTMLURL  string    `json:"html_url"`
	PushedAt time.Time `json:"pushed_at"`
	Archived bool      `json:"archived"`
}

// NewGitHubClient creates a new GitHub client.
// If GITHUB_TOKEN is set, uses direct API calls; otherwise falls back to gh CLI.
func NewGitHubClient() GitHubClient {
	if token := os.Getenv(githubTokenEnv); token != "" {
		return &tokenClient{token: token, httpClient: &http.Client{Timeout: 30 * time.Second}}
	}
	return &ghCLIClient{}
}

// tokenClient implements GitHubClient using a personal access token.
type tokenClient struct {
	token      string
	httpClient *http.Client
}

// FetchRepositories retrieves all repositories using the GitHub API with a token.
func (c *tokenClient) FetchRepositories(org string) ([]Repository, error) {
	var allRepos []Repository
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("%s/orgs/%s/repos?type=all&per_page=%d&page=%d",
			githubAPIBaseURL, org, perPage, page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch repositories: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API error: %s (status %d)", string(body), resp.StatusCode)
		}

		var repos []ghRepo
		if err := json.Unmarshal(body, &repos); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		if len(repos) == 0 {
			break
		}

		for _, repo := range repos {
			if repo.Archived {
				continue
			}
			allRepos = append(allRepos, Repository{
				Name:        repo.Name,
				FullName:    repo.FullName,
				LastUpdated: repo.PushedAt,
				HTMLURL:     repo.HTMLURL,
			})
		}

		// Check if there are more pages
		if !hasNextPage(resp) {
			break
		}
		page++
	}

	return allRepos, nil
}

// hasNextPage checks the Link header for pagination.
func hasNextPage(resp *http.Response) bool {
	link := resp.Header.Get("Link")
	return strings.Contains(link, `rel="next"`)
}

// ghCLIClient implements GitHubClient using the gh CLI.
type ghCLIClient struct{}

// FetchRepositories retrieves all repositories using the gh CLI.
func (c *ghCLIClient) FetchRepositories(org string) ([]Repository, error) {
	var allRepos []Repository
	page := 1
	perPage := 100

	for {
		args := []string{
			"api",
			fmt.Sprintf("/orgs/%s/repos", org),
			"--paginate",
			"-q", ".",
			"-F", "per_page=" + strconv.Itoa(perPage),
			"-F", "page=" + strconv.Itoa(page),
			"-F", "type=all",
		}

		stdout, _, err := gh.Exec(args...)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch repositories: %w", err)
		}

		var repos []ghRepo
		if err := json.Unmarshal(stdout.Bytes(), &repos); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		if len(repos) == 0 {
			break
		}

		for _, repo := range repos {
			if repo.Archived {
				continue
			}
			allRepos = append(allRepos, Repository{
				Name:        repo.Name,
				FullName:    repo.FullName,
				LastUpdated: repo.PushedAt,
				HTMLURL:     repo.HTMLURL,
			})
		}

		// gh --paginate handles pagination automatically, so we only need one iteration
		break
	}

	return allRepos, nil
}

// Scanner provides methods for scanning organizations.
type Scanner struct {
	client GitHubClient
	cache  *Cache
}

// NewScanner creates a new Scanner with the default GitHub client and cache.
func NewScanner() (*Scanner, error) {
	cache, err := NewCache()
	if err != nil {
		return nil, err
	}

	return &Scanner{
		client: NewGitHubClient(),
		cache:  cache,
	}, nil
}

// NewScannerWithDeps creates a Scanner with custom dependencies (useful for testing).
func NewScannerWithDeps(client GitHubClient, cache *Cache) *Scanner {
	return &Scanner{
		client: client,
		cache:  cache,
	}
}

// ScanOptions configures the scan behaviour.
type ScanOptions struct {
	Refresh bool // Force refresh even if cache is valid
}

// ScanResult contains the results of scanning an organization.
type ScanResult struct {
	Organization string
	Repositories []Repository
	FetchedAt    time.Time
	FromCache    bool
}

// Scan retrieves repository data for an organization, using cache if available.
func (s *Scanner) Scan(org string, opts ScanOptions) (*ScanResult, error) {
	now := time.Now()

	// Try to use cache unless refresh is requested
	if !opts.Refresh {
		cached, err := s.cache.Load(org)
		if err == nil {
			return &ScanResult{
				Organization: org,
				Repositories: cached.Repositories,
				FetchedAt:    cached.FetchedAt,
				FromCache:    true,
			}, nil
		}
	}

	// Fetch fresh data
	repos, err := s.client.FetchRepositories(org)
	if err != nil {
		return nil, err
	}

	// Save to cache
	cacheData := OrganizationCache{
		Organization: org,
		Repositories: repos,
		FetchedAt:    now,
	}
	if err := s.cache.Save(cacheData); err != nil {
		// Log but don't fail if cache save fails
		fmt.Printf("Warning: failed to save cache: %v\n", err)
	}

	return &ScanResult{
		Organization: org,
		Repositories: repos,
		FetchedAt:    now,
		FromCache:    false,
	}, nil
}

// FreshnessSummary contains counts of repositories by freshness level.
type FreshnessSummary struct {
	Green  int
	Yellow int
	Red    int
	Total  int
}

// CalculateSummary computes the freshness summary for a list of repositories.
func CalculateSummary(repos []Repository, now time.Time) FreshnessSummary {
	var summary FreshnessSummary
	summary.Total = len(repos)

	for _, repo := range repos {
		switch CalculateFreshness(repo.LastUpdated, now) {
		case FreshnessGreen:
			summary.Green++
		case FreshnessYellow:
			summary.Yellow++
		case FreshnessRed:
			summary.Red++
		}
	}

	return summary
}

// SortByAge sorts repositories by last update time, oldest first.
func SortByAge(repos []Repository) {
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].LastUpdated.Before(repos[j].LastUpdated)
	})
}

// SortByAgeDesc sorts repositories by last update time, newest first.
func SortByAgeDesc(repos []Repository) {
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].LastUpdated.After(repos[j].LastUpdated)
	})
}

// FilterByFreshness returns repositories matching the specified freshness level.
func FilterByFreshness(repos []Repository, freshness Freshness, now time.Time) []Repository {
	var filtered []Repository
	for _, repo := range repos {
		if CalculateFreshness(repo.LastUpdated, now) == freshness {
			filtered = append(filtered, repo)
		}
	}
	return filtered
}

// GetTopStale returns the n oldest repositories.
func GetTopStale(repos []Repository, n int) []Repository {
	if len(repos) == 0 {
		return nil
	}

	// Create a copy to avoid modifying the original
	sorted := make([]Repository, len(repos))
	copy(sorted, repos)
	SortByAge(sorted)

	if n > len(sorted) {
		n = len(sorted)
	}

	return sorted[:n]
}
