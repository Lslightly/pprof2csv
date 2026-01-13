package loader

import (
	"fmt"
	"os"
)

// ProfileLoader handles loading pprof profile data from files
type ProfileLoader struct{}

// New creates a new ProfileLoader instance
func New() *ProfileLoader {
	return &ProfileLoader{}
}

// Load loads a pprof profile from the given file path
func (l *ProfileLoader) Load(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file %s: %w", filePath, err)
	}
	return data, nil
}
