package test

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

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
