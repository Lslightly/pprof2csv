package imexporter

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Lslightly/pprof2csv/common"
	"github.com/Lslightly/pprof2csv/models"
)

// CSVExporter converts source line timing data to CSV format
type CSVExporter struct{}

// New creates a new CSVExporter instance
func New() *CSVExporter {
	return &CSVExporter{}
}

// Export writes the source line data to a CSV writer
func (e *CSVExporter) Export(w io.Writer, lines []*models.SourceLine) error {
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header
	header := []string{"file", "line", "function", "flat", "cum"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, line := range lines {
		// Convert nanoseconds to human-readable time format (e.g., 1.234ms)
		cumulativeTimeStr := common.FormatDuration(time.Duration(line.Cum))
		flatTimeStr := common.FormatDuration(time.Duration(line.Flat))

		record := []string{
			line.Filename,
			fmt.Sprintf("%d", line.LineNumber),
			line.FunctionName,
			flatTimeStr,
			cumulativeTimeStr,
		}

		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record for %s:%d: %w", line.Filename, line.LineNumber, err)
		}
	}

	// Check for any errors during writing
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("error flushing CSV data: %w", err)
	}

	return nil
}

// buildSourceLine build SourceLine from record
func buildSourceLine(record []string) *models.SourceLine {
	return &models.SourceLine{
		Filename:     record[0],
		LineNumber:   common.ParseInt(record[1]),
		FunctionName: record[2],
		Cum:          common.ParseDuration(record[4]),
		Flat:         common.ParseDuration(record[3]),
	}
}

func Import(r io.Reader) (sls []*models.SourceLine) {
	csvReader := csv.NewReader(r)
	rs, err := csvReader.ReadAll()
	if err != nil {
		log.Panicf("error reading csv: %v", err)
	}
	for _, record := range rs[1:] { // ignore header
		sls = append(sls, buildSourceLine(record))
	}
	return
}
