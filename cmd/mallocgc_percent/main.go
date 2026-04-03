package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Lslightly/pprof2csv/analyzer"
)

var (
	inputProfile = flag.String("i", "", "Input pprof profile file")
	showFrom     = flag.String("show_from", "", "Only include mallocgc samples whose stacktrace contains this function (for numerator)")
	denomFunc    = flag.String("denom_func", "", "Function name to use as denominator (default: total profile sample time/show_from if the option is provided)")
	format       = flag.String("format", "text", "Output format: text or json")
)

func validateFlags() error {
	if *inputProfile == "" {
		return fmt.Errorf("input file is required\nUsage: mallocgc_percent -i <profile.pprof> [-show_from <function>] [-denom_func <function>] [-format text|json]")
	}
	if *format != "text" && *format != "json" {
		return fmt.Errorf("format must be 'text' or 'json'")
	}
	if *showFrom != "" && *denomFunc == "" {
		*denomFunc = *showFrom
	}
	return nil
}

type Result struct {
	MallocgcTime  time.Duration `json:"mallocgc_time"`
	Denominator   time.Duration `json:"denominator"`
	Percentage    float64       `json:"percentage"`
	ShowFrom      string        `json:"show_from,omitempty"`
	DenomFuncName string        `json:"denom_func_name,omitempty"`
}

func mallocgcPercent(profilePath, showFrom, denomFunc string) (Result, error) {
	_, funcStats, err := analyzer.LoadProfileDataWithFunctionStats(profilePath, showFrom)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading profile: %v\n", err)
		os.Exit(1)
	}

	_, mallocgcTime := analyzer.SumFuncTime(funcStats, func(name string) bool { return name == "runtime.mallocgc" })

	var denominator time.Duration

	if denomFunc != "" {
		if stat, exists := funcStats[denomFunc]; exists {
			denominator = stat.Cum
		} else {
			if showFrom != "" {
				return Result{}, fmt.Errorf("Error: denominator function '%s' not found in profile when filtered by show_from '%s'\n", denomFunc, showFrom)
			} else {
				return Result{}, fmt.Errorf("Error: denominator function '%s' not found in profile\n", denomFunc)
			}
		}
	} else {
		denominator, err = analyzer.GetTotalProfileTime(*inputProfile)
		if err != nil {
			return Result{}, fmt.Errorf("Error getting total profile time: %v\n", err)
		}
	}

	if denominator == 0 {
		return Result{}, fmt.Errorf("Error: denominator is zero, cannot calculate percentage")
	}

	percentage := float64(mallocgcTime) / float64(denominator) * 100
	return Result{
		MallocgcTime:  mallocgcTime,
		Denominator:   denominator,
		DenomFuncName: denomFunc,
		ShowFrom:      showFrom,
		Percentage:    percentage,
	}, nil
}

func main() {
	flag.Parse()
	if err := validateFlags(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	result, err := mallocgcPercent(*inputProfile, *showFrom, *denomFunc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *format == "json" {
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	} else {
		fmt.Println("MallocGC Time Percentage Analysis")
		fmt.Println("==================================")
		if *showFrom != "" {
			fmt.Printf("Show From: %s\n", *showFrom)
		}
		if result.DenomFuncName != "" {
			fmt.Printf("Denominator Function: %s\n", result.DenomFuncName)
		} else {
			fmt.Println("Denominator: Total Profile Time")
		}
		fmt.Printf("MallocGC Time: %s\n", result.MallocgcTime)
		fmt.Printf("Denominator:   %s\n", result.Denominator)
		fmt.Printf("Percentage:    %f%%\n", result.Percentage)
	}
}
