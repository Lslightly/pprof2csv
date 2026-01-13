package internal

import "time"

// SourceLine represents timing information for a specific source line
type SourceLine struct {
	Filename     string
	LineNumber   int
	FunctionName string
	Cum          time.Duration // Cumulative time
	Flat         time.Duration // Flat time (time spent directly in this function)
}
