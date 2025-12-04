package patina

import (
	"fmt"
	"time"
)

// Freshness represents the staleness level of a repository.
type Freshness string

const (
	FreshnessGreen  Freshness = "green"
	FreshnessYellow Freshness = "yellow"
	FreshnessRed    Freshness = "red"
)

const (
	yellowThreshold = 2 * 30 * 24 * time.Hour  // ~2 months
	redThreshold    = 6 * 30 * 24 * time.Hour  // ~6 months
)

// CalculateFreshness determines the freshness level based on the last update time.
func CalculateFreshness(lastUpdated time.Time, now time.Time) Freshness {
	age := now.Sub(lastUpdated)

	if age > redThreshold {
		return FreshnessRed
	}
	if age > yellowThreshold {
		return FreshnessYellow
	}
	return FreshnessGreen
}

// FreshnessColour returns the ANSI colour code for terminal output.
func (f Freshness) Colour() string {
	switch f {
	case FreshnessGreen:
		return "\033[32m" // Green
	case FreshnessYellow:
		return "\033[33m" // Yellow
	case FreshnessRed:
		return "\033[31m" // Red
	default:
		return "\033[0m" // Reset
	}
}

// Reset returns the ANSI reset code.
func ColourReset() string {
	return "\033[0m"
}

// Emoji returns the emoji indicator for the freshness level.
func (f Freshness) Emoji() string {
	switch f {
	case FreshnessGreen:
		return "🟢"
	case FreshnessYellow:
		return "🟡"
	case FreshnessRed:
		return "🔴"
	default:
		return "⚪"
	}
}

// String returns the string representation of freshness.
func (f Freshness) String() string {
	return string(f)
}

// ParseFreshness converts a string to a Freshness value.
func ParseFreshness(s string) (Freshness, bool) {
	switch s {
	case "green":
		return FreshnessGreen, true
	case "yellow":
		return FreshnessYellow, true
	case "red":
		return FreshnessRed, true
	default:
		return "", false
	}
}

// Age returns a human-readable age string.
func Age(lastUpdated time.Time, now time.Time) string {
	duration := now.Sub(lastUpdated)

	days := int(duration.Hours() / 24)
	if days < 1 {
		return "today"
	}
	if days == 1 {
		return "1 day ago"
	}
	if days < 30 {
		return pluralize(days, "day") + " ago"
	}

	months := days / 30
	if months < 12 {
		return pluralize(months, "month") + " ago"
	}

	years := months / 12
	remainingMonths := months % 12
	if remainingMonths == 0 {
		return pluralize(years, "year") + " ago"
	}
	return pluralize(years, "year") + ", " + pluralize(remainingMonths, "month") + " ago"
}

func pluralize(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", n, unit)
}
