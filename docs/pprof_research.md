# Go CPU Profile pprof Format Research Report

## Overview

Go CPU Profile uses [protobuf](https://developers.google.com/protocol-buffers) format, following the [`perftools.profiles`](https://github.com/google/pprof/blob/main/proto/profile.proto) schema. This format is defined by the [Google pprof tool](https://github.com/google/pprof) and can be used for various performance analysis scenarios.

## Core Data Structures

### 1. Profile Message (Root Message)

| Field | Type | Description | Proto Link |
|-------|------|-------------|------------|
| sample_type | ValueType[] | Describes the type and unit of sample values | [sample_type](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53) |
| sample | Sample[] | All recorded samples | [sample](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53) |
| mapping | Mapping[] | Mapping from address ranges to binary files | [mapping](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53) |
| location | Location[] | Location information referenced by samples | [location](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53) |
| function | Function[] | Function information referenced by locations | [function](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53) |
| string_table | string[] | Shared string table | [string_table](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53) |
| period | int64 | Sampling interval (number of events per sample) | [period](https://github.com/google/pprof/blob/main/proto/profile.proto#L84) |
| time_nanos | int64 | Collection timestamp | [time_nanos](https://github.com/google/pprof/blob/main/proto/profile.proto#L73) |
| duration_nanos | int64 | Collection duration | [duration_nanos](https://github.com/google/pprof/blob/main/proto/profile.proto#L76) |

### 2. Sample Type for Go CPU Profile

For CPU profiles, two SampleTypes are defined:

1. **Index 0**: `"samples"` / `"count"` - Sample count
2. **Index 1**: `"cpu"` / `"nanoseconds"` - CPU time (nanoseconds)

### 3. Sample Message (Single Sample)

| Field | Type | Description | Proto Link |
|-------|------|-------------|------------|
| location_id | uint64[] | Array of location IDs in the call stack, location_id[0] is the stack top (leaf) | [location_id](https://github.com/google/pprof/blob/main/proto/profile.proto#L95-L102) |
| value | int64[] | Values corresponding to sample_type. For CPU profile:<br>- value[0]: Sample count<br>- value[1]: CPU time (count * period) | [value](https://github.com/google/pprof/blob/main/proto/profile.proto#L95-L102) |
| label | Label[] | Additional context information (e.g., thread ID) | [label](https://github.com/google/pprof/blob/main/proto/profile.proto#L95-L102) |

### 4. Location Message (Code Location)

| Field | Type | Description | Proto Link |
|-------|------|-------------|------------|
| id | uint64 | Unique identifier | [id](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) |
| mapping_id | uint64 | Associated Mapping ID | [mapping_id](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) |
| address | uint64 | Instruction address | [address](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) |
| line | Line[] | Array of inline function information; the last entry is the caller | [line](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) |
| is_folded | bool | Whether multiple symbols map to the same address | [is_folded](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) |

### 5. Line Message (Source Code Line)

| Field | Type | Description | Proto Link |
|-------|------|-------------|------------|
| function_id | uint64 | Associated Function ID | [function_id](https://github.com/google/pprof/blob/main/proto/profile.proto#L174-L180) |
| line | int64 | Source code line number | [line](https://github.com/google/pprof/blob/main/proto/profile.proto#L174-L180) |
| column | int64 | Source code column number | [column](https://github.com/google/pprof/blob/main/proto/profile.proto#L174-L180) |

### 6. Function Message (Function Information)

| Field | Type | Description | Proto Link |
|-------|------|-------------|------------|
| id | uint64 | Unique identifier | [id](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) |
| name | int64 | Function name (string_table index) | [name](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) |
| system_name | int64 | System-recognized name | [system_name](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) |
| filename | int64 | Source file path | [filename](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) |
| start_line | int64 | Function start line number | [start_line](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) |

## Go CPU Profile Construction Details

### Sampling Period Calculation

```
period = 1e9 / sampling_rate_in_hz
```

For example: 100 Hz sampling rate → period = 10,000,000 nanoseconds

### Sample Value Calculation

```go
values[0] = e.count        // Sample count
values[1] = e.count * b.period  // CPU time (nanoseconds)
```

### Call Stack Structure

- location_id[0]: Stack top (function being executed)
- location_id[n]: Stack bottom (caller in the call chain)

## Data Relationship Diagram

```
Profile
  ├─ sample[] ──┬─ location_id[] ──> Location[]
  │              │                   ├─ function_id ──> Function[]
  │              │                   └─ mapping_id ──> Mapping[]
  │              └─ value[] (corresponds to sample_type)
  ├─ sample_type[] (describes meaning of values)
  ├─ string_table[] (storage for all strings)
  └─ period (sampling interval)
```

## Key Observations

1. **Multi-value Samples**: Each sample can have multiple values (count and CPU time)
2. **Shared String Table**: All strings (function names, file names) are stored in string_table and referenced by index
3. **ID References**: Location, Function, and Mapping are related by ID, not pointers
4. **Inline Function Support**: A Location can contain multiple Lines, representing the inline call chain
5. **Sampling Granularity**: Each sample represents a call stack snapshot over a period of time

## Data Analysis Dimensions

Based on the pprof format, you can analyze:

1. **Function-level Statistics**: Flat time (self time) vs. cum time (cumulative time)
2. **Source Line-level Statistics**: Execution time for each line of code
3. **Call Relationships**: Caller-callee graph
4. **Time Distribution**: Proportion of time spent in each function
5. **Hotspot Identification**: Frequently called functions and code lines

## References

- [Google pprof protobuf schema](https://github.com/google/pprof/blob/main/proto/profile.proto)
- [Go CPU Profile Implementation](https://github.com/golang/go/blob/go1.24.2/src/runtime/pprof/proto.go)