package schema

// SampleRow represents a row in the samples.csv file
// Each call stack sample is expanded into multiple rows, one per stack depth
type SampleRow struct {
	SampleID    int    `csv:"sample_id"`     // Unique sample identifier
	StackDepth  int    `csv:"stack_depth"`   // Total depth of the call stack
	LocationID  int    `csv:"location_id"`   // Foreign key to locations table
	Depth       int    `csv:"depth"`         // Stack position (0 = leaf/stack top)
	Count       int    `csv:"count"`          // Sample count
	CpuNanos    int    `csv:"cpu_nanos"`      // CPU time in nanoseconds
}

// LocationRow represents a row in the locations.csv file
type LocationRow struct {
	LocationID   int    `csv:"location_id"`    // Unique location identifier (pprof Location.id)
	MappingID    int    `csv:"mapping_id"`     // Foreign key to mappings table
	Address      string `csv:"address"`        // Instruction address (hex string)
	FunctionID   int    `csv:"function_id"`   // Foreign key to functions table
	Filename     string `csv:"filename"`       // Source file path
	LineNumber   int    `csv:"line_number"`   // Source code line number
	ColumnNumber int    `csv:"column_number"`  // Source code column number
	IsFolded     bool   `csv:"is_folded"`     // Whether multiple symbols map to this address
}

// FunctionRow represents a row in the functions.csv file
type FunctionRow struct {
	FunctionID  int    `csv:"function_id"`   // Unique function identifier (pprof Function.id)
	Name        string `csv:"name"`           // Function name
	SystemName  string `csv:"system_name"`   // System-identified name
	StartLine   int    `csv:"start_line"`    // Function starting line number
}

// MappingRow represents a row in the mappings.csv file
type MappingRow struct {
	MappingID    int    `csv:"mapping_id"`     // Unique mapping identifier
	MemoryStart  string `csv:"memory_start"`   // Binary load address
	MemoryLimit  string `csv:"memory_limit"`   // Address range limit
	Filename     string `csv:"filename"`       // Binary file path
	BuildID      string `csv:"build_id"`       // Unique program version identifier
	HasFunctions bool   `csv:"has_functions"`  // Symbolic info resolution flag
	HasFilenames bool   `csv:"has_filenames"`  // Symbolic info resolution flag
}

// ProfileInfoRow represents a row in the profile_info.csv file
type ProfileInfoRow struct {
	Key   string `csv:"key"`
	Value string `csv:"value"`
}
