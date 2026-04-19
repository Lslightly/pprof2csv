# 实现过程记录：Go CPU Profile 转 CSV Schema 导出

## 任务背景

用户希望将 Go CPU Profile pprof 数据转换为关系型 CSV 格式，用于：
1. 关系数据库存储和分析
2. 通过 CodeQL `--external` 选项导入，实现动态数据与静态分析结合

## 实现过程

### 1. 调研阶段

由于 WebSearch 无法正常返回结果，用户安装了 Tavily MCP。

**使用 MCP 搜索获取的信息：**
- pprof protobuf schema (`perftools.profiles`)
- Go CPU Profile 的 SampleType：samples/count + cpu/nanoseconds
- Profile 的核心结构：Sample, Location, Line, Function, Mapping
- CodeQL 外部数据表格式（未找到完整文档，但明确了需求）

**生成的调研文档：**
- `docs/pprof_research.md` - pprof protobuf 格式详解
- `docs/csv_schema_design.md` - 关系型 CSV schema 设计

### 2. 设计阶段

设计了 5 个关系型 CSV 表：

| 表名 | 说明 |
|------|------|
| samples.csv | 调用栈样本（每层展开一行） |
| locations.csv | 代码位置（内联函数展开） |
| functions.csv | 函数信息 |
| mappings.csv | 二进制映射 |
| profile_info.csv | profile 元数据 |

### 3. 实现阶段

**创建的文件：**

1. `schema/models.go` - 数据结构定义
2. `schema/exporter.go` - CSV 导出器
3. `cmd/schema_export/main.go` - CLI 工具

**修复的问题：**
- Mapping 字段：MemoryStart → Start, MemoryLimit → Limit, Filename → File
- DefaultSampleType 类型：int → string

### 4. 测试验证

```bash
./schema_export -i test/loop/cpu.pprof -o schema_test/
```

成功生成 5 个 CSV 文件，验证通过。

## 使用方法

```bash
# 构建
go build -o schema_export ./cmd/schema_export/

# 运行
./schema_export -i profile.pprof -o output_dir/
```

## 后续可扩展

- 添加 `--show_from` 过滤功能（复用 analyzer 包的逻辑）
- 添加 `--format` 选项支持 JSON 输出
- 集成到主工具 pprof2csv 作为 `--schema` 选项