package lib

import (
	"fmt"
	"os"
	"time"

	"github.com/Lslightly/pprof2csv/analyzer"
)

type Result struct {
	MallocgcTime  time.Duration `json:"mallocgc_time"`
	Denominator   time.Duration `json:"denominator"`
	Percentage    float64       `json:"percentage"`
	ShowFrom      string        `json:"show_from,omitempty"`
	DenomFuncName string        `json:"denom_func_name,omitempty"`
}

func MallocgcPercent(profilePath, showFrom, denomFunc string) (Result, error) {
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
		denominator, err = analyzer.GetTotalProfileTime(profilePath)
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
