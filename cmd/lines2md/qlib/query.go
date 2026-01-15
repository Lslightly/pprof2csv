package qlib

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Lslightly/pprof2csv/common"
	"github.com/Lslightly/pprof2csv/models"
)

// QueryResult represents query results with code information
type QueryResult struct {
	Filename   string
	LineNumber int
	Code       string
	Cum        time.Duration // Cumulative time
	Flat       time.Duration // Flat time (time spent directly in this function)
}

// QuerySection represents a section in the query file
type QuerySection struct {
	FunctionName string
	Queries      []Query
}

// Query represents a single line query with file:line and code comment
type Query struct {
	models.SourceLine
	Code string
}

// createQuerySection creates a QuerySection from function name and query lines
func CreateQuerySection(funcName string, lines []string) QuerySection {
	querySection := QuerySection{
		FunctionName: funcName,
		Queries:      []Query{},
	}

	for _, queryLine := range lines {
		parts := strings.SplitN(queryLine, ",", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "Warning: invalid query format: %s\n", queryLine)
			continue
		}

		fileLine := parts[0]
		code := parts[1]

		// Extract file and line number
		re := regexp.MustCompile(`(.+):(\d+)`)
		matches := re.FindStringSubmatch(fileLine)
		if len(matches) != 3 {
			fmt.Fprintf(os.Stderr, "Warning: invalid file:line format: %s\n", fileLine)
			continue
		}

		filePath := matches[1]
		lineNum := common.ParseInt(matches[2])

		querySection.Queries = append(querySection.Queries, Query{
			SourceLine: models.SourceLine{
				Filename:     filePath,
				LineNumber:   lineNum,
				FunctionName: funcName,
			},
			Code: code,
		})
	}

	return querySection
}
func ParseQueryFile(filename string) ([]QuerySection, error) {
	// Read query file
	queryFile := common.OpenFile(filename)

	// Process queries by sections (separated by empty lines)
	scanner := bufio.NewScanner(queryFile)
	var currentSection []string
	sections := [][]string{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			if len(currentSection) > 0 {
				sections = append(sections, currentSection)
				currentSection = []string{}
			}
		} else {
			currentSection = append(currentSection, line)
		}
	}
	// Add the last section if it exists
	if len(currentSection) > 0 {
		sections = append(sections, currentSection)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading query file: %v", err)
	}

	// Convert to QuerySection objects
	var result []QuerySection
	for _, section := range sections {
		if len(section) > 0 {
			querySection := CreateQuerySection(section[0], section[1:])
			result = append(result, querySection)
		}
	}

	return result, nil
}

func findMatchingLines(allLines []*models.SourceLine, query Query) []*models.SourceLine {
	var matchedLines []*models.SourceLine
	for _, sl := range allLines {
		// Check if filename ends with the queried path (supporting suffix matching)
		// And check if function name matches and line number matches
		if strings.HasSuffix(sl.Filename, query.Filename) &&
			sl.FunctionName == query.FunctionName &&
			sl.LineNumber == query.LineNumber {
			matchedLines = append(matchedLines, sl)
		}
	}

	// If no matches found, try to find any function with this name at this location
	if len(matchedLines) == 0 {
		for _, sl := range allLines {
			if strings.HasSuffix(sl.Filename, query.Filename) && sl.LineNumber == query.LineNumber {
				matchedLines = append(matchedLines, sl)
				fmt.Fprintf(os.Stderr, "Warning: using function %s instead of %s at %s:%d\n",
					sl.FunctionName, query.FunctionName, query.Filename, query.LineNumber)
			}
		}
	}

	return matchedLines
}

func GenerateMarkdownContent(querySections []QuerySection, matchedResults map[string]*models.SourceLine) string {
	var markdownContent strings.Builder

	for _, section := range querySections {
		if len(section.Queries) == 0 {
			continue
		}

		var table strings.Builder
		table.WriteString("| line | code | flat | cum |\n|---|---|---|---|\n")

		for _, query := range section.Queries {
			key := fmt.Sprintf("%s:%d", query.Filename, query.LineNumber)
			item := matchedResults[key]

			cumStr := common.FormatDuration(item.Cum)
			flatStr := common.FormatDuration(item.Flat)

			fmt.Fprintf(&table, "| %s:%d | %s | %s | %s |\n", query.Filename, query.LineNumber, query.Code, flatStr, cumStr)
		}

		fmt.Fprintf(&markdownContent, "## %s\n\n%s\n", section.FunctionName, table.String())
	}

	return markdownContent.String()
}

func (querySection QuerySection) WriteFunctionCSV(outputDir string, matchedResults map[string]*models.SourceLine) error {
	csvFilename := filepath.Join(outputDir, fmt.Sprintf("%s.csv", querySection.FunctionName))
	csvFile, err := os.Create(csvFilename)
	if err != nil {
		return fmt.Errorf("error creating CSV file: %v", err)
	}
	defer csvFile.Close()

	// Create QueryResult objects for export
	var queryResults []*QueryResult
	for _, query := range querySection.Queries {
		key := fmt.Sprintf("%s:%d", query.Filename, query.LineNumber)
		if resultLine, exists := matchedResults[key]; exists {
			queryResults = append(queryResults, &QueryResult{
				Filename:   query.Filename,
				LineNumber: query.LineNumber,
				Code:       query.Code,
				Cum:        resultLine.Cum,
				Flat:       resultLine.Flat,
			})
		}
	}

	err = exportQueryResults(csvFile, queryResults)
	if err != nil {
		return fmt.Errorf("error writing CSV file: %v", err)
	}

	return nil
}

// exportQueryResults writes QueryResult data to CSV with header: file,line,code,cum,flat
func exportQueryResults(w io.Writer, results []*QueryResult) error {
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header
	header := []string{"file", "line", "code", "cum", "flat"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, result := range results {
		cumulativeTimeStr := common.FormatDuration(result.Cum)
		flatTimeStr := common.FormatDuration(result.Flat)

		record := []string{
			result.Filename,
			fmt.Sprintf("%d", result.LineNumber),
			result.Code,
			cumulativeTimeStr,
			flatTimeStr,
		}

		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record for %s:%d: %w", result.Filename, result.LineNumber, err)
		}
	}

	// Check for any errors during writing
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("error flushing CSV data: %w", err)
	}

	return nil
}

func MatchQueries(querySections []QuerySection, allLines []*models.SourceLine) (matchedResults map[string]*models.SourceLine) {
	matchedResults = make(map[string]*models.SourceLine)
	for _, section := range querySections {
		for _, query := range section.Queries {
			matchedLines := findMatchingLines(allLines, query)

			// Create a new SourceLine to accumulate results
			resultLine := &models.SourceLine{
				Filename:     query.Filename,
				LineNumber:   query.LineNumber,
				FunctionName: query.FunctionName,
				Cum:          0,
				Flat:         0,
			}

			// Sum up all matching lines
			for _, sl := range matchedLines {
				resultLine.Cum += sl.Cum
				resultLine.Flat += sl.Flat
			}

			key := fmt.Sprintf("%s:%d", query.Filename, query.LineNumber)
			matchedResults[key] = resultLine
		}
	}
	return
}
