package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Lslightly/pprof2csv/analyzer"
	"github.com/Lslightly/pprof2csv/cmd/lines2md/qlib"
	"github.com/Lslightly/pprof2csv/models"
)

// CLI flags
var (
	inputProfile = flag.String("i", "", "Input pprof profile file")
	queryFile    = flag.String("q", "", "Query file containing lines to analyze")
	outputDir    = flag.String("dir", ".", "Output directory for results")
)

func init() {
	flag.Parse()
}

func validateFlags() error {
	if *inputProfile == "" {
		return fmt.Errorf("input file is required\nUsage: lines2md -i <profile.pprof> -q <query.txt> [-dir <output_dir>]")
	}
	if *queryFile == "" {
		return fmt.Errorf("query file is required\nUsage: lines2md -i <profile.pprof> -q <query.txt> [-dir <output_dir>]")
	}
	return nil
}

func createOutputDirectory() error {
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}
	return nil
}

func writeCollectMD(markdownContent string) error {
	collectFile := filepath.Join(*outputDir, "collect.md")
	collect, err := os.Create(collectFile)
	if err != nil {
		return fmt.Errorf("error creating collect.md: %v", err)
	}
	defer collect.Close()

	fmt.Fprint(collect, markdownContent)
	return nil
}

// Removed convertCSVRowsToSourceLines as it's no longer needed

func main() {
	// Validate command-line flags
	if err := validateFlags(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Create output directory
	if err := createOutputDirectory(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Load and analyze profile data
	allLines, err := analyzer.LoadProfileData(*inputProfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Parse query file
	querySections, err := qlib.ParseQueryFile(*queryFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Find matching lines and calculate cumulative values
	matchedResults := qlib.MatchQueries(querySections, allLines)

	// Generate and write CSV files for each function
	exportMatchedResults(querySections, matchedResults)
}

func exportMatchedResults(querySections []qlib.QuerySection, matchedResults map[string]*models.SourceLine) {
	for _, section := range querySections {
		if len(section.Queries) == 0 {
			continue
		}

		if err := section.WriteFunctionCSV(*outputDir, matchedResults); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}

	// Generate and write markdown content
	markdownContent := qlib.GenerateMarkdownContent(querySections, matchedResults)
	if err := writeCollectMD(markdownContent); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Successfully generated results in %s\n", *outputDir)
}
