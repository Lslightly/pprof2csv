# CSV Schema Design Report

## Design Goals

Convert Go CPU Profile pprof data to CSV format for:
1. Relational database storage and analysis
2. Import via CodeQL `--external` option to combine dynamic data with static analysis

## Design Principles

1. **Relational Design**: Decompose pprof's graph structure into multiple tables
2. **Normalization**: Avoid data redundancy, use foreign key references
3. **Queryability**: Support common performance analysis queries
4. **CodeQL Compatibility**: Follow CodeQL external data format requirements

## Design Rationale

This schema design references the core data structures from the [pprof format research report](./pprof_research.md). Key design decisions:

1. **Expanding Sample into multiple rows**: According to [pprof_research - Call Stack Structure](./pprof_research.md#call-stack-structure), `location_id[0]` is the stack top (leaf), and `location_id[n]` is the stack bottom. By expanding the call stack, we can easily query caller-callee relationships.

2. **Expanding Location's Line array**: According to [pprof_research - Location Message](./pprof_research.md#4-location-message-code-location), a Location can contain multiple Lines, representing the inline call chain. The last entry is the caller.

3. **Foreign key relationships**: According to [pprof_research - Key Observations](./pprof_research.md#key-observations), in pprof, Location, Function, and Mapping are related by ID, not pointers. Therefore, the schema uses foreign keys (location_id, mapping_id, function_id) to establish relationships.

## Schema Design

### Table 1: samples.csv (Samples Table)

Stores raw data for each call stack sample.

| Column | Type | Description | pprof Field | CodeQL Type |
|--------|------|-------------|-------------|-------------|
| sample_id | int | Unique sample identifier | New field | int |
| stack_depth | int | Call stack depth | Sample.location_id length | int |
| location_id | int | Foreign key to Location table (location ID for this stack level) | Sample.location_id[] | int |
| depth | int | Depth in the stack (0 = stack top / leaf) | New field (location_id array index) | int |
| count | int | Sample count | [Sample.value[0]](https://github.com/google/pprof/blob/main/proto/profile.proto#L95-L102) | int |
| cpu_nanos | int | CPU time (nanoseconds) | [Sample.value[1]](https://github.com/google/pprof/blob/main/proto/profile.proto#L95-L102) | int |

**Description**:
- Each call stack sample is expanded into multiple rows, one for each stack level
- This enables convenient caller-callee relationship queries

**Sample Data**:
```csv
sample_id,stack_depth,location_id,depth,count,cpu_nanos
1,3,101,0,10,100000000
1,3,102,1,10,100000000
1,3,103,2,10,100000000
2,2,201,0,5,50000000
2,2,202,1,5,50000000
```

### Table 2: locations.csv (Locations Table)

Stores code location information (address, file, line number).

| Column | Type | Description | pprof Field | CodeQL Type |
|--------|------|-------------|-------------|-------------|
| location_id | int | Unique location identifier | [Location.id](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) | int |
| mapping_id | int | Foreign key to Mapping table | [Location.mapping_id](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) | int |
| address | string | Instruction address (hex string) | [Location.address](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) | string |
| function_id | int | Foreign key to Function table | [Line.function_id](https://github.com/google/pprof/blob/main/proto/profile.proto#L174-L180) | int |
| filename | string | Source file path | [Function.filename](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) | string |
| line_number | int | Source code line number | [Line.line](https://github.com/google/pprof/blob/main/proto/profile.proto#L174-L180) | int |
| column_number | int | Source code column number | [Line.column](https://github.com/google/pprof/blob/main/proto/profile.proto#L174-L180) | int |
| is_folded | boolean | Whether multiple symbols map to the same address | [Location.is_folded](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) | boolean |

**Description**:
- For Locations containing inline functions, each inline level creates a separate row
- In main(), if a location contains 3 Line entries, each entry generates one row

**Sample Data**:
```csv
location_id,mapping_id,address,function_id,filename,line_number,column_number,is_folded
101,1,0x401234,501,/path/to/file1.go,42,0,false
102,1,0x401567,502,/path/to/file2.go,15,0,false
```

### Table 3: functions.csv (Functions Table)

Stores function information.

| Column | Type | Description | pprof Field | CodeQL Type |
|--------|------|-------------|-------------|-------------|
| function_id | int | Unique function identifier | [Function.id](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) | int |
| name | string | Function name | [Function.name](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) | string |
| system_name | string | System-recognized name | [Function.system_name](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) | string |
| start_line | int | Function start line number | [Function.start_line](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) | int |

**Sample Data**:
```csv
function_id,name,system_name,start_line
501,main.main,main.main,1
5012,example.Process,example.Process,10
```

### Table 4: mappings.csv (Mappings Table)

Stores binary file mapping information.

| Column | Type | Description | pprof Field | CodeQL Type |
|--------|------|-------------|-------------|-------------|
| mapping_id | int | Unique mapping identifier | [Mapping.id](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) | int |
| memory_start | string | Binary load address | [Mapping.memory_start](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) | string |
| memory_limit | string | Upper limit of address range | [Mapping.memory_limit](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) | string |
| filename | string | Binary file path | [Mapping.filename](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) | string |
| build_id | string | Build version identifier | [Mapping.build_id](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) | string |
| has_functions | boolean | Whether it contains symbol information | [Mapping.has_functions](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) | boolean |
| has_filenames | boolean | Whether it contains filename information | [Mapping.has_filenames](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) | boolean |

**Sample Data**:
```csv
mapping_id,memory_start,memory_limit,filename,build_id,has_functions,has_filenames
1,0x400000,0x500000,/path/to/binary,abc123,true,true
```

### Table 5: profile_info.csv (Profile Info Table)

Stores overall profile information.

| Column | Type | Description | pprof Field | CodeQL Type |
|--------|------|-------------|-------------|-------------|
| key | string | Configuration item name | - | string |
| value | string | Configuration item value | - | string |

**Sample Data**:
```csv
key,value
sampling_period,10000000
profile_duration,30000000000
sample_type_0,samples
sample_type_1,cpu
```

**Description**: The profile_info table stores the following metadata:
- `sampling_period`: Sampling interval, corresponds to [Profile.period](https://github.com/google/pprof/blob/main/proto/profile.proto#L84)
- `profile_duration`: Collection duration, corresponds to [Profile.duration_nanos](https://github.com/google/pprof/blob/main/proto/profile.proto#L76)
- `sample_type_*`: Sample type, corresponds to [Profile.sample_type](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53)

## Index Recommendations

For efficient queries, it is recommended to create indexes on the following columns:

```sql
-- samples table
CREATE INDEX idx_sample_id ON samples(sample_id);
CREATE INDEX idx_location_id ON samples(location_id);

-- locations table
CREATE INDEX idx_loc_func ON locations(function_id);
CREATE INDEX idx_addr ON locations(address);

-- functions table
CREATE INDEX idx_func_name ON functions(name);
```

## CodeQL Usage Examples

### Defining External Data Tables

```ql
external data sampleCount(
  int sample_id,
  int count
)

external data sampleLocation(
  int sample_id,
  int location_id
)

external data locationFunction(
  int location_id,
  int function_id
)

external data functionName(
  int function_id,
  string name
)
```

### Query Example: Finding Hot Functions

```ql
from string name, int count
where functionName(_, name) and sampleCount(_, count)
select name, count order by count desc
```

## Export Strategy

1. **samples.csv**: Iterate through all [Sample](https://github.com/google/pprof/blob/main/proto/profile.proto#L95-L102), expand call stacks to generate multiple rows
2. **locations.csv**: Iterate through all [Location](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171), generate multiple rows for each Location's [Line](https://github.com/google/pprof/blob/main/proto/profile.proto#L174-L180) array
3. **functions.csv**: Export all [Function](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) directly
4. **mappings.csv**: Export all [Mapping](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) directly
5. **profile_info.csv**: Export [Profile](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53) metadata

## Compatibility Considerations

1. **Backward Compatibility**: Keep the original source line → flat/cum format
2. **Optional Export**: New schema can be enabled via new flags (e.g., `--schema=relational`)
3. **Incremental Export**: Support exporting only some tables (e.g., only functions.csv)

## Next Steps for Implementation

1. Add new export functions in the `analyzer` package
2. Create `schema` package to define CSV schema
3. Add `--schema` flag to CLI to select export format
4. Write tests to verify export data correctness