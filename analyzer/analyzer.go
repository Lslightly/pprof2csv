package analyzer

import (
	"fmt"
	"sort"
	"time"

	"github.com/Lslightly/pprof2csv/internal"
	"github.com/google/pprof/profile"
)

// ProfileAnalyzer analyzes pprof profile data to extract source line timing information
type ProfileAnalyzer struct{}

// New creates a new ProfileAnalyzer instance
func New() *ProfileAnalyzer {
	return &ProfileAnalyzer{}
}

func convTimeUnit(s string) time.Duration {
	switch s {
	case "nanoseconds":
		return time.Nanosecond
	default:
		fmt.Printf("unknown time unit %s\n", s)
		return time.Nanosecond
	}
}

// Analyze parses the pprof profile data and extracts source line timing information
func (a *ProfileAnalyzer) Analyze(data []byte) ([]*internal.SourceLine, error) {
	p, err := profile.ParseData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile data: %w", err)
	}

	timeUnit := convTimeUnit(p.SampleType[1].Unit)

	// Create map to aggregate time by source line
	lineMap := make(map[string]*internal.SourceLine)
	// Process each sample in the profile
	for _, sample := range p.Sample {
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

				// Update or create entry for this source line
				if sl, exists := lineMap[key]; exists {
					sl.Cum += time.Duration(value) * timeUnit
					sl.Flat += time.Duration(flatTime) * timeUnit
				} else {
					lineMap[key] = &internal.SourceLine{
						Filename:     line.Function.Filename,
						LineNumber:   int(line.Line),
						FunctionName: line.Function.Name,
						Cum:          time.Duration(value) * timeUnit,
						Flat:         time.Duration(flatTime) * timeUnit,
					}
				}
			}
		}
	}

	// Convert map to sorted slice
	result := make([]*internal.SourceLine, 0, len(lineMap))
	for _, line := range lineMap {
		result = append(result, line)
	}

	// Sort by cumulative time (descending)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Cum > result[j].Cum
	})

	return result, nil
}
