package test

import (
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Lslightly/pprof2csv/common"
	"github.com/Lslightly/pprof2csv/imexporter"
	"github.com/Lslightly/pprof2csv/models"
	"github.com/stretchr/testify/assert"
)

func runcmd(cwd, name string, args ...string) (cmd *exec.Cmd, err error) {
	cmd = exec.Command(name, args...)
	cmd.Dir = cwd
	return cmd, cmd.Run()
}

func runcmdCheck(cwd, name string, args ...string) {
	cmd, err := runcmd(cwd, name, args...)
	if err != nil {
		log.Panicf("cmd %s run error(return code %d): %v", cmd.String(), cmd.ProcessState.ExitCode(), err)
	}
}

func curdir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func rootdir() string {
	return filepath.Dir(curdir())
}

func assertCum(t *testing.T, cum time.Duration, sls []*models.SourceLine, fileEndPart string, line int) {
	for _, sl := range sls {
		if !strings.HasSuffix(sl.Filename, fileEndPart) {
			continue
		}
		if sl.LineNumber != line {
			continue
		}
		assert.Equal(t, cum, sl.Cum, "expected and got cum are not equal")
	}
}

var loopOnce sync.Once

func loopInit() (csvPath string) {
	loopDir := filepath.Join(curdir(), "loop")
	profPath := filepath.Join(loopDir, "cpu.pprof")
	csvPath = filepath.Join(loopDir, "final_output.csv")
	loopOnce.Do(func() {
		runcmdCheck(rootdir(), "go", "run", ".", "-i", profPath, "-o", csvPath)
	})
	return
}

func TestAnalyze(t *testing.T) {
	csvFile := common.OpenFile(loopInit())
	defer csvFile.Close()
	sls := imexporter.Import(csvFile)
	assertCum(t, common.ParseDuration("6.06s"), sls, "pprof2csv/testdata/test.go", 69)
}
