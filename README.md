# patina

A CLI tool to scan GitHub organizations and assess repository freshness.

## Overview

`patina` helps you find repositories that haven't been updated recently. It categorizes repositories by freshness:

- 🟢 **Green**: Updated within last 2 months (active)
- 🟡 **Yellow**: Updated between 2-6 months ago (aging)
- 🔴 **Red**: Not updated in over 6 months (stale)

## Installation

### Prerequisites

- [Go](https://golang.org/dl/) 1.21 or later
- [GitHub CLI](https://cli.github.com/) (`gh`) - for authentication (unless using `GITHUB_TOKEN`)
- [Task](https://taskfile.dev/) (optional, for development)

### From Source

```bash
go install github.com/scottbrown/patina/cmd/patina@latest
```

### Build Locally

```bash
git clone https://github.com/scottbrown/patina.git
cd patina
task build
# or: go build -o patina ./cmd/patina
```

## Authentication

`patina` supports two authentication methods:

### Option 1: Environment Variable (recommended for CI/CD)

Set the `GITHUB_TOKEN` environment variable with a personal access token:

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx
patina scan my-org
```

The token requires the `repo` scope to access private repositories, or `public_repo` for public repositories only.

### Option 2: GitHub CLI

If `GITHUB_TOKEN` is not set, `patina` falls back to using the GitHub CLI for authentication:

```bash
gh auth login
patina scan my-org
```

This provides access to both public and private repositories in your organizations.

## Usage

### Scan Command

Scan an organization and display a freshness summary with the top 10 most stale repositories:

```bash
patina scan <organization>
```

Example output:

```
Scanning organization: my-org

Repository Freshness Summary
============================

Total repositories: 42

🟢 Green  (≤2 months):  25
🟡 Yellow (2-6 months): 10
🔴 Red    (>6 months):  7

Top 10 Most Stale Repositories
==============================

 1. 🔴 legacy-api          2 years, 3 months ago
 2. 🔴 old-frontend        1 year, 8 months ago
 3. 🔴 deprecated-utils    1 year, 2 months ago
...
```

### List Command

List all repositories with their age and freshness indicator:

```bash
patina list <organization>
```

Filter by freshness status:

```bash
patina list <organization> --freshness red      # Show only stale repos
patina list <organization> --freshness yellow   # Show only aging repos
patina list <organization> --freshness green    # Show only active repos
```

### Report Command

Generate a standalone HTML report with visual charts and a complete repository table:

```bash
patina report <organization>
patina report <organization> -o my-report.html
```

The report includes:
- Summary cards with colour-coded counts
- Pie chart showing freshness distribution
- Sortable table of all repositories with links

### Options

All commands support:

- `-r, --refresh`: Force refresh from GitHub API (bypass cache)

The list command additionally supports:

- `-f, --freshness <colour>`: Filter by freshness (green, yellow, red)

The report command additionally supports:

- `-o, --output <file>`: Output file path (default: `patina-report.html`)

## Caching

Repository data is cached locally for 30 days to speed up subsequent commands. The cache is stored in:

- macOS: `~/Library/Caches/patina/`
- Linux: `~/.cache/patina/`

Use the `--refresh` flag to force a fresh fetch from GitHub.

## Development

### Running Tests

```bash
task test           # Run all tests
task test:coverage  # Run tests with coverage report
```

### Building

```bash
task build          # Build the binary
task install        # Install to GOPATH/bin
```

### Other Tasks

```bash
task fmt            # Format code
task lint           # Run linter
task tidy           # Tidy go modules
task clean          # Clean build artifacts
```

## Licence

MIT
