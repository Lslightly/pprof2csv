# Go CPU Profile pprof 格式调研报告

## 概述

Go CPU Profile 使用 [protobuf](https://developers.google.com/protocol-buffers) 格式存储，遵循 [`perftools.profiles`](https://github.com/google/pprof/blob/main/proto/profile.proto) schema。该格式由 [Google pprof 工具](https://github.com/google/pprof) 定义，可用于多种性能分析场景。

## 核心数据结构

### 1. Profile Message (根消息)

| 字段 | 类型 | 说明 | Proto 链接 |
|------|------|------|------------|
| sample_type | ValueType[] | 描述样本值的类型和单位 | [sample_type](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53) |
| sample | Sample[] | 记录的所有样本 | [sample](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53) |
| mapping | Mapping[] | 地址范围到二进制文件的映射 | [mapping](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53) |
| location | Location[] | 样本引用的位置信息 | [location](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53) |
| function | Function[] | 位置引用的函数信息 | [function](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53) |
| string_table | string[] | 共享字符串表 | [string_table](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53) |
| period | int64 | 采样间隔（每个样本的事件数） | [period](https://github.com/google/pprof/blob/main/proto/profile.proto#L84) |
| time_nanos | int64 | 采集时间 | [time_nanos](https://github.com/google/pprof/blob/main/proto/profile.proto#L73) |
| duration_nanos | int64 | 采集持续时间 | [duration_nanos](https://github.com/google/pprof/blob/main/proto/profile.proto#L76) |

### 2. Go CPU Profile 的 sample_type

对于 CPU profile，定义了两个 SampleType：

1. **Index 0**: `"samples"` / `"count"` - 样本计数
2. **Index 1**: `"cpu"` / `"nanoseconds"` - CPU 时间（纳秒）

### 3. Sample Message (单个样本)

| 字段 | 类型 | 说明 | Proto 链接 |
|------|------|------|------------|
| location_id | uint64[] | 调用栈位置ID数组，location_id[0] 是栈顶（leaf） | [location_id](https://github.com/google/pprof/blob/main/proto/profile.proto#L95-L102) |
| value | int64[] | 对应 sample_type 的值，对于 CPU profile：<br>- value[0]: 样本计数<br>- value[1]: CPU 时间 (count * period) | [value](https://github.com/google/pprof/blob/main/proto/profile.proto#L95-L102) |
| label | Label[] | 额外上下文信息（如线程ID） | [label](https://github.com/google/pprof/blob/main/proto/profile.proto#L95-L102) |

### 4. Location Message (代码位置)

| 字段 | 类型 | 说明 | Proto 链接 |
|------|------|------|------------|
| id | uint64 | 唯一标识符 | [id](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) |
| mapping_id | uint64 | 关联的 Mapping ID | [mapping_id](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) |
| address | uint64 | 指令地址 | [address](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) |
| line | Line[] | 内联函数信息数组，最后一个条目是调用者 | [line](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) |
| is_folded | bool | 是否多个符号映射到同一地址 | [is_folded](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) |

### 5. Line Message (源代码行)

| 字段 | 类型 | 说明 | Proto 链接 |
|------|------|------|------------|
| function_id | uint64 | 关联的 Function ID | [function_id](https://github.com/google/pprof/blob/main/proto/profile.proto#L174-L180) |
| line | int64 | 源代码行号 | [line](https://github.com/google/pprof/blob/main/proto/profile.proto#L174-L180) |
| column | int64 | 源代码列号 | [column](https://github.com/google/pprof/blob/main/proto/profile.proto#L174-L180) |

### 6. Function Message (函数信息)

| 字段 | 类型 | 说明 | Proto 链接 |
|------|------|------|------------|
| id | uint64 | 唯一标识符 | [id](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) |
| name | int64 | 函数名（string_table 索引） | [name](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) |
| system_name | int64 | 系统识别的名称 | [system_name](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) |
| filename | int64 | 源文件路径 | [filename](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) |
| start_line | int64 | 函数起始行号 | [start_line](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) |

## Go CPU Profile 构建细节

### 采样周期计算

```
period = 1e9 / sampling_rate_in_hz
```

例如：100 Hz 采样率 → period = 10,000,000 纳秒

### 样本值计算

```go
values[0] = e.count        // 样本计数
values[1] = e.count * b.period  // CPU 时间（纳秒）
```

### 调用栈结构

- location_id[0]: 栈顶（正在执行的函数）
- location_id[n]: 栈底（调用链上层函数）

## 数据关系图

```
Profile
  ├─ sample[] ──┬─ location_id[] ──> Location[]
  │              │                   ├─ function_id ──> Function[]
  │              │                   └─ mapping_id ──> Mapping[]
  │              └─ value[] (对应 sample_type)
  ├─ sample_type[] (描述 value 的含义)
  ├─ string_table[] (所有字符串的存储)
  └─ period (采样间隔)
```

## 关键观察

1. **多值样本**: 每个样本可以有多个值（count 和 cpu time）
2. **共享字符串表**: 所有字符串（函数名、文件名）存储在 string_table 中，通过索引引用
3. **ID 引用**: Location、Function、Mapping通过 ID 关联，而非指针
4. **内联函数支持**: 一个 Location 可以包含多个 Line，代表内联调用链
5. **采样粒度**: 每个 sample 代表一段时间内的调用栈快照

## 数据分析维度

基于 pprof 格式，可以分析：

1. **函数级统计**: flat time (自身时间) vs cum time (累计时间)
2. **源代码行级统计**: 每行代码的执行时间
3. **调用关系**: caller-callee 图
4. **时间分布**: 各函数占总时间的比例
5. **热点识别**: 高频调用的函数和代码行

## 参考来源

- [Google pprof protobuf schema](https://github.com/google/pprof/blob/main/proto/profile.proto)
- [Go CPU Profile 实现](https://github.com/golang/go/blob/go1.24.2/src/runtime/pprof/proto.go)