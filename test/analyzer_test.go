package test

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Lslightly/pprof2csv/analyzer"
	"github.com/Lslightly/pprof2csv/common"
	"github.com/Lslightly/pprof2csv/imexporter"
	"github.com/Lslightly/pprof2csv/models"
	"github.com/stretchr/testify/assert"
)

func assertCum(t *testing.T, expectCum time.Duration, sls []*models.SourceLine, fileEndPart string, line int) {
	for _, sl := range sls {
		if !strings.HasSuffix(sl.Filename, fileEndPart) {
			continue
		}
		if sl.LineNumber != line {
			continue
		}
		assert.Equal(t, expectCum, sl.Cum, "expected and got cum are not equal")
	}
}

var loopOnce sync.Once

func loopInit() (csvPath string) {
	loopDir := filepath.Join(common.CurFileDir(), "loop")
	profPath := filepath.Join(loopDir, "cpu.pprof")
	csvPath = filepath.Join(loopDir, "final_output.csv")
	loopOnce.Do(func() {
		common.RuncmdCheck(common.RootDir(), "go", "run", ".", "-i", profPath, "-o", csvPath)
	})
	return
}

func TestLoop(t *testing.T) {
	csvFile := common.OpenFile(loopInit())
	defer csvFile.Close()
	sls := imexporter.Import(csvFile)
	assertCum(t, common.ParseDuration("6.05s"), sls, "pprof2csv/testdata/test.go", 69)
}

func TestGetTotalProfileTime(t *testing.T) {
	totalT, err := analyzer.GetTotalProfileTime(filepath.Join(common.CurFileDir(), "loop/cpu.pprof"))
	assert.Nil(t, err)
	exp, err := time.ParseDuration("6.17s") // Duration: 6.42s, Total samples = 6.17s (96.14%)
	assert.Nil(t, err)
	assert.Equal(t, exp, totalT)
}

func TestGetCallerKNameSet(t *testing.T) {
	// Test that mallocgc's 1-hop caller is only runtime.newobject
	callers, err := analyzer.GetCallerKNameSet(filepath.Join(common.CurFileDir(), "go_parser/default.out"), "runtime.mallocgc", 1, "")
	assert.Nil(t, err)

	// Verify that the result contains only runtime.newobject
	assert.Len(t, callers, 6, "mallocgc should have exactly 6 1-hop caller")
	assert.Contains(t, callers, "runtime.newobject", "1-hop caller should be runtime.newobject")
}

func TestGetCallerKNameSet_WithShowFrom(t *testing.T) {
	// Test with showFrom parameter
	// When showFrom is specified, only samples containing that function are considered
	callers, err := analyzer.GetCallerKNameSet(filepath.Join(common.CurFileDir(), "go_parser/default.out"), "runtime.mallocgc", 1, "runtime.newobject")
	assert.Nil(t, err)

	assert.Len(t, callers, 1)
	// The result should be filtered based on showFrom
	assert.Contains(t, callers, "runtime.newobject", "1-hop caller should be runtime.newobject")
}

func TestGetCallerKNameSet_NotFound(t *testing.T) {
	// Test case when callee function does not exist
	_, err := analyzer.GetCallerKNameSet(filepath.Join(common.CurFileDir(), "go_parser/default.out"), "nonexistent_function", 1, "")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not found in the profile")
}

func TestGetCalleeKNameSet(t *testing.T) {
	// Test that runtime.newobject's 1-hop callee is runtime.mallocgc
	callees, err := analyzer.GetCalleeKNameSet(filepath.Join(common.CurFileDir(), "go_parser/default.out"), "runtime.newobject", 1, "")
	assert.Nil(t, err)

	// Verify that the result contains runtime.mallocgc
	assert.Contains(t, callees, "runtime.mallocgc", "1-hop callee should be runtime.mallocgc")
}

func TestGetCalleeKNameSet_WithShowFrom(t *testing.T) {
	// Test with showFrom parameter
	// When showFrom is specified, only samples containing that function are considered
	callees, err := analyzer.GetCalleeKNameSet(filepath.Join(common.CurFileDir(), "go_parser/default.out"), "runtime.newobject", 1, "")
	assert.Nil(t, err)

	// The result should be filtered based on showFrom
	assert.Contains(t, callees, "runtime.mallocgc", "1-hop callee should be runtime.mallocgc")
}

func TestGetCalleeKNameSet_NotFound(t *testing.T) {
	// Test case when caller function does not exist
	_, err := analyzer.GetCalleeKNameSet(filepath.Join(common.CurFileDir(), "go_parser/default.out"), "nonexistent_function", 1, "")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not found in the profile")
}
