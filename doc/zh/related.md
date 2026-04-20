# 相关工作调研

pprof2csv 目前相关的工具好像支持很少。相关的包括:
- [google/csv2pprof](https://github.com/google/csv2pprof)， csv 转 pprof。
- [conbon/pprof2csv](https://github.com/conbon/pprof2csv) 1 star，通过正则匹配 `pprof -text` 和 `pprof -top` 的输出得到。效果没有测试。我认为可能缺失一些信息，而且无法了解 pprof 是怎么运作的。
- [felixge/pprofutils](https://github.com/felixge/pprofutils) 148 stars，2 年前最后一次更新。README 搜 csv 没有搜到，可能不支持 csv 转换。功能包括如下：
	- anon匿名化 pkg, file, function names with human readable hashes
	- avg 将一个 block/mutex profile 转换为每次竞争的平均时间
	- folded pprof 和 Brendan Gregg's folded text format 之间互相转换
	- heappage 添加虚拟帧来展示 Go 内存分配的平均生命期
	- jemalloc 将 jemalloc heap profile 转为 pprof 格式
	- json pprof 和 json 格式相互转换

> 注意 Google 的 pprof 工具和 `go tool pprof` 并不完全一致。`go tool pprof` 一般会用更老一些的版本。比如 go1.24.2 依赖的 `google/pprof` 版本为 `v0.0.0-20241101162523-b92577c0c142`。此外根据 `go tool pprof` [自己的说法](https://github.com/golang/go/blob/49860cf92a9a3ba434d2bc393faaefabe48d181e/src/cmd/pprof/README)，由于 pprof 实际上对 C++, Java, Go 都适用，因此它的功能会过于泛化。这导致 `go tool pprof` 会保留抽象不动（没明白什么意思，可能是实现了一些不必要的功能来满足抽象），开发者不要将 `go tool pprof` 中的抽象层级当做例子。

