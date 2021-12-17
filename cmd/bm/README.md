# Benchmarks

Very helpful information on Go benchmarking: https://dave.cheney.net/high-performance-go-workshop/gopherchina-2019.html

## Recent results

### `fluent-forward-go` vs `fluent-logger-golang` v1.8.0

#### Running the comparisons

The benchmark packages must be run separately. Running them together generates an error because `fluent-forward-go` and `fluent-logger-golang` each tries to register the same extension with `msgp`, which results in an error.

1) Start Fluent Bit

```shell
❱❱ docker run -p 127.0.0.1:24224:24224/tcp -v `pwd`:`pwd` -w `pwd` -ti fluent/fluent-bit:1.8.2 /fluent-bit/bin/fluent-bit   -c $(pwd)/fixtures/fluent.conf
```

2) Run the benchmarks

```shell
# no ack
go test -benchmem -run=^$ -bench ^.*Message$ -benchtime=100000x -count=10 github.com/IBM/fluent-forward-go/cmd/bm/fluent_forward_go
go test -benchmem -run=^$ -bench ^.*Message$ -benchtime=100000x -count=10 github.com/IBM/fluent-forward-go/cmd/bm/fluent_logger_golang

# with ack
go test -benchmem -run=^$ -bench ^.*MessageAck$ -benchtime=5000x -count=10 github.com/IBM/fluent-forward-go/cmd/bm/fluent_forward_go
go test -benchmem -run=^$ -bench ^.*MessageAck$ -benchtime=5000x -count=10 github.com/IBM/fluent-forward-go/cmd/bm/fluent_logger_golang
```

#### Best of 10: create and send single message

```shell
Benchmark_Fluent_Forward_Go_SingleMessage-8       100000             17063 ns/op             400 B/op          3 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-8    100000             19639 ns/op            1216 B/op         16 allocs/op
```

#### Best of 10: create and send single message with `ack`

```shell
Benchmark_Fluent_Forward_Go_SingleMessageAck-8              5000           1013919 ns/op             538 B/op          8 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-8           5000           1089037 ns/op            4721 B/op         28 allocs/op
```

#### Worst of 10: create and send single message

```shell
Benchmark_Fluent_Forward_Go_SingleMessage-8       100000             19191 ns/op             400 B/op          3 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-8    100000             21201 ns/op            1216 B/op         16 allocs/op
```

#### Worst of 10: create and send single message with `ack`

```shell
Benchmark_Fluent_Forward_Go_SingleMessageAck-8              5000           1125134 ns/op             538 B/op          8 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-8           5000           1493819 ns/op            4721 B/op         28 allocs/op
```

#### Full results

##### `fluent-forward-go`

```shell
pkg: github.com/IBM/fluent-forward-go/cmd/bm/fluent_forward_go
cpu: Intel(R) Core(TM) i7-7820HQ CPU @ 2.90GHz

Benchmark_Fluent_Forward_Go_SingleMessage-8       100000             19191 ns/op             400 B/op          3 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-8       100000             17063 ns/op             400 B/op          3 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-8       100000             17558 ns/op             400 B/op          3 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-8       100000             16959 ns/op             400 B/op          3 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-8       100000             17168 ns/op             400 B/op          3 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-8       100000             17335 ns/op             400 B/op          3 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-8       100000             18235 ns/op             400 B/op          3 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-8       100000             19008 ns/op             400 B/op          3 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-8       100000             17168 ns/op             400 B/op          3 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-8       100000             17734 ns/op             400 B/op          3 allocs/op

Benchmark_Fluent_Forward_Go_SingleMessageAck-8              5000           1125134 ns/op             538 B/op          8 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-8              5000           1052649 ns/op             538 B/op          8 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-8              5000           1038997 ns/op             538 B/op          8 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-8              5000           1038209 ns/op             537 B/op          8 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-8              5000           1046011 ns/op             538 B/op          8 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-8              5000           1026626 ns/op             538 B/op          8 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-8              5000           1013919 ns/op             538 B/op          8 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-8              5000           1048560 ns/op             538 B/op          8 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-8              5000           1032720 ns/op             537 B/op          8 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-8              5000           1046877 ns/op             538 B/op          8 allocs/op
```

#### `fluent-logger-golang`

```shell
pkg: github.com/IBM/fluent-forward-go/cmd/bm/fluent_logger_golang
cpu: Intel(R) Core(TM) i7-7820HQ CPU @ 2.90GHz

Benchmark_Fluent_Logger_Golang_SingleMessage-8            100000             20853 ns/op            1216 B/op         16 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-8            100000             20245 ns/op            1216 B/op         16 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-8            100000             21038 ns/op            1216 B/op         16 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-8            100000             20528 ns/op            1216 B/op         16 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-8            100000             20236 ns/op            1216 B/op         16 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-8            100000             19639 ns/op            1216 B/op         16 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-8            100000             20868 ns/op            1216 B/op         16 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-8            100000             20231 ns/op            1216 B/op         16 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-8            100000             20784 ns/op            1216 B/op         16 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-8            100000             21201 ns/op            1216 B/op         16 allocs/op

Benchmark_Fluent_Logger_Golang_SingleMessageAck-8           5000           1095497 ns/op            4721 B/op         28 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-8           5000           1089037 ns/op            4721 B/op         28 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-8           5000           1106823 ns/op            4721 B/op         28 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-8           5000           1195079 ns/op            4721 B/op         28 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-8           5000           1386415 ns/op            4721 B/op         28 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-8           5000           1493819 ns/op            4721 B/op         28 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-8           5000           1424519 ns/op            4721 B/op         28 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-8           5000           1439423 ns/op            4721 B/op         28 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-8           5000           1365458 ns/op            4721 B/op         28 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-8           5000           1457163 ns/op            4721 B/op         28 allocs/op
```
