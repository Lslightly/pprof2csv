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
	allLines, err := analyzer.LoadProfileData(cpuProfPath, "")
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
	const spanFreeIndexForScanUpdateLine = "src/runtime/malloc.go:1399"
	const mallocgcSmallNoscanLine = "src/runtime/malloc.go:1308"

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
				spanFreeIndexForScanUpdateLine: {
					"1.01s", "1.01s",
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
				spanFreeIndexForScanUpdateLine: {
					"1.24s", "1.24s",
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
				spanFreeIndexForScanUpdateLine: {
					"970ms", "970ms",
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
				spanFreeIndexForScanUpdateLine: {
					"840ms", "840ms",
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
				spanFreeIndexForScanUpdateLine: {
					"880ms", "880ms",
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
				spanFreeIndexForScanUpdateLine: {
					"2.86s", "2.86s",
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
				spanFreeIndexForScanUpdateLine: {
					"3.27s", "3.27s",
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
				spanFreeIndexForScanUpdateLine: {
					"3.03s", "3.03s",
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
				spanFreeIndexForScanUpdateLine: {
					"710ms", "710ms",
				},
				mallocgcSmallNoscanLine: {
					"240ms", "240ms",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Load profile data using analyzer API
			cpuProfPath := common.AbsPathFromRoot(tc.profilePath)
			allLines, err := analyzer.LoadProfileData(cpuProfPath, "")
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
						assert.NotNil(t, resultLine)
						assert.Equal(t, common.ParseDuration(expected.flat), resultLine.Flat, "Flat time mismatch for %s", key)
						assert.Equal(t, common.ParseDuration(expected.cum), resultLine.Cum, "Cum time mismatch for %s", key)
					}
				}
			}
		})
	}
}

func TestShowFrom(t *testing.T) {
	// Parse query file using qlib API (shared across all tests)
	queryPath := common.AbsPathFromRoot("test/protoactor-go/default.txt")
	querySections, err := ParseQueryFile(queryPath)
	if err != nil {
		t.Fatalf("Failed to parse query file %v", err)
	}
	testCases := []struct {
		name            string
		profilePath     string
		showFrom        string
		expectedResults map[string]struct {
			flat string
			cum  string
		}
	}{
		{
			name:        "BenchmarkPIDSet_Add_cpu",
			profilePath: "test/protoactor-go/BenchmarkPIDSet_Add/cpu-100-default.out",
			showFrom:    "runtime.gcDrainMarkWorkerIdle",
			expectedResults: map[string]struct {
				flat string
				cum  string
			}{
				"src/runtime/mgcmark.go:1213": { // b := gcw.tryGetFast()
					"13.17s", "13.17s",
				},
			},
		},
		{
			name:        "BenchmarkPIDSet_Add_cpu",
			profilePath: "test/protoactor-go/BenchmarkPIDSet_Add/cpu-100-default.out",
			showFrom:    "runtime.gcDrainMarkWorkerIdle",
			expectedResults: map[string]struct {
				flat string
				cum  string
			}{
				"src/runtime/mgcmark.go:1228": { // scanobject(b, gcw)
					"10ms", "14.44s",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Load profile data using analyzer API
			cpuProfPath := common.AbsPathFromRoot(tc.profilePath)
			allLines, err := analyzer.LoadProfileData(cpuProfPath, tc.showFrom)
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
						assert.NotNil(t, resultLine)
						assert.Equal(t, common.ParseDuration(expected.flat), resultLine.Flat, "Flat time mismatch for %s", key)
						assert.Equal(t, common.ParseDuration(expected.cum), resultLine.Cum, "Cum time mismatch for %s", key)
					}
				}
			}
		})
	}
}
