# mallocgc_percent

Analyzes the percentage of `runtime.mallocgc` cumulative time in a Go pprof profile.

## Usage

```bash
mallocgc_percent -i <profile.pprof> [-show_from <function>] [-denom_func <function>] [-format text|json]
```

## Flags

- `-i`: Input pprof profile file (required)
- `-show_from`: Only include mallocgc samples whose stacktrace contains this function (for numerator)
- `-denom_func`: Function name to use as denominator (default: total profile sample time, or show_from if provided)
- `-format`: Output format: text or json (default: text)

## Features

- Calculates mallocgc time as percentage of total profile time
- Supports filtering mallocgc samples by specific function call stack
- Supports comparing mallocgc time against specific function time

## Examples

Go to [go_parser](../../test/go_parser/) directory.

Analyze mallocgc percentage of total profile time:

```bash
mallocgc_percent -i default.out
```

Analyze mallocgc time under a specific function compared to that function's time:

```bash
mallocgc_percent -i default.out -show_from go/parser.BenchmarkParseOnly
```

Analyze mallocgc percentage using a custom denominator function:

```bash
mallocgc_percent -i default.out -show_from go/parser.BenchmarkParseOnly -denom_func go/parser.BenchmarkParseOnly
```