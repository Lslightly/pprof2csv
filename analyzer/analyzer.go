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
		if showFrom != "" {
			found := false
		locationLoop:
			for _, loc := range sample.Location {
				for _, le := range loc.Line {
					if le.Function != nil && le.Function.Name == showFrom {
						found = true
						break locationLoop
					}
				}
			}
			if !found {
				continue
			}
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

// GetCallerKNameSet retrieves the set of k-hop caller function names for a given callee.
//
// Parameters:
//   - filename: path to the pprof profile file
//   - callee: the target function name to find in the call stack
//   - k: the number of hops up the call stack to find the caller (k=1 for direct caller)
//
// Returns:
//   - []string: unique set of k-hop caller function names
//   - error: an error if the callee is not found in the profile or if the profile cannot be parsed
//
// The function searches through all sample call stacks in the profile.
// For each sample, it finds the callee in the call stack and then returns the function
// name that is k positions above it in the stack. The result is deduplicated before returning.
//
// Example:
//
//	If the call stack is: [main] -> [foo] -> [bar] -> [baz] (where baz is the leaf)
//	- GetCallerKNameSet("profile.pprof", "baz", 1) returns ["bar"]
//	- GetCallerKNameSet("profile.pprof", "baz", 2) returns ["foo"]
//	- GetCallerKNameSet("profile.pprof", "baz", 3) returns ["main"]
func GetCallerKNameSet(filename string, callee string, k int) (result []string, err error) {
	// Load and parse profile data
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading profile: %v", err)
	}

	p, err := profile.ParseData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile data: %w", err)
	}

	callerSet := make(map[string]struct{})
	calleeFound := false

	// Process each sample's call stack
	for _, sample := range p.Sample {
		// Search for callee in the call stack
		for calleeIdx, loc := range sample.Location {
			// Skip locations without lines
			if len(loc.Line) == 0 {
				continue
			}

			// Check if any line in this location matches the callee
			isCallee := false
			for _, lineEntry := range loc.Line {
				if lineEntry.Function != nil && lineEntry.Function.Name == callee {
					isCallee = true
					break
				}
			}

			if isCallee {
				calleeFound = true
				// Calculate the index of the k-hop caller
				callerIdx := calleeIdx + k

				// Check if caller index is within valid range
				if callerIdx >= 0 && callerIdx < len(sample.Location) {
					callerLoc := sample.Location[callerIdx]

					// Get function name from caller location
					for _, lineEntry := range callerLoc.Line {
						if lineEntry.Function != nil && lineEntry.Function.Name != "" {
							callerSet[lineEntry.Function.Name] = struct{}{}
							break // Only need one function name per location
						}
					}
				}
				break // Found callee in this sample, move to next sample
			}
		}
	}

	// Return error if callee was never found
	if !calleeFound {
		return nil, fmt.Errorf("callee function '%s' not found in the profile", callee)
	}

	// Convert map to slice
	result = make([]string, 0, len(callerSet))
	for name := range callerSet {
		result = append(result, name)
	}

	return result, nil
}

// GetCalleeKNameSet retrieves the set of k-hop callee function names for a given caller.
//
// Parameters:
//   - filename: path to the pprof profile file
//   - caller: the target function name to find in the call stack
//   - k: the number of hops down the call stack to find the callee (k=1 for direct callee)
//
// Returns:
//   - []string: unique set of k-hop callee function names
//   - error: an error if the caller is not found in the profile or if the profile cannot be parsed
//
// The function searches through all sample call stacks in the profile.
// For each sample, it finds the caller in the call stack and then returns the function
// name that is k positions below it in the stack. The result is deduplicated before returning.
//
// Example:
//
//	If the call stack is: [main] -> [foo] -> [bar] -> [baz] (where baz is the leaf)
//	- GetCalleeKNameSet("profile.pprof", "main", 1) returns ["foo"]
//	- GetCalleeKNameSet("profile.pprof", "main", 2) returns ["bar"]
//	- GetCalleeKNameSet("profile.pprof", "main", 3) returns ["baz"]
func GetCalleeKNameSet(filename string, caller string, k int) (result []string, err error) {
	// Load and parse profile data
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading profile: %v", err)
	}

	p, err := profile.ParseData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile data: %w", err)
	}

	calleeSet := make(map[string]struct{})
	callerFound := false

	// Process each sample's call stack
	for _, sample := range p.Sample {
		// Search for caller in the call stack
		for callerIdx, loc := range sample.Location {
			// Skip locations without lines
			if len(loc.Line) == 0 {
				continue
			}

			// Check if any line in this location matches the caller
			isCaller := false
			for _, lineEntry := range loc.Line {
				if lineEntry.Function != nil && lineEntry.Function.Name == caller {
					isCaller = true
					break
				}
			}

			if isCaller {
				callerFound = true
				// Calculate the index of the k-hop callee
				calleeIdx := callerIdx - k

				// Check if callee index is within valid range
				if calleeIdx >= 0 && calleeIdx < len(sample.Location) {
					calleeLoc := sample.Location[calleeIdx]

					// Get function name from callee location
					for _, lineEntry := range calleeLoc.Line {
						if lineEntry.Function != nil && lineEntry.Function.Name != "" {
							calleeSet[lineEntry.Function.Name] = struct{}{}
							break // Only need one function name per location
						}
					}
				}
				break // Found caller in this sample, move to next sample
			}
		}
	}

	// Return error if caller was never found
	if !callerFound {
		return nil, fmt.Errorf("caller function '%s' not found in the profile", caller)
	}

	// Convert map to slice
	result = make([]string, 0, len(calleeSet))
	for name := range calleeSet {
		result = append(result, name)
	}

	return result, nil
}
