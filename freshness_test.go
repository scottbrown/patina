package patina

import (
	"testing"
	"time"
)

func TestCalculateFreshness(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		lastUpdated time.Time
		want        Freshness
	}{
		{
			name:        "updated today is green",
			lastUpdated: now,
			want:        FreshnessGreen,
		},
		{
			name:        "updated 1 day ago is green",
			lastUpdated: now.AddDate(0, 0, -1),
			want:        FreshnessGreen,
		},
		{
			name:        "updated 30 days ago is green",
			lastUpdated: now.AddDate(0, 0, -30),
			want:        FreshnessGreen,
		},
		{
			name:        "updated 60 days ago is green (at threshold)",
			lastUpdated: now.AddDate(0, 0, -60),
			want:        FreshnessGreen,
		},
		{
			name:        "updated 61 days ago is yellow",
			lastUpdated: now.AddDate(0, 0, -61),
			want:        FreshnessYellow,
		},
		{
			name:        "updated 90 days ago is yellow",
			lastUpdated: now.AddDate(0, 0, -90),
			want:        FreshnessYellow,
		},
		{
			name:        "updated 180 days ago is yellow (at threshold)",
			lastUpdated: now.AddDate(0, 0, -180),
			want:        FreshnessYellow,
		},
		{
			name:        "updated 181 days ago is red",
			lastUpdated: now.AddDate(0, 0, -181),
			want:        FreshnessRed,
		},
		{
			name:        "updated 1 year ago is red",
			lastUpdated: now.AddDate(-1, 0, 0),
			want:        FreshnessRed,
		},
		{
			name:        "updated 2 years ago is red",
			lastUpdated: now.AddDate(-2, 0, 0),
			want:        FreshnessRed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateFreshness(tt.lastUpdated, now)
			if got != tt.want {
				t.Errorf("CalculateFreshness() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFreshnessColour(t *testing.T) {
	tests := []struct {
		freshness Freshness
		want      string
	}{
		{FreshnessGreen, "\033[32m"},
		{FreshnessYellow, "\033[33m"},
		{FreshnessRed, "\033[31m"},
		{Freshness("unknown"), "\033[0m"},
	}

	for _, tt := range tests {
		t.Run(string(tt.freshness), func(t *testing.T) {
			got := tt.freshness.Colour()
			if got != tt.want {
				t.Errorf("Colour() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFreshnessEmoji(t *testing.T) {
	tests := []struct {
		freshness Freshness
		want      string
	}{
		{FreshnessGreen, "🟢"},
		{FreshnessYellow, "🟡"},
		{FreshnessRed, "🔴"},
		{Freshness("unknown"), "⚪"},
	}

	for _, tt := range tests {
		t.Run(string(tt.freshness), func(t *testing.T) {
			got := tt.freshness.Emoji()
			if got != tt.want {
				t.Errorf("Emoji() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFreshness(t *testing.T) {
	tests := []struct {
		input   string
		want    Freshness
		wantOk  bool
	}{
		{"green", FreshnessGreen, true},
		{"yellow", FreshnessYellow, true},
		{"red", FreshnessRed, true},
		{"invalid", "", false},
		{"GREEN", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseFreshness(tt.input)
			if got != tt.want || ok != tt.wantOk {
				t.Errorf("ParseFreshness(%q) = (%v, %v), want (%v, %v)", tt.input, got, ok, tt.want, tt.wantOk)
			}
		})
	}
}

func TestAge(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		lastUpdated time.Time
		want        string
	}{
		{
			name:        "same day",
			lastUpdated: now.Add(-1 * time.Hour),
			want:        "today",
		},
		{
			name:        "1 day ago",
			lastUpdated: now.AddDate(0, 0, -1),
			want:        "1 day ago",
		},
		{
			name:        "5 days ago",
			lastUpdated: now.AddDate(0, 0, -5),
			want:        "5 days ago",
		},
		{
			name:        "29 days ago",
			lastUpdated: now.AddDate(0, 0, -29),
			want:        "29 days ago",
		},
		{
			name:        "30 days ago is 1 month",
			lastUpdated: now.AddDate(0, 0, -30),
			want:        "1 month ago",
		},
		{
			name:        "60 days ago is 2 months",
			lastUpdated: now.AddDate(0, 0, -60),
			want:        "2 months ago",
		},
		{
			name:        "365 days ago is 1 year",
			lastUpdated: now.AddDate(0, 0, -365),
			want:        "1 year ago",
		},
		{
			name:        "400 days ago is 1 year, 1 month",
			lastUpdated: now.AddDate(0, 0, -400),
			want:        "1 year, 1 month ago",
		},
		{
			name:        "730 days ago is 2 years",
			lastUpdated: now.AddDate(0, 0, -730),
			want:        "2 years ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Age(tt.lastUpdated, now)
			if got != tt.want {
				t.Errorf("Age() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFreshnessString(t *testing.T) {
	if FreshnessGreen.String() != "green" {
		t.Errorf("FreshnessGreen.String() = %q, want %q", FreshnessGreen.String(), "green")
	}
	if FreshnessYellow.String() != "yellow" {
		t.Errorf("FreshnessYellow.String() = %q, want %q", FreshnessYellow.String(), "yellow")
	}
	if FreshnessRed.String() != "red" {
		t.Errorf("FreshnessRed.String() = %q, want %q", FreshnessRed.String(), "red")
	}
}

func TestColourReset(t *testing.T) {
	if ColourReset() != "\033[0m" {
		t.Errorf("ColourReset() = %q, want %q", ColourReset(), "\033[0m")
	}
}
