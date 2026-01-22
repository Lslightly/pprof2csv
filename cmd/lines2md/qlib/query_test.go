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
	cpuProfPath := common.AbsPathFromRoot("test/loop/cpu.pprof")
	allLines, err := analyzer.LoadProfileData(cpuProfPath)
	if err != nil {
		t.Fatalf("Failed to load and analyze profile: %v", err)
	}

	// Parse query file using qlib API
	queryPath := common.AbsPathFromRoot("test/loop/query.txt")
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

func TestMatchQueriesProtoActorGo(t *testing.T) {
	// Parse query file using qlib API (shared across all tests)
	queryPath := common.AbsPathFromRoot("test/protoactor-go/default.txt")
	querySections, err := ParseQueryFile(queryPath)
	if err != nil {
		t.Fatalf("Failed to parse query file %v", err)
	}

	// Test cases for all .out files with their expected results
	testCases := []struct {
		name            string
		profilePath     string
		expectedResults map[string]struct {
			flat string
			cum  string
		}
	}{
		{
			name:        "BenchmarkPIDSet_Add_cpu-100-default",
			profilePath: "test/protoactor-go/BenchmarkPIDSet_Add/cpu-100-default.out",
			expectedResults: map[string]struct {
				flat string
				cum  string
			}{
				// TODO: Fill in expected flat and cum times
				"src/runtime/malloc.go:1399": {
					"1.01s", "1.01s",
				},
				"src/runtime/malloc.go:1308": {
					"0ns", "0ns",
				},
			},
		},
		{
			name:        "BenchmarkPIDSet_Add_cpu-500-default",
			profilePath: "test/protoactor-go/BenchmarkPIDSet_Add/cpu-500-default.out",
			expectedResults: map[string]struct {
				flat string
				cum  string
			}{
				// TODO: Fill in expected flat and cum times
				"src/runtime/malloc.go:1399": {
					"1.24s", "1.24s",
				},
				"src/runtime/malloc.go:1308": {
					"0ns", "0ns",
				},
			},
		},
		{
			name:        "BenchmarkPIDSet_Add_cpu-off-default",
			profilePath: "test/protoactor-go/BenchmarkPIDSet_Add/cpu-off-default.out",
			expectedResults: map[string]struct {
				flat string
				cum  string
			}{
				// TODO: Fill in expected flat and cum times
				"src/runtime/malloc.go:1399": {
					"970ms", "970ms",
				},
				"src/runtime/malloc.go:1308": {
					"0ns", "0ns",
				},
			},
		},
		{
			name:        "BenchmarkPIDSet_AddRemove_cpu-100-default",
			profilePath: "test/protoactor-go/BenchmarkPIDSet_AddRemove/cpu-100-default.out",
			expectedResults: map[string]struct {
				flat string
				cum  string
			}{
				// TODO: Fill in expected flat and cum times
				"src/runtime/malloc.go:1399": {
					"840ms", "840ms",
				},
				"src/runtime/malloc.go:1308": {
					"0ns", "0ns",
				},
			},
		},
		{
			name:        "BenchmarkPIDSet_AddRemove_cpu-500-default",
			profilePath: "test/protoactor-go/BenchmarkPIDSet_AddRemove/cpu-500-default.out",
			expectedResults: map[string]struct {
				flat string
				cum  string
			}{
				// TODO: Fill in expected flat and cum times
				"src/runtime/malloc.go:1399": {
					"880ms", "880ms",
				},
				"src/runtime/malloc.go:1308": {
					"0ns", "0ns",
				},
			},
		},
		{
			name:        "BenchmarkPushPop_cpu-100-default",
			profilePath: "test/protoactor-go/BenchmarkPushPop/cpu-100-default.out",
			expectedResults: map[string]struct {
				flat string
				cum  string
			}{
				// TODO: Fill in expected flat and cum times
				"src/runtime/malloc.go:1399": {
					"2.86s", "2.86s",
				},
				"src/runtime/malloc.go:1308": {
					"0ns", "0ns",
				},
			},
		},
		{
			name:        "BenchmarkPushPop_cpu-500-default",
			profilePath: "test/protoactor-go/BenchmarkPushPop/cpu-500-default.out",
			expectedResults: map[string]struct {
				flat string
				cum  string
			}{
				// TODO: Fill in expected flat and cum times
				"src/runtime/malloc.go:1399": {
					"3.27s", "3.27s",
				},
				"src/runtime/malloc.go:1308": {
					"0ns", "0ns",
				},
			},
		},
		{
			name:        "BenchmarkPushPop_cpu-off-default",
			profilePath: "test/protoactor-go/BenchmarkPushPop/cpu-off-default.out",
			expectedResults: map[string]struct {
				flat string
				cum  string
			}{
				// TODO: Fill in expected flat and cum times
				"src/runtime/malloc.go:1399": {
					"3.03s", "3.03s",
				},
				"src/runtime/malloc.go:1308": {
					"0ns", "0ns",
				},
			},
		},
		{
			name:        "Benchmark_Rendezvous_Get_cpu-100-default",
			profilePath: "test/protoactor-go/Benchmark_Rendezvous_Get/cpu-100-default.out",
			expectedResults: map[string]struct {
				flat string
				cum  string
			}{
				// TODO: Fill in expected flat and cum times
				"src/runtime/malloc.go:1399": {
					"710ms", "710ms",
				},
				"src/runtime/malloc.go:1308": {
					"240ms", "240ms",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Load profile data using analyzer API
			cpuProfPath := common.AbsPathFromRoot(tc.profilePath)
			allLines, err := analyzer.LoadProfileData(cpuProfPath)
			if err != nil {
				t.Fatalf("Failed to load and analyze profile: %v", err)
			}

			// Match queries using qlib API
			matchedResults := MatchQueries(querySections, allLines)

			// Verify expected results
			for _, section := range querySections {
				for _, query := range section.Queries {
					key := fmt.Sprintf("%s:%d", query.Filename, query.LineNumber)
					if expected, exists := tc.expectedResults[key]; exists {
						resultLine := matchedResults[key]
						assert.Equal(t, common.ParseDuration(expected.flat), resultLine.Flat, "Flat time mismatch for %s", key)
						assert.Equal(t, common.ParseDuration(expected.cum), resultLine.Cum, "Cum time mismatch for %s", key)
					}
				}
			}
		})
	}
}
