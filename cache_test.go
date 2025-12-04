package patina

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	data := OrganizationCache{
		Organization: "test-org",
		Repositories: []Repository{
			{
				Name:        "repo1",
				FullName:    "test-org/repo1",
				LastUpdated: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				HTMLURL:     "https://github.com/test-org/repo1",
			},
			{
				Name:        "repo2",
				FullName:    "test-org/repo2",
				LastUpdated: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
				HTMLURL:     "https://github.com/test-org/repo2",
			},
		},
	}

	if err := cache.Save(data); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := cache.Load("test-org")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Organization != data.Organization {
		t.Errorf("Organization = %v, want %v", loaded.Organization, data.Organization)
	}

	if len(loaded.Repositories) != len(data.Repositories) {
		t.Errorf("len(Repositories) = %v, want %v", len(loaded.Repositories), len(data.Repositories))
	}

	for i, repo := range loaded.Repositories {
		if repo.Name != data.Repositories[i].Name {
			t.Errorf("Repositories[%d].Name = %v, want %v", i, repo.Name, data.Repositories[i].Name)
		}
		if repo.FullName != data.Repositories[i].FullName {
			t.Errorf("Repositories[%d].FullName = %v, want %v", i, repo.FullName, data.Repositories[i].FullName)
		}
	}
}

func TestCacheNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	_, err := cache.Load("nonexistent-org")
	if err != ErrCacheNotFound {
		t.Errorf("Load() error = %v, want %v", err, ErrCacheNotFound)
	}
}

func TestCacheExpired(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	data := OrganizationCache{
		Organization: "test-org",
		Repositories: []Repository{},
	}

	if err := cache.Save(data); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load with time 31 days in the future
	futureTime := time.Now().Add(31 * 24 * time.Hour)
	_, err := cache.LoadWithTime("test-org", futureTime)
	if err != ErrCacheExpired {
		t.Errorf("LoadWithTime() error = %v, want %v", err, ErrCacheExpired)
	}
}

func TestCacheIsValid(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	if cache.IsValid("test-org") {
		t.Error("IsValid() = true for nonexistent cache, want false")
	}

	data := OrganizationCache{
		Organization: "test-org",
		Repositories: []Repository{},
	}

	if err := cache.Save(data); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if !cache.IsValid("test-org") {
		t.Error("IsValid() = false after Save(), want true")
	}

	// Check with expired time
	futureTime := time.Now().Add(31 * 24 * time.Hour)
	if cache.IsValidWithTime("test-org", futureTime) {
		t.Error("IsValidWithTime() = true for expired cache, want false")
	}
}

func TestCacheClear(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	data := OrganizationCache{
		Organization: "test-org",
		Repositories: []Repository{},
	}

	if err := cache.Save(data); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := cache.Clear("test-org"); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if cache.IsValid("test-org") {
		t.Error("IsValid() = true after Clear(), want false")
	}
}

func TestCacheClearNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	// Clearing nonexistent cache should not error
	if err := cache.Clear("nonexistent-org"); err != nil {
		t.Errorf("Clear() error = %v for nonexistent org, want nil", err)
	}
}

func TestCacheClearAll(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	orgs := []string{"org1", "org2", "org3"}
	for _, org := range orgs {
		data := OrganizationCache{
			Organization: org,
			Repositories: []Repository{},
		}
		if err := cache.Save(data); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	if err := cache.ClearAll(); err != nil {
		t.Fatalf("ClearAll() error = %v", err)
	}

	for _, org := range orgs {
		if cache.IsValid(org) {
			t.Errorf("IsValid(%q) = true after ClearAll(), want false", org)
		}
	}
}

func TestCacheDir(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	if cache.CacheDir() != tmpDir {
		t.Errorf("CacheDir() = %v, want %v", cache.CacheDir(), tmpDir)
	}
}

func TestCacheFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	expected := filepath.Join(tmpDir, "my-org.json")
	if got := cache.cacheFilePath("my-org"); got != expected {
		t.Errorf("cacheFilePath() = %v, want %v", got, expected)
	}
}

func TestNewCache(t *testing.T) {
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	cacheDir, _ := os.UserCacheDir()
	expected := filepath.Join(cacheDir, cacheDirName)
	if cache.CacheDir() != expected {
		t.Errorf("CacheDir() = %v, want %v", cache.CacheDir(), expected)
	}
}

func TestCacheFetchedAtIsSetOnSave(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCacheWithDir(tmpDir)

	beforeSave := time.Now()

	data := OrganizationCache{
		Organization: "test-org",
		Repositories: []Repository{},
	}

	if err := cache.Save(data); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	afterSave := time.Now()

	loaded, err := cache.Load("test-org")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.FetchedAt.Before(beforeSave) || loaded.FetchedAt.After(afterSave) {
		t.Errorf("FetchedAt = %v, want between %v and %v", loaded.FetchedAt, beforeSave, afterSave)
	}
}
