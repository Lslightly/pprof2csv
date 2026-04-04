package lib

import (
	"path/filepath"
	"testing"

	"github.com/Lslightly/pprof2csv/common"
	"github.com/stretchr/testify/assert"
)

func TestMallocgcPercent(t *testing.T) {
	res, err := MallocgcPercent(filepath.Join(common.RootDir(), "test/go_parser/default.out"), "go/parser.BenchmarkParseOnly", "go/parser.BenchmarkParseOnly")
	assert.Nil(t, err)
	assert.Equal(t, 43.28628302569671, res.Percentage)
}
