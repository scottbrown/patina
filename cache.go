package patina

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheDirName  = "patina"
	cacheValidity = 30 * 24 * time.Hour // 30 days
)

var (
	ErrCacheExpired  = errors.New("cache expired")
	ErrCacheNotFound = errors.New("cache not found")
)

// Repository represents a GitHub repository with its last update timestamp.
type Repository struct {
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	LastUpdated time.Time `json:"last_updated"`
	HTMLURL     string    `json:"html_url"`
}

// OrganizationCache holds cached repository data for an organization.
type OrganizationCache struct {
	Organization string       `json:"organization"`
	FetchedAt    time.Time    `json:"fetched_at"`
	Repositories []Repository `json:"repositories"`
}

// Cache provides methods for storing and retrieving organization data.
type Cache struct {
	baseDir string
}

// NewCache creates a new Cache instance with the default cache directory.
func NewCache() (*Cache, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Join(cacheDir, cacheDirName)
	return &Cache{baseDir: baseDir}, nil
}

// NewCacheWithDir creates a Cache with a custom base directory (useful for testing).
func NewCacheWithDir(baseDir string) *Cache {
	return &Cache{baseDir: baseDir}
}

// cacheFilePath returns the path to the cache file for an organization.
func (c *Cache) cacheFilePath(org string) string {
	return filepath.Join(c.baseDir, org+".json")
}

// Save stores organization repository data to the cache.
func (c *Cache) Save(data OrganizationCache) error {
	if err := os.MkdirAll(c.baseDir, 0755); err != nil {
		return err
	}

	data.FetchedAt = time.Now()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.cacheFilePath(data.Organization), jsonData, 0644)
}

// Load retrieves organization repository data from the cache.
// Returns ErrCacheNotFound if no cache exists, or ErrCacheExpired if cache is stale.
func (c *Cache) Load(org string) (OrganizationCache, error) {
	return c.LoadWithTime(org, time.Now())
}

// LoadWithTime retrieves organization data using a specific reference time (for testing).
func (c *Cache) LoadWithTime(org string, now time.Time) (OrganizationCache, error) {
	var data OrganizationCache

	jsonData, err := os.ReadFile(c.cacheFilePath(org))
	if err != nil {
		if os.IsNotExist(err) {
			return data, ErrCacheNotFound
		}
		return data, err
	}

	if err := json.Unmarshal(jsonData, &data); err != nil {
		return data, err
	}

	if now.Sub(data.FetchedAt) > cacheValidity {
		return data, ErrCacheExpired
	}

	return data, nil
}

// IsValid checks if a valid (non-expired) cache exists for the organization.
func (c *Cache) IsValid(org string) bool {
	_, err := c.Load(org)
	return err == nil
}

// IsValidWithTime checks cache validity using a specific reference time.
func (c *Cache) IsValidWithTime(org string, now time.Time) bool {
	_, err := c.LoadWithTime(org, now)
	return err == nil
}

// Clear removes the cache file for an organization.
func (c *Cache) Clear(org string) error {
	err := os.Remove(c.cacheFilePath(org))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// ClearAll removes all cache files.
func (c *Cache) ClearAll() error {
	return os.RemoveAll(c.baseDir)
}

// CacheDir returns the cache directory path.
func (c *Cache) CacheDir() string {
	return c.baseDir
}
