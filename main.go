package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Lslightly/pprof2csv/analyzer"
	"github.com/Lslightly/pprof2csv/imexporter"
)

// Version of the tool (set at build time)
var version = "dev"

func main() {
	// Define command-line flags
	var (
		versionFlag = flag.Bool("version", false, "Show version information")
		outputFile  = flag.String("o", "", "Output CSV file (default: stdout)")
		inputFile   = flag.String("i", "", "Input pprof profile file")
	)

	// Parse flags
	flag.Parse()

	// Handle version flag
	if *versionFlag {
		fmt.Printf("pprof2csv version %s\n", version)
		os.Exit(0)
	}

	// Validate input file
	if *inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: input file is required")
		fmt.Fprintln(os.Stderr, "Usage: pprof2csv -i <profile.pprof>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Create components
	csvExporter := imexporter.New()

	// Load profile data
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading profile: %v\n", err)
		os.Exit(1)
	}

	// Analyze profile data
	sourceLines, err := analyzer.Analyze(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing profile: %v\n", err)
		os.Exit(1)
	}

	// Export to CSV
	var output *os.File
	if *outputFile == "" {
		// Output to stdout
		output = os.Stdout
	} else {
		// Create output file
		var err error
		output, err = os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer output.Close()
	}

	err = csvExporter.Export(output, sourceLines)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting CSV: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Successfully converted %s to CSV format\n", *inputFile)
}
