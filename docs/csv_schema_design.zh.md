# CSV Schema 设计报告

## 设计目标

将 Go CPU Profile pprof 数据转换为 CSV 格式，用于：
1. 关系数据库存储和分析
2. 通过 CodeQL `--external` 选项导入，实现动态数据与静态分析结合

## 设计原则

1. **关系型设计**: 将 pprof 的图结构分解为多个表
2. **规范化**: 避免数据冗余，使用外键引用
3. **可查询性**: 支持常见的性能分析查询
4. **CodeQL 兼容**: 遵循 CodeQL 外部数据格式要求

## 设计理由

本 schema 设计参考了 [pprof 格式调研报告](./pprof_research.zh.md) 中的核心数据结构，主要设计理由如下：

1. **将 Sample 展开为多行**: 根据 [pprof_research - 调用栈结构](./pprof_research.zh.md#调用栈结构)，`location_id[0]` 是栈顶（leaf），`location_id[n]` 是栈底。通过展开调用栈，可以方便地进行 caller-callee 关系查询。

2. **Location 表展开 Line 数组**: 根据 [pprof_research - Location Message](./pprof_research.zh.md#4-location-message-代码位置)，一个 Location 可以包含多个 Line，代表内联调用链。最后一个条目是调用者。

3. **外键关联**: 根据 [pprof_research - 关键观察](./pprof_research.zh.md#关键观察)，pprof 中 Location、Function、Mapping 通过 ID 关联，而非指针。因此 schema 使用外键（location_id, mapping_id, function_id）来建立关系。

## Schema 设计

### 表 1: samples.csv (样本表)

存储每个调用栈样本的原始数据。

| 列名 | 类型 | 说明 | pprof 对应字段 | CodeQL 类型 |
|------|------|------|----------------|-------------|
| sample_id | int | 唯一样本标识符 | 新增 | int |
| stack_depth | int | 调用栈深度 | Sample.location_id 长度 | int |
| location_id | int | Location 表的外键（该层调用的位置ID） | Sample.location_id[] | int |
| depth | int | 栈中的深度（0=栈顶/leaf） | 新增（location_id 数组索引） | int |
| count | int | 样本计数 | [Sample.value[0]](https://github.com/google/pprof/blob/main/proto/profile.proto#L95-L102) | int |
| cpu_nanos | int | CPU 时间（纳秒） | [Sample.value[1]](https://github.com/google/pprof/blob/main/proto/profile.proto#L95-L102) | int |

**说明**:
- 每个调用栈样本会被展开为多行，每行对应栈中的一层
- 这样可以方便地进行 caller-callee 关系查询

**示例数据**:
```csv
sample_id,stack_depth,location_id,depth,count,cpu_nanos
1,3,101,0,10,100000000
1,3,102,1,10,100000000
1,3,103,2,10,100000000
2,2,201,0,5,50000000
2,2,202,1,5,50000000
```

### 表 2: locations.csv (代码位置表)

存储代码位置信息（地址、文件、行号）。

| 列名 | 类型 | 说明 | pprof 对应字段 | CodeQL 类型 |
|------|------|------|----------------|-------------|
| location_id | int | 唯一位置标识符 | [Location.id](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) | int |
| mapping_id | int | Mapping 表的外键 | [Location.mapping_id](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) | int |
| address | string | 指令地址（十六进制字符串） | [Location.address](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) | string |
| function_id | int | Function 表的外键 | [Line.function_id](https://github.com/google/pprof/blob/main/proto/profile.proto#L174-L180) | int |
| filename | string | 源文件路径 | [Function.filename](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) | string |
| line_number | int | 源代码行号 | [Line.line](https://github.com/google/pprof/blob/main/proto/profile.proto#L174-L180) | int |
| column_number | int | 源代码列号 | [Line.column](https://github.com/google/pprof/blob/main/proto/profile.proto#L174-L180) | int |
| is_folded | boolean | 是否多个符号映射到同一地址 | [Location.is_folded](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171) | boolean |

**说明**:
- 对于包含内联函数的 Location，每个内联层级会创建单独的行
- main() 中，如果 location 包含 3 个 Line 条目，每个条目会生成一行

**示例数据**:
```csv
location_id,mapping_id,address,function_id,filename,line_number,column_number,is_folded
101,1,0x401234,501,/path/to/file1.go,42,0,false
102,1,0x401567,502,/path/to/file2.go,15,0,false
```

### 表 3: functions.csv (函数表)

存储函数信息。

| 列名 | 类型 | 说明 | pprof 对应字段 | CodeQL 类型 |
|------|------|------|----------------|-------------|
| function_id | int | 唯一函数标识符 | [Function.id](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) | int |
| name | string | 函数名 | [Function.name](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) | string |
| system_name | string | 系统识别的名称 | [Function.system_name](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) | string |
| start_line | int | 函数起始行号 | [Function.start_line](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198) | int |

**示例数据**:
```csv
function_id,name,system_name,start_line
501,main.main,main.main,1
5012,example.Process,example.Process,10
```

### 表 4: mappings.csv (二进制映射表)

存储二进制文件映射信息。

| 列名 | 类型 | 说明 | pprof 对应字段 | CodeQL 类型 |
|------|------|------|----------------|-------------|
| mapping_id | int | 唯一映射标识符 | [Mapping.id](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) | int |
| memory_start | string | 二进制加载地址 | [Mapping.memory_start](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) | string |
| memory_limit | string | 地址范围上限 | [Mapping.memory_limit](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) | string |
| filename | string | 二进制文件路径 | [Mapping.filename](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) | string |
| build_id | string | 构建版本标识 | [Mapping.build_id](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) | string |
| has_functions | boolean | 是否包含符号信息 | [Mapping.has_functions](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) | boolean |
| has_filenames | boolean | 是否包含文件名信息 | [Mapping.has_filenames](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152) | boolean |

**示例数据**:
```csv
mapping_id,memory_start,memory_limit,filename,build_id,has_functions,has_filenames
1,0x400000,0x500000,/path/to/binary,abc123,true,true
```

### 表 5: profile_info.csv (概要信息表)

存储 profile 的整体信息。

| 列名 | 类型 | 说明 | pprof 对应字段 | CodeQL 类型 |
|------|------|------|----------------|-------------|
| key | string | 配置项名称 | - | string |
| value | string | 配置项值 | - | string |

**示例数据**:
```csv
key,value
sampling_period,10000000
profile_duration,30000000000
sample_type_0,samples
sample_type_1,cpu
```

**说明**: profile_info 表存储以下元数据：
- `sampling_period`: 采样间隔，对应 [Profile.period](https://github.com/google/pprof/blob/main/proto/profile.proto#L84)
- `profile_duration`: 采集持续时间，对应 [Profile.duration_nanos](https://github.com/google/pprof/blob/main/proto/profile.proto#L76)
- `sample_type_*`: 样本类型，对应 [Profile.sample_type](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53)

## 索引建议

为了高效查询，建议在以下列上创建索引：

```sql
-- samples 表
CREATE INDEX idx_sample_id ON samples(sample_id);
CREATE INDEX idx_location_id ON samples(location_id);

-- locations 表
CREATE INDEX idx_loc_func ON locations(function_id);
CREATE INDEX idx_addr ON locations(address);

-- functions 表
CREATE INDEX idx_func_name ON functions(name);
```

## CodeQL 使用示例

### 定义外部数据表

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

### 查询示例：查找热点函数

```ql
from string name, int count
where functionName(_, name) and sampleCount(_, count)
select name, count order by count desc
```

## 导出策略

1. **samples.csv**: 遍历所有 [Sample](https://github.com/google/pprof/blob/main/proto/profile.proto#L95-L102)，展开调用栈生成多行
2. **locations.csv**: 遍历所有 [Location](https://github.com/google/pprof/blob/main/proto/profile.proto#L154-L171)，对每个 Location 的 [Line](https://github.com/google/pprof/blob/main/proto/profile.proto#L174-L180) 数组生成多行
3. **functions.csv**: 直接导出所有 [Function](https://github.com/google/pprof/blob/main/proto/profile.proto#L183-L198)
4. **mappings.csv**: 直接导出所有 [Mapping](https://github.com/google/pprof/blob/main/proto/profile.proto#L119-L152)
5. **profile_info.csv**: 导出 [Profile](https://github.com/google/pprof/blob/main/proto/profile.proto#L47-L53) 的元数据

## 兼容性考虑

1. **向后兼容**: 保留原有的 source line → flat/cum 格式
2. **可选导出**: 新的 schema 可通过新的 flag（如 `--schema=relational`）启用
3. **增量导出**: 支持只导出部分表（如只导出 functions.csv）

## 下一步实现

1. 在 `analyzer` 包中添加新的导出函数
2. 创建 `schema` 包定义 CSV schema
3. 在 CLI 添加 `--schema` flag 选择导出格式
4. 编写测试验证导出数据的正确性