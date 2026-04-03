package analyzer

import (
	"time"

	"github.com/Lslightly/pprof2csv/models"
)

func SumFuncTime(funcStats map[string]*models.FunctionStat, satisfy func(name string) bool) (flat time.Duration, cum time.Duration) {
	for name, stat := range funcStats {
		if satisfy(name) {
			flat += stat.Flat
			cum += stat.Cum
		}
	}
	return
}
