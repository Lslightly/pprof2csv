# pprof to csv

Convert pprof profile to csv files for further analysis.

Features:
- Source line - Time mapping
- Source line - Memory Consumption mapping
- ...


## pprof protobuf design

- pprof protobuf design: [proto/README.md](https://github.com/google/pprof/blob/main/proto/README.md) in [google/pprof](https://github.com/google/pprof)
- Go CPUProfile format: [`(*profileBuilder).build`](https://github.com/golang/go/blob/go1.24.2/src/runtime/pprof/proto.go#L348-L392)
- Go MemProfile format: [`writeHeapProto`](https://github.com/golang/go/blob/go1.24.2/src/runtime/pprof/protomem.go#L16-L68)
    - [ ] Go MemProfile is not supported yet.

