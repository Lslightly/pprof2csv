package common

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

func ParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Panicf("error parsing time %s: %v", s, err)
	}
	return d
}

func ParseInt(s string) (n int) {
	n, err := strconv.Atoi(s)
	if err != nil {
		log.Panicf("error parsing %s: %v", s, err)
	}
	return
}

func OpenFile(p string) (res *os.File) {
	res, err := os.Open(p)
	if err != nil {
		log.Panicf("error open file %s: %v", p, err)
	}
	return
}

// FormatDuration formats a duration for display in tables.
// If unit is empty, it returns the default Go duration string (e.g., "1m30s").
// If unit is specified, it formats in that unit (e.g., "s" -> "90.00s", "ms" -> "90000.00ms").
func FormatDuration(d time.Duration, unit string) string {
	if d == 0 {
		return "0ns"
	}
	if unit == "" {
		return d.String()
	}
	switch unit {
	case "s":
		return fmt.Sprintf("%.2f", d.Seconds())
	case "ms":
		return fmt.Sprintf("%d", d.Milliseconds())
	case "us", "µs":
		return fmt.Sprintf("%d", d.Microseconds())
	case "ns":
		return fmt.Sprintf("%d", d.Nanoseconds())
	default:
		return d.String()
	}
}

/*
callerDir return the dir of caller of callerDir

skip is the number of stack frame to skip, with 0 identifying the caller of callerDir
*/
func callerDir(skip int) string {
	_, file, _, _ := runtime.Caller(skip + 1)
	return filepath.Dir(file)
}

func RootDir() string {
	return filepath.Dir(callerDir(0))
}

func CurFileDir() string {
	return callerDir(1)
}

func AbsPathFromRoot(relativePath string) string {
	return filepath.Join(RootDir(), relativePath)
}

func AbsPathFromCur(relativePath string) string {
	return filepath.Join(callerDir(2), relativePath)
}

func Runcmd(cwd, name string, args ...string) (cmd *exec.Cmd, err error) {
	cmd = exec.Command(name, args...)
	cmd.Dir = cwd
	return cmd, cmd.Run()
}

func RuncmdCheck(cwd, name string, args ...string) {
	cmd, err := Runcmd(cwd, name, args...)
	if err != nil {
		log.Panicf("cmd %s run error(return code %d): %v", cmd.String(), cmd.ProcessState.ExitCode(), err)
	}
}
