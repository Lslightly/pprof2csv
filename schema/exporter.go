package schema

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/pprof/profile"
)

// Exporter exports profile data to relational CSV format
type Exporter struct {
	outputDir string
}

// NewExporter creates a new Exporter
func NewExporter(outputDir string) *Exporter {
	return &Exporter{outputDir: outputDir}
}

// Export exports the profile to all CSV tables
func (e *Exporter) Export(p *profile.Profile) error {
	// Generate CSV data
	samples, err := e.generateSamples(p)
	if err != nil {
		return fmt.Errorf("failed to generate samples: %w", err)
	}

	locations := e.generateLocations(p)
	functions := e.generateFunctions(p)
	mappings := e.generateMappings(p)
	profileInfo := e.generateProfileInfo(p)

	// Write CSV files
	if err := e.writeCSV("samples.csv", samples, []string{
		"sample_id", "stack_depth", "location_id", "depth", "count", "cpu_nanos",
	}); err != nil {
		return err
	}

	if err := e.writeCSV("locations.csv", locations, []string{
		"location_id", "mapping_id", "address", "function_id", "filename",
		"line_number", "column_number", "is_folded",
	}); err != nil {
		return err
	}

	if err := e.writeCSV("functions.csv", functions, []string{
		"function_id", "name", "system_name", "start_line",
	}); err != nil {
		return err
	}

	if err := e.writeCSV("mappings.csv", mappings, []string{
		"mapping_id", "memory_start", "memory_limit", "filename", "build_id",
		"has_functions", "has_filenames",
	}); err != nil {
		return err
	}

	return e.writeCSV("profile_info.csv", profileInfo, []string{"key", "value"})
}

// generateSamples expands each sample into multiple rows (one per stack depth)
func (e *Exporter) generateSamples(p *profile.Profile) ([][]string, error) {
	result := [][]string{}
	sampleID := 0

	for _, sample := range p.Sample {
		stackDepth := len(sample.Location)

		// Get count and cpu time from sample values
		count := int64(0)
		cpuNanos := int64(0)
		if len(sample.Value) > 0 {
			count = sample.Value[0]
		}
		if len(sample.Value) > 1 {
			cpuNanos = sample.Value[1]
		}

		// Expand call stack: each location gets a row
		for depth, loc := range sample.Location {
			result = append(result, []string{
				strconv.Itoa(sampleID),
				strconv.Itoa(stackDepth),
				strconv.FormatUint(loc.ID, 10),
				strconv.Itoa(depth),
				strconv.FormatInt(count, 10),
				strconv.FormatInt(cpuNanos, 10),
			})
		}

		sampleID++
	}

	return result, nil
}

// generateLocations generates location rows, expanding inline functions
func (e *Exporter) generateLocations(p *profile.Profile) [][]string {
	result := [][]string{}

	for _, loc := range p.Location {
		mappingID := uint64(0)
		if loc.Mapping != nil {
			mappingID = loc.Mapping.ID
		}
		address := strconv.FormatUint(loc.Address, 16)
		isFolded := "false"
		if loc.IsFolded {
			isFolded = "true"
		}

		// If no lines, still create a row with empty line info
		if len(loc.Line) == 0 {
			result = append(result, []string{
				strconv.FormatUint(loc.ID, 10),
				strconv.FormatUint(mappingID, 10),
				address,
				"",
				"",
				"0",
				"0",
				isFolded,
			})
			continue
		}

		// Expand lines: each inline function gets a row
		for _, line := range loc.Line {
			functionID := uint64(0)
			filename := ""

			if line.Function != nil {
				functionID = line.Function.ID
				filename = line.Function.Filename
			}

			result = append(result, []string{
				strconv.FormatUint(loc.ID, 10),
				strconv.FormatUint(mappingID, 10),
				address,
				strconv.FormatUint(functionID, 10),
				filename,
				strconv.FormatInt(line.Line, 10),
				strconv.FormatInt(line.Column, 10),
				isFolded,
			})
		}
	}

	return result
}

// generateFunctions generates function rows
func (e *Exporter) generateFunctions(p *profile.Profile) [][]string {
	result := [][]string{}

	for _, fn := range p.Function {
		systemName := fn.SystemName
		if systemName == "" {
			systemName = fn.Name
		}

		result = append(result, []string{
			strconv.FormatUint(fn.ID, 10),
			fn.Name,
			systemName,
			strconv.FormatInt(fn.StartLine, 10),
		})
	}

	return result
}

// generateMappings generates mapping rows
func (e *Exporter) generateMappings(p *profile.Profile) [][]string {
	result := [][]string{}

	for _, m := range p.Mapping {
		hasFunctions := "false"
		hasFilenames := "false"
		if m.HasFunctions {
			hasFunctions = "true"
		}
		if m.HasFilenames {
			hasFilenames = "true"
		}

		result = append(result, []string{
			strconv.FormatUint(m.ID, 10),
			strconv.FormatUint(m.Start, 16),
			strconv.FormatUint(m.Limit, 16),
			m.File,
			m.BuildID,
			hasFunctions,
			hasFilenames,
		})
	}

	return result
}

// generateProfileInfo generates profile metadata rows
func (e *Exporter) generateProfileInfo(p *profile.Profile) [][]string {
	result := [][]string{}

	// Add time and duration
	if p.TimeNanos != 0 {
		result = append(result, []string{"time_nanos", strconv.FormatInt(p.TimeNanos, 10)})
	}
	if p.DurationNanos != 0 {
		result = append(result, []string{"duration_nanos", strconv.FormatInt(p.DurationNanos, 10)})
	}

	// Add period
	if p.Period != 0 {
		result = append(result, []string{"sampling_period", strconv.FormatInt(p.Period, 10)})
	}

	// Add sample types
	for i, st := range p.SampleType {
		result = append(result, []string{
			fmt.Sprintf("sample_type_%d_type", i),
			st.Type,
		})
		result = append(result, []string{
			fmt.Sprintf("sample_type_%d_unit", i),
			st.Unit,
		})
	}

	// Add default sample type
	if p.DefaultSampleType != "" {
		result = append(result, []string{
			"default_sample_type",
			p.DefaultSampleType,
		})
	}

	// Add period type if exists
	if p.PeriodType != nil {
		result = append(result, []string{"period_type_type", p.PeriodType.Type})
		result = append(result, []string{"period_type_unit", p.PeriodType.Unit})
	}

	return result
}

// writeCSV writes data to a CSV file
func (e *Exporter) writeCSV(filename string, data [][]string, header []string) error {
	path := filepath.Join(e.outputDir, filename)
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header to %s: %w", filename, err)
	}

	for _, row := range data {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row to %s: %w", filename, err)
		}
	}

	return nil
}

// ExportFromBytes exports profile data from bytes
func ExportFromBytes(data []byte, outputDir string) error {
	p, err := profile.ParseData(data)
	if err != nil {
		return fmt.Errorf("failed to parse profile: %w", err)
	}

	exporter := NewExporter(outputDir)
	return exporter.Export(p)
}
