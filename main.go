package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vegasq/ostop/internal/client"
	"github.com/vegasq/ostop/internal/ui"
)

var (
	version = "0.1.0"
)

func main() {
	// CLI flags
	endpoint := flag.String("endpoint", "", "OpenSearch endpoint URL (required)")
	region := flag.String("region", "", "AWS region (required for AWS OpenSearch)")
	profile := flag.String("profile", "", "AWS profile name (optional)")
	insecure := flag.Bool("insecure", false, "Skip TLS verification (development only)")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("ostop version %s\n", version)
		os.Exit(0)
	}

	// Validate required flags
	if *endpoint == "" {
		fmt.Fprintln(os.Stderr, "Error: --endpoint is required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nExamples:")
		fmt.Fprintln(os.Stderr, "  Local:  ostop --endpoint http://localhost:9200")
		fmt.Fprintln(os.Stderr, "  AWS:    ostop --endpoint https://search-xxx.us-east-1.es.amazonaws.com --region us-east-1")
		os.Exit(1)
	}

	// Create OpenSearch client
	osClient, err := client.NewClient(*endpoint, *region, *profile, *insecure)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating OpenSearch client: %v\n", err)
		os.Exit(1)
	}

	// Initialize Bubble Tea application
	app := ui.NewApp(osClient, *endpoint)
	p := tea.NewProgram(app, tea.WithAltScreen())

	// Run the TUI
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
