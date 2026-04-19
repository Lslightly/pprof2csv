# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development

```bash
# Build main tool
go build -o pprof2csv .

# Build lines2md tool
go build -o lines2md ./cmd/lines2md/

# Build mallocgc_percent tool
go build -o mallocgc_percent ./cmd/mallocgc_percent/

# Build schema_export tool (relational CSV for CodeQL)
go build -o schema_export ./cmd/schema_export/

# Run tests
go test ./...

# Run specific test
go test ./analyzer/...
go test ./cmd/mallocgc_percent/lib/...
```

## Architecture

This is a Go toolset for converting pprof profiles (CPU/memory profiles) to CSV/Markdown for analysis. The codebase has four main CLI tools:

1. **pprof2csv** (main.go) - Basic profile to CSV conversion
2. **lines2md** (cmd/lines2md/) - Query specific source lines and generate CSV/markdown with timing
3. **mallocgc_percent** (cmd/mallocgc_percent/) - Calculate mallocgc time percentage
4. **schema_export** (cmd/schema_export/) - Relational CSV for CodeQL external data

### Core Packages

- **analyzer/** - Uses `github.com/google/pprof/profile` to parse profiles and
  extract timing data. Key functions:
  - `AnalyzeWithFunctionStats()` - Returns both per-line and per-function stats
  - `GetCallerKNameSet()` / `GetCalleeKNameSet()` - K-hop caller/callee analysis
  - `GetTotalProfileTime()` - Calculate total profile time

- **models/** - Data structures:
  - `SourceLine` - Per-source-line timing (flat/cum)
  - `FunctionStat` - Per-function aggregated timing

- **imexporter/** - CSV export/import functionality

- **common/** - Shared utilities for file I/O, time formatting, parsing

- **cmd/lines2md/qlib/** - Query file parsing, matching lines from profiles,
  generating markdown/CSV output

## Key Features

- **show_from flag**: Filter to only samples whose stacktrace contains specified function
- **Function statistics**: Both flat (self time) and cumulative (self + callees) timing
- **K-hop analysis**: Find caller/callees at specified depth in call graph
- **Multiple output formats**: CSV, Markdown, JSON

## Usage

```bash
# Basic conversion
./pprof2csv -i profile.pprof -o output.csv

# With filtering
./pprof2csv -i profile.pprof -show_from "myFunction" -o filtered.csv

# lines2md with query file
./lines2md -i profile.pprof -q queries.txt -dir results/

# mallocgc percentage
./mallocgc_percent -i profile.pprof -show_from "myFunction"

# schema_export: relational CSV for CodeQL
./schema_export -i profile.pprof -o output_dir/
```

Generates 5 CSV files:
- `samples.csv` - call stack samples (expanded)
- `locations.csv` - code locations
- `functions.csv` - function info
- `mappings.csv` - binary mappings
- `profile_info.csv` - profile metadata

## Dependencies

- `github.com/google/pprof` - Profile parsing library
- Go 1.24.2 required
