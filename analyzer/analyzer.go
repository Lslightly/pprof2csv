package analyzer

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/Lslightly/pprof2csv/models"
	"github.com/google/pprof/profile"
)

func convTimeUnit(s string) time.Duration {
	switch s {
	case "nanoseconds":
		return time.Nanosecond
	default:
		fmt.Printf("unknown time unit %s\n", s)
		return time.Nanosecond
	}
}

// AnalyzeWithFunctionStats parses the pprof profile data and extracts both
// source line timing information and per-function aggregated timing
// (flat: self time, cum: self + callees).
// If showFrom is non-empty, only samples whose stacktrace contains the specified
// function are included in the analysis.
// It returns:
//   - lines: per-source-line stats sorted by cumulative time descending
//   - funcStats: map keyed by function name with flat/cum times.
func AnalyzeWithFunctionStats(data []byte, showFrom string) ([]*models.SourceLine, map[string]*models.FunctionStat, error) {
	p, err := profile.ParseData(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse profile data: %w", err)
	}

	timeUnit := convTimeUnit(p.SampleType[1].Unit)

	// Create maps to aggregate time by source line and by function
	lineMap := make(map[string]*models.SourceLine)
	funcMap := make(map[string]*models.FunctionStat)

	// Process each sample in the profile
	for _, sample := range p.Sample {
		// Filter: skip sample if showFrom specified but not found in stacktrace
		if showFrom != "" && !findShowFrom(sample.Location, showFrom) {
			continue
		}

		// Get the value (time) for this sample
		var value int64
		if len(sample.Value) > 0 {
			value = sample.Value[1]
		}

		// Process each location in the stack trace
		for i, loc := range sample.Location {
			// Skip locations without lines
			if len(loc.Line) == 0 {
				continue
			}

			// Process all lines in the location, as a location may map to multiple source lines
			for _, lineEntry := range loc.Line {
				line := lineEntry

				// Skip if no filename or function name
				if line.Function == nil || line.Function.Filename == "" || line.Function.Name == "" {
					continue
				}

				// Create a unique key for this source line
				key := fmt.Sprintf("%s:%d:%s", line.Function.Filename, line.Line, line.Function.Name)

				// Flat time is the time spent directly in this function (leaf node in call stack)
				// Only the innermost function call (i == 0) gets the full sample value as flat time
				flatTime := int64(0)
				if i == 0 {
					flatTime = value
				}

				cumDelta := time.Duration(value) * timeUnit
				flatDelta := time.Duration(flatTime) * timeUnit

				// Update or create entry for this source line
				if sl, exists := lineMap[key]; exists {
					sl.Cum += cumDelta
					sl.Flat += flatDelta
				} else {
					lineMap[key] = &models.SourceLine{
						Filename:     line.Function.Filename,
						LineNumber:   int(line.Line),
						FunctionName: line.Function.Name,
						Cum:          cumDelta,
						Flat:         flatDelta,
					}
				}

				// Update or create entry for this function (function-level stats)
				fn := line.Function.Name
				if fs, exists := funcMap[fn]; exists {
					fs.Cum += cumDelta
					fs.Flat += flatDelta
				} else {
					funcMap[fn] = &models.FunctionStat{
						FunctionName: fn,
						Cum:          cumDelta,
						Flat:         flatDelta,
					}
				}
			}
		}
	}

	// Convert map to sorted slice for line-level stats
	result := make([]*models.SourceLine, 0, len(lineMap))
	for _, line := range lineMap {
		result = append(result, line)
	}

	// Sort by cumulative time (descending)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Cum > result[j].Cum
	})

	return result, funcMap, nil
}

// Analyze parses the pprof profile data and extracts source line timing information.
// If showFrom is non-empty, only samples whose stacktrace contains the specified
// function are included in the analysis.
// It is kept for backward compatibility and discards function-level aggregation.
func Analyze(data []byte, showFrom string) ([]*models.SourceLine, error) {
	lines, _, err := AnalyzeWithFunctionStats(data, showFrom)
	return lines, err
}

// LoadProfileDataWithFunctionStats loads profile data from the specified file
// and returns both per-line and per-function statistics.
// If showFrom is non-empty, only samples whose stacktrace contains the specified
// function are included in the analysis.
func LoadProfileDataWithFunctionStats(filename string, showFrom string) ([]*models.SourceLine, map[string]*models.FunctionStat, error) {
	// Load profile data
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("error loading profile: %v", err)
	}

	// Analyze profile data
	allLines, funcStats, err := AnalyzeWithFunctionStats(data, showFrom)
	if err != nil {
		return nil, nil, fmt.Errorf("error analyzing profile: %v", err)
	}

	return allLines, funcStats, nil
}

// LoadProfileData loads profile data and only returns per-source-line statistics.
func LoadProfileData(filename string, showFrom string) ([]*models.SourceLine, error) {
	allLines, _, err := LoadProfileDataWithFunctionStats(filename, showFrom)
	return allLines, err
}

// GetTotalProfileTime calculates the total time by summing all sample values in the profile.
// This returns the actual total profile time regardless of any filtering.
func GetTotalProfileTime(filename string) (time.Duration, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, fmt.Errorf("error loading profile: %v", err)
	}

	p, err := profile.ParseData(data)
	if err != nil {
		return 0, fmt.Errorf("failed to parse profile data: %w", err)
	}

	timeUnit := convTimeUnit(p.SampleType[1].Unit)
	var total int64

	for _, sample := range p.Sample {
		if len(sample.Value) > 1 {
			total += sample.Value[1]
		}
	}

	return time.Duration(total) * timeUnit, nil
}
