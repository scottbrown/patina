package main

import (
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/scottbrown/patina"
	"github.com/spf13/cobra"
)

var (
	reportOutput  string
	reportRefresh bool
)

var reportCmd = &cobra.Command{
	Use:   "report <organization>",
	Short: "Generate an HTML report of repository freshness",
	Long: `Report generates a standalone HTML file containing a visual summary
of repository freshness for a GitHub organization.

The report includes:
  - Freshness summary with colour-coded counts
  - Visual pie chart of the distribution
  - Complete table of all repositories with links

Example:
  patina report my-org -o report.html`,
	Args: cobra.ExactArgs(1),
	RunE: runReport,
}

func init() {
	reportCmd.Flags().StringVarP(&reportOutput, "output", "o", "patina-report.html", "Output file path")
	reportCmd.Flags().BoolVarP(&reportRefresh, "refresh", "r", false, "Force refresh from GitHub API")
}

type reportData struct {
	Organization string
	GeneratedAt  string
	Summary      patina.FreshnessSummary
	Repositories []repoData
	GreenPct     float64
	YellowPct    float64
	RedPct       float64
}

type repoData struct {
	Name        string
	FullName    string
	URL         string
	Age         string
	Freshness   string
	ColourClass string
}

func runReport(cmd *cobra.Command, args []string) error {
	org := args[0]

	scanner, err := patina.NewScanner()
	if err != nil {
		return fmt.Errorf("failed to initialize scanner: %w", err)
	}

	fmt.Printf("Scanning organization: %s\n", org)

	result, err := scanner.Scan(org, patina.ScanOptions{Refresh: reportRefresh})
	if err != nil {
		return fmt.Errorf("failed to scan organization: %w", err)
	}

	now := time.Now()

	if result.FromCache {
		fmt.Printf("Using cached data from %s\n", result.FetchedAt.Format("2006-01-02 15:04:05"))
	}

	// Prepare report data
	summary := patina.CalculateSummary(result.Repositories, now)

	// Sort by age (oldest first)
	patina.SortByAge(result.Repositories)

	var repos []repoData
	for _, repo := range result.Repositories {
		freshness := patina.CalculateFreshness(repo.LastUpdated, now)
		repos = append(repos, repoData{
			Name:        repo.Name,
			FullName:    repo.FullName,
			URL:         repo.HTMLURL,
			Age:         patina.Age(repo.LastUpdated, now),
			Freshness:   string(freshness),
			ColourClass: string(freshness),
		})
	}

	// Calculate percentages for pie chart
	var greenPct, yellowPct, redPct float64
	if summary.Total > 0 {
		greenPct = float64(summary.Green) / float64(summary.Total) * 100
		yellowPct = float64(summary.Yellow) / float64(summary.Total) * 100
		redPct = float64(summary.Red) / float64(summary.Total) * 100
	}

	data := reportData{
		Organization: org,
		GeneratedAt:  now.Format("2006-01-02 15:04:05"),
		Summary:      summary,
		Repositories: repos,
		GreenPct:     greenPct,
		YellowPct:    yellowPct,
		RedPct:       redPct,
	}

	// Generate HTML
	funcMap := template.FuncMap{
		"add": func(a, b interface{}) float64 {
			var af, bf float64
			switch v := a.(type) {
			case int:
				af = float64(v)
			case float64:
				af = v
			}
			switch v := b.(type) {
			case int:
				bf = float64(v)
			case float64:
				bf = v
			}
			return af + bf
		},
	}
	tmpl, err := template.New("report").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	f, err := os.Create(reportOutput)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	fmt.Printf("Report generated: %s\n", reportOutput)
	return nil
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Repository Freshness Report - {{.Organization}}</title>
    <style>
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f5f5f5;
            padding: 2rem;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        h1 {
            color: #24292e;
            margin-bottom: 0.5rem;
        }
        .subtitle {
            color: #586069;
            margin-bottom: 2rem;
        }
        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
            margin-bottom: 2rem;
        }
        .summary-card {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            text-align: center;
        }
        .summary-card.green { border-left: 4px solid #28a745; }
        .summary-card.yellow { border-left: 4px solid #ffc107; }
        .summary-card.red { border-left: 4px solid #dc3545; }
        .summary-card.total { border-left: 4px solid #6c757d; }
        .summary-number {
            font-size: 2.5rem;
            font-weight: bold;
        }
        .summary-card.green .summary-number { color: #28a745; }
        .summary-card.yellow .summary-number { color: #b8860b; }
        .summary-card.red .summary-number { color: #dc3545; }
        .summary-label {
            color: #586069;
            font-size: 0.9rem;
        }
        .chart-section {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            margin-bottom: 2rem;
        }
        .chart-title {
            font-size: 1.1rem;
            margin-bottom: 1rem;
            color: #24292e;
        }
        .pie-chart {
            width: 200px;
            height: 200px;
            border-radius: 50%;
            margin: 0 auto;
            background: conic-gradient(
                #28a745 0deg {{printf "%.1f" .GreenPct}}%,
                #ffc107 {{printf "%.1f" .GreenPct}}% {{printf "%.1f" (add .GreenPct .YellowPct)}}%,
                #dc3545 {{printf "%.1f" (add .GreenPct .YellowPct)}}% 100%
            );
        }
        .legend {
            display: flex;
            justify-content: center;
            gap: 2rem;
            margin-top: 1rem;
        }
        .legend-item {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        .legend-colour {
            width: 16px;
            height: 16px;
            border-radius: 3px;
        }
        .legend-colour.green { background: #28a745; }
        .legend-colour.yellow { background: #ffc107; }
        .legend-colour.red { background: #dc3545; }
        .table-section {
            background: white;
            border-radius: 8px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .table-header {
            padding: 1rem 1.5rem;
            border-bottom: 1px solid #e1e4e8;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th {
            text-align: left;
            padding: 0.75rem 1.5rem;
            background: #f6f8fa;
            border-bottom: 1px solid #e1e4e8;
            font-weight: 600;
            color: #24292e;
        }
        td {
            padding: 0.75rem 1.5rem;
            border-bottom: 1px solid #e1e4e8;
        }
        tr:hover {
            background: #f6f8fa;
        }
        .status-badge {
            display: inline-block;
            padding: 0.25rem 0.75rem;
            border-radius: 12px;
            font-size: 0.85rem;
            font-weight: 500;
        }
        .status-badge.green {
            background: #dcffe4;
            color: #22863a;
        }
        .status-badge.yellow {
            background: #fff3cd;
            color: #856404;
        }
        .status-badge.red {
            background: #ffeef0;
            color: #cb2431;
        }
        a {
            color: #0366d6;
            text-decoration: none;
        }
        a:hover {
            text-decoration: underline;
        }
        .footer {
            text-align: center;
            margin-top: 2rem;
            color: #586069;
            font-size: 0.85rem;
        }
        .table-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .filter-buttons {
            display: flex;
            gap: 0.5rem;
        }
        .filter-btn {
            padding: 0.4rem 0.8rem;
            border: 1px solid #e1e4e8;
            border-radius: 6px;
            background: white;
            cursor: pointer;
            font-size: 0.85rem;
            transition: all 0.15s ease;
        }
        .filter-btn:hover {
            background: #f6f8fa;
        }
        .filter-btn.active {
            border-color: #0366d6;
            background: #f1f8ff;
            color: #0366d6;
        }
        .filter-btn.green.active {
            border-color: #28a745;
            background: #dcffe4;
            color: #22863a;
        }
        .filter-btn.yellow.active {
            border-color: #b8860b;
            background: #fff3cd;
            color: #856404;
        }
        .filter-btn.red.active {
            border-color: #dc3545;
            background: #ffeef0;
            color: #cb2431;
        }
        tr.hidden {
            display: none;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Repository Freshness Report</h1>
        <p class="subtitle">Organisation: <strong>{{.Organization}}</strong> | Generated: {{.GeneratedAt}}</p>

        <div class="summary-grid">
            <div class="summary-card total">
                <div class="summary-number">{{.Summary.Total}}</div>
                <div class="summary-label">Total Repositories</div>
            </div>
            <div class="summary-card green">
                <div class="summary-number">{{.Summary.Green}}</div>
                <div class="summary-label">Active (≤2 months)</div>
            </div>
            <div class="summary-card yellow">
                <div class="summary-number">{{.Summary.Yellow}}</div>
                <div class="summary-label">Aging (2-6 months)</div>
            </div>
            <div class="summary-card red">
                <div class="summary-number">{{.Summary.Red}}</div>
                <div class="summary-label">Stale (>6 months)</div>
            </div>
        </div>

        {{if gt .Summary.Total 0}}
        <div class="chart-section">
            <div class="chart-title">Distribution</div>
            <div class="pie-chart"></div>
            <div class="legend">
                <div class="legend-item">
                    <div class="legend-colour green"></div>
                    <span>Active ({{printf "%.1f" .GreenPct}}%)</span>
                </div>
                <div class="legend-item">
                    <div class="legend-colour yellow"></div>
                    <span>Aging ({{printf "%.1f" .YellowPct}}%)</span>
                </div>
                <div class="legend-item">
                    <div class="legend-colour red"></div>
                    <span>Stale ({{printf "%.1f" .RedPct}}%)</span>
                </div>
            </div>
        </div>
        {{end}}

        <div class="table-section">
            <div class="table-header">
                <div><strong>All Repositories</strong> (sorted by age, oldest first)</div>
                <div class="filter-buttons">
                    <button class="filter-btn active" data-filter="all" onclick="filterTable('all')">All</button>
                    <button class="filter-btn red" data-filter="red" onclick="filterTable('red')">Red</button>
                    <button class="filter-btn yellow" data-filter="yellow" onclick="filterTable('yellow')">Yellow</button>
                    <button class="filter-btn green" data-filter="green" onclick="filterTable('green')">Green</button>
                </div>
            </div>
            <table id="repo-table">
                <thead>
                    <tr>
                        <th>#</th>
                        <th>Repository</th>
                        <th>Last Updated</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    {{range $i, $repo := .Repositories}}
                    <tr data-status="{{$repo.ColourClass}}">
                        <td>{{add $i 1}}</td>
                        <td><a href="{{$repo.URL}}" target="_blank">{{$repo.FullName}}</a></td>
                        <td>{{$repo.Age}}</td>
                        <td><span class="status-badge {{$repo.ColourClass}}">{{$repo.Freshness}}</span></td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>

        <div class="footer">
            Generated by <strong>patina</strong>
        </div>
    </div>

    <script>
        function filterTable(status) {
            const rows = document.querySelectorAll('#repo-table tbody tr');
            const buttons = document.querySelectorAll('.filter-btn');

            buttons.forEach(btn => {
                btn.classList.remove('active');
                if (btn.dataset.filter === status) {
                    btn.classList.add('active');
                }
            });

            rows.forEach(row => {
                if (status === 'all' || row.dataset.status === status) {
                    row.classList.remove('hidden');
                } else {
                    row.classList.add('hidden');
                }
            });
        }
    </script>
</body>
</html>`
