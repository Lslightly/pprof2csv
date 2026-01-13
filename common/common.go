package common

import (
	"log"
	"os"
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
