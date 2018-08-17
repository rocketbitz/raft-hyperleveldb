# raft-hyperleveldb
Raft backend using HyperDex's HyperLevelDB

## Use Case

Drop it in instead of BoltDB if your workload is highly concurrent and write intensive. See their [writeup](http://hyperdex.org/performance/leveldb/) for more details

## Example
```go
logStore, _ := rafthyperleveldb.NewHyperLevelDBStore("/path/to/raft/db", rafthyperleveldb.High)

myRaft, _ := raft.NewRaft(config, (*fsm)(c), logStore, logStore, snapshots, transport)
```
## Benchmark
HyperLevelDB:
```
BenchmarkHyperLevelDBStore_FirstIndex-8          1000000              1031 ns/op
BenchmarkHyperLevelDBStore_LastIndex-8           1000000              1477 ns/op
BenchmarkHyperLevelDBStore_GetLog-8              2000000               935 ns/op
BenchmarkHyperLevelDBStore_StoreLog-8             200000              9662 ns/op
BenchmarkHyperLevelDBStore_StoreLogs-8            100000             17323 ns/op
BenchmarkHyperLevelDBStore_DeleteRange-8          100000             19645 ns/op
BenchmarkHyperLevelDBStore_Set-8                  200000              9936 ns/op
BenchmarkHyperLevelDBStore_Get-8                 2000000               795 ns/op
BenchmarkHyperLevelDBStore_SetUint64-8            200000              9086 ns/op
BenchmarkHyperLevelDBStore_GetUint64-8           2000000               856 ns/op
```

## Remarks

Credit to https://github.com/tidwall/raft-leveldb for the original raft-leveldb
