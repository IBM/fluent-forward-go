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
go test -benchmem -run=^$ -bench ^.*Message$ -benchtime=10000x -count=10 github.com/aanujj/fluent-forward-go/cmd/bm/fluent_forward_go
go test -benchmem -run=^$ -bench ^.*Message$ -benchtime=10000x -count=10 github.com/aanujj/fluent-forward-go/cmd/bm/fluent_logger_golang

# with ack
go test -benchmem -run=^$ -bench ^.*MessageAck$ -benchtime=10000x -count=10 github.com/aanujj/fluent-forward-go/cmd/bm/fluent_forward_go
go test -benchmem -run=^$ -bench ^.*MessageAck$ -benchtime=10000x -count=10 github.com/aanujj/fluent-forward-go/cmd/bm/fluent_logger_golang
```

#### Best of 10: create and send single message

```shell
Benchmark_Fluent_Forward_Go_SingleMessage-16        10000	     11355 ns/op	      48 B/op	       1 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-16     10000	     19687 ns/op	    2169 B/op	      33 allocs/op
```

#### Best of 10: create and send single message with `ack`

```shell
Benchmark_Fluent_Forward_Go_SingleMessageAck-16        10000	    768743 ns/op	     185 B/op	       6 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-16     10000	    793360 ns/op	    6015 B/op	      47 allocs/op
```

#### Full results

##### `fluent-forward-go`

```shell
pkg: github.com/aanujj/fluent-forward-go/cmd/bm/fluent_forward_go
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz

Benchmark_Fluent_Forward_Go_SingleMessage-16        	   10000	     13153 ns/op	      48 B/op	       1 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-16        	   10000	     12776 ns/op	      48 B/op	       1 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-16        	   10000	     12710 ns/op	      48 B/op	       1 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-16        	   10000	     13048 ns/op	      48 B/op	       1 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-16        	   10000	     12228 ns/op	      48 B/op	       1 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-16        	   10000	     12250 ns/op	      48 B/op	       1 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-16        	   10000	     11355 ns/op	      48 B/op	       1 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-16        	   10000	     12445 ns/op	      48 B/op	       1 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-16        	   10000	     12959 ns/op	      48 B/op	       1 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessage-16        	   10000	     11597 ns/op	      48 B/op	       1 allocs/op

Benchmark_Fluent_Forward_Go_SingleMessageAck-16      	   10000	    777020 ns/op	     184 B/op	       6 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-16      	   10000	    768743 ns/op	     185 B/op	       6 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-16      	   10000	    787335 ns/op	     185 B/op	       6 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-16      	   10000	    786457 ns/op	     185 B/op	       6 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-16      	   10000	    796123 ns/op	     185 B/op	       6 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-16      	   10000	    781143 ns/op	     185 B/op	       6 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-16      	   10000	    819758 ns/op	     185 B/op	       6 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-16      	   10000	    811781 ns/op	     185 B/op	       6 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-16      	   10000	    800595 ns/op	     185 B/op	       6 allocs/op
Benchmark_Fluent_Forward_Go_SingleMessageAck-16      	   10000	    885662 ns/op	     185 B/op	       6 allocs/op
```

#### `fluent-logger-golang`

```shell
goos: darwin
goarch: amd64
pkg: github.com/aanujj/fluent-forward-go/cmd/bm/fluent_logger_golang
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz

Benchmark_Fluent_Logger_Golang_SingleMessage-16    	   10000	     20002 ns/op	    2171 B/op	      33 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-16    	   10000	     20167 ns/op	    2170 B/op	      33 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-16    	   10000	     20983 ns/op	    2169 B/op	      33 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-16    	   10000	     19779 ns/op	    2170 B/op	      33 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-16    	   10000	     19687 ns/op	    2169 B/op	      33 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-16    	   10000	     19893 ns/op	    2169 B/op	      33 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-16    	   10000	     20014 ns/op	    2170 B/op	      33 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-16    	   10000	     20163 ns/op	    2170 B/op	      33 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-16    	   10000	     19819 ns/op	    2170 B/op	      33 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-16    	   10000	     19796 ns/op	    2169 B/op	      33 allocs/op

Benchmark_Fluent_Logger_Golang_SingleMessageAck-16       10000	    823867 ns/op	    6015 B/op	      47 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-16       10000	    891730 ns/op	    6013 B/op	      47 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-16       10000	    800438 ns/op	    6012 B/op	      47 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-16       10000	    793360 ns/op	    6015 B/op	      47 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-16       10000	    843148 ns/op	    6014 B/op	      47 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-16       10000	    816468 ns/op	    6011 B/op	      47 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-16       10000	    833102 ns/op	    6013 B/op	      47 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-16       10000	    809983 ns/op	    6014 B/op	      47 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-16       10000	    848345 ns/op	    6015 B/op	      47 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-16       10000	    846259 ns/op	    6013 B/op	      47 allocs/op
```
