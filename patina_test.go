package patina

import (
	"testing"
	"time"
)

// mockGitHubClient implements GitHubClient for testing.
type mockGitHubClient struct {
	repos []Repository
	err   error
}

func (m *mockGitHubClient) FetchRepositories(org string) ([]Repository, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.repos, nil
}

func TestCalculateSummary(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	repos := []Repository{
		{Name: "fresh1", LastUpdated: now.AddDate(0, 0, -1)},         // green
		{Name: "fresh2", LastUpdated: now.AddDate(0, 0, -30)},        // green
		{Name: "aging1", LastUpdated: now.AddDate(0, 0, -90)},        // yellow
		{Name: "aging2", LastUpdated: now.AddDate(0, 0, -120)},       // yellow
		{Name: "stale1", LastUpdated: now.AddDate(-1, 0, 0)},         // red
		{Name: "stale2", LastUpdated: now.AddDate(-2, 0, 0)},         // red
		{Name: "stale3", LastUpdated: now.AddDate(0, 0, -200)},       // red
	}

	summary := CalculateSummary(repos, now)

	if summary.Green != 2 {
		t.Errorf("Green = %d, want 2", summary.Green)
	}
	if summary.Yellow != 2 {
		t.Errorf("Yellow = %d, want 2", summary.Yellow)
	}
	if summary.Red != 3 {
		t.Errorf("Red = %d, want 3", summary.Red)
	}
	if summary.Total != 7 {
		t.Errorf("Total = %d, want 7", summary.Total)
	}
}

func TestSortByAge(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	repos := []Repository{
		{Name: "middle", LastUpdated: now.AddDate(0, 0, -30)},
		{Name: "newest", LastUpdated: now.AddDate(0, 0, -1)},
		{Name: "oldest", LastUpdated: now.AddDate(-1, 0, 0)},
	}

	SortByAge(repos)

	if repos[0].Name != "oldest" {
		t.Errorf("repos[0].Name = %s, want oldest", repos[0].Name)
	}
	if repos[1].Name != "middle" {
		t.Errorf("repos[1].Name = %s, want middle", repos[1].Name)
	}
	if repos[2].Name != "newest" {
		t.Errorf("repos[2].Name = %s, want newest", repos[2].Name)
	}
}

func TestSortByAgeDesc(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	repos := []Repository{
		{Name: "middle", LastUpdated: now.AddDate(0, 0, -30)},
		{Name: "oldest", LastUpdated: now.AddDate(-1, 0, 0)},
		{Name: "newest", LastUpdated: now.AddDate(0, 0, -1)},
	}

	SortByAgeDesc(repos)

	if repos[0].Name != "newest" {
		t.Errorf("repos[0].Name = %s, want newest", repos[0].Name)
	}
	if repos[1].Name != "middle" {
		t.Errorf("repos[1].Name = %s, want middle", repos[1].Name)
	}
	if repos[2].Name != "oldest" {
		t.Errorf("repos[2].Name = %s, want oldest", repos[2].Name)
	}
}

func TestFilterByFreshness(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	repos := []Repository{
		{Name: "green1", LastUpdated: now.AddDate(0, 0, -1)},
		{Name: "green2", LastUpdated: now.AddDate(0, 0, -30)},
		{Name: "yellow1", LastUpdated: now.AddDate(0, 0, -90)},
		{Name: "red1", LastUpdated: now.AddDate(-1, 0, 0)},
		{Name: "red2", LastUpdated: now.AddDate(-2, 0, 0)},
	}

	tests := []struct {
		freshness Freshness
		wantCount int
		wantNames []string
	}{
		{FreshnessGreen, 2, []string{"green1", "green2"}},
		{FreshnessYellow, 1, []string{"yellow1"}},
		{FreshnessRed, 2, []string{"red1", "red2"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.freshness), func(t *testing.T) {
			filtered := FilterByFreshness(repos, tt.freshness, now)
			if len(filtered) != tt.wantCount {
				t.Errorf("len(filtered) = %d, want %d", len(filtered), tt.wantCount)
			}
		})
	}
}

func TestGetTopStale(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	repos := []Repository{
		{Name: "repo1", LastUpdated: now.AddDate(0, 0, -30)},
		{Name: "repo2", LastUpdated: now.AddDate(-2, 0, 0)},
		{Name: "repo3", LastUpdated: now.AddDate(-1, 0, 0)},
		{Name: "repo4", LastUpdated: now.AddDate(0, 0, -90)},
		{Name: "repo5", LastUpdated: now.AddDate(0, 0, -1)},
	}

	top3 := GetTopStale(repos, 3)

	if len(top3) != 3 {
		t.Fatalf("len(top3) = %d, want 3", len(top3))
	}

	// Should be ordered oldest first
	if top3[0].Name != "repo2" {
		t.Errorf("top3[0].Name = %s, want repo2", top3[0].Name)
	}
	if top3[1].Name != "repo3" {
		t.Errorf("top3[1].Name = %s, want repo3", top3[1].Name)
	}
	if top3[2].Name != "repo4" {
		t.Errorf("top3[2].Name = %s, want repo4", top3[2].Name)
	}
}

func TestGetTopStaleWithFewerRepos(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	repos := []Repository{
		{Name: "repo1", LastUpdated: now.AddDate(0, 0, -30)},
		{Name: "repo2", LastUpdated: now.AddDate(-1, 0, 0)},
	}

	// Request more than available
	top5 := GetTopStale(repos, 5)

	if len(top5) != 2 {
		t.Errorf("len(top5) = %d, want 2", len(top5))
	}
}

func TestGetTopStaleEmpty(t *testing.T) {
	top := GetTopStale(nil, 10)
	if top != nil {
		t.Errorf("GetTopStale(nil) = %v, want nil", top)
	}
}

func TestScannerWithMock(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	mockClient := &mockGitHubClient{
		repos: []Repository{
			{Name: "repo1", FullName: "org/repo1", LastUpdated: now.AddDate(0, 0, -30)},
			{Name: "repo2", FullName: "org/repo2", LastUpdated: now.AddDate(-1, 0, 0)},
		},
	}

	scanner := NewScannerWithDeps(mockClient, cache)

	// First scan should fetch from API
	result, err := scanner.Scan("org", ScanOptions{})
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result.FromCache {
		t.Error("result.FromCache = true on first scan, want false")
	}
	if len(result.Repositories) != 2 {
		t.Errorf("len(result.Repositories) = %d, want 2", len(result.Repositories))
	}

	// Second scan should use cache
	result2, err := scanner.Scan("org", ScanOptions{})
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if !result2.FromCache {
		t.Error("result2.FromCache = false on second scan, want true")
	}

	// Scan with refresh should fetch from API again
	result3, err := scanner.Scan("org", ScanOptions{Refresh: true})
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result3.FromCache {
		t.Error("result3.FromCache = true with Refresh option, want false")
	}
}

func TestCalculateSummaryEmpty(t *testing.T) {
	now := time.Now()
	summary := CalculateSummary(nil, now)

	if summary.Total != 0 {
		t.Errorf("Total = %d, want 0", summary.Total)
	}
	if summary.Green != 0 {
		t.Errorf("Green = %d, want 0", summary.Green)
	}
	if summary.Yellow != 0 {
		t.Errorf("Yellow = %d, want 0", summary.Yellow)
	}
	if summary.Red != 0 {
		t.Errorf("Red = %d, want 0", summary.Red)
	}
}
