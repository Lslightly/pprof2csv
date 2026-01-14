package qlib

import (
	"fmt"
	"testing"

	"github.com/Lslightly/pprof2csv/analyzer"
	"github.com/Lslightly/pprof2csv/common"
	"github.com/stretchr/testify/assert"
)

func TestMatchQueries(t *testing.T) {
	// Load profile data using analyzer API
	cpuProfPath := common.AbsPath("test/loop/cpu.pprof")
	allLines, err := analyzer.LoadProfileData(cpuProfPath)
	if err != nil {
		t.Fatalf("Failed to load and analyze profile: %v", err)
	}

	// Parse query file using qlib API
	queryPath := common.AbsPath("test/loop/query.txt")
	querySections, err := ParseQueryFile(queryPath)
	if err != nil {
		t.Fatalf("Failed to parse query file %v", err)
	}

	// Match queries using qlib API
	matchedResults := MatchQueries(querySections, allLines)

	// Test results based on test/loop/final_output.csv
	// Verify that the matched results match the expected values from the CSV file
	expectedResults := map[string]struct {
		flat string
		cum  string
	}{
		"test/loop/test.go:13": {flat: "510ms", cum: "510ms"},
		"test/loop/test.go:29": {flat: "4.19s", cum: "4.2s"},
		"test/loop/test.go:23": {flat: "0ns", cum: "30ms"},
		"test/loop/test.go:30": {flat: "250ms", cum: "250ms"},
		"test/loop/test.go:69": {flat: "0ns", cum: "6.05s"},
	}

	for _, section := range querySections {
		for _, query := range section.Queries {
			key := fmt.Sprintf("%s:%d", query.Filename, query.LineNumber)
			if expected, exists := expectedResults[key]; exists {
				resultLine := matchedResults[key]
				assert.Equal(t, common.ParseDuration(expected.flat), resultLine.Flat, "Flat time mismatch for %s", key)
				assert.Equal(t, common.ParseDuration(expected.cum), resultLine.Cum, "Cum time mismatch for %s", key)
			}
		}
	}
}
