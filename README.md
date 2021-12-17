# fluent-forward-go

`fluent-forward-go` is a fast, memory-efficient implementation of the [Fluent Forward protocol v1](https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1). It allows you to send events to [Fluentd](https://www.fluentd.org/), [Fluent Bit](https://fluentbit.io/), and other endpoints supporting the Fluent protocol. It also includes a websocket client for high-speed proxying of Fluent events over ports such as `80` and `443`.

Features include:

- shared-key authentication
- support for all [Fluent message modes](https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1#message-modes)
- [`gzip` compression](https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1#compressedpackedforward-mode)
- ability to send byte-encoded messages
- `ack` support
- a websocket client for proxying Fluent messages

TODOs:

- TLS support

## Installation

```shell
go get github.com/IBM/fluent-forward-go
```

## Examples

### Create a simple tcp client pointing to `localhost:24224`

```go
c := client.New(client.ConnectionOptions{})
if err := c.Connect(); err != nil {
  // ...
}

defer c.Disconnect()
```

### Send a new log message

```go
record := map[string]interface{}{
  "first": "Sir",
  "last":  "Gawain",
  "equipment": []string{
    "sword",
  },
}
msg := protocol.NewMessage("tag", record)
if err := c.SendMessage(msg); err != nil {
  // ...
}
```

### Send a byte-encoded message

```go
raw := protocol.RawMessage(myMessageBytes)
if err := c.SendMessage(raw); err != nil {
  // ...
}
```

### Wait for `ack`

NOTE: For types other than `RawMessage`, the `SendMessage` function sets the "chunk" option before sending. A `RawMessage` is immutable and must already contain a "chunk" value. The behavior is otherwise identical.

```go
c.RequireAck = true
if err := c.SendMessage(myMsg); err != nil {
  // ...
}
```

## Performance

**tl;dr** `fluent-forward-go` is fast and lean

You can read more about the benchmarks [here](cmd/bm/README.md).

### SendMessage

Run on `localhost`

```shell
Benchmark_Fluent_Forward_Go_SendOnly-8            100000             15726 ns/op               0 B/op          0 allocs/op
```

### Comparisons with `fluent-logger-golang`

The benchmarks below compare `fluent-forward-go` with the official package, [`fluent-logger-golang`](https://github.com/fluent/fluent-logger-golang).

The differences in execution times can vary from one test run to another. The differences in memory allocations, however, are constant.

#### Create and send single message

```shell
# Best of 10
Benchmark_Fluent_Forward_Go_SingleMessage-8       100000             17063 ns/op             400 B/op          3 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-8    100000             19639 ns/op            1216 B/op         16 allocs/op

# Worst of 10
Benchmark_Fluent_Forward_Go_SingleMessage-8       100000             19191 ns/op             400 B/op          3 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-8    100000             21201 ns/op            1216 B/op         16 allocs/op
```

#### Create and send single message with `ack`

```shell
# Best of 10
Benchmark_Fluent_Forward_Go_SingleMessageAck-8              5000           1013919 ns/op             538 B/op          8 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-8           5000           1089037 ns/op            4721 B/op         28 allocs/op

# Worst of 10
Benchmark_Fluent_Forward_Go_SingleMessageAck-8              5000           1125134 ns/op             538 B/op          8 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-8           5000           1493819 ns/op            4721 B/op         28 allocs/op
```

## Developing

### Installation instructions

Before running the generate tool, you must have msgp installed.  To install run:

```shell
go get github.com/tinylib/msgp
```

Afterwards, generate the msgp packets with:

```shell
go generate ./...
```

### Testing

To test against fluent-bit, start up fluent-bit in a docker container with:

```shell
docker pull fluent/fluent-bit:1.8.2
docker run -p 127.0.0.1:24224:24224/tcp -v `pwd`:`pwd` -w `pwd` \
  -ti fluent/fluent-bit:1.8.2 /fluent-bit/bin/fluent-bit \
  -c $(pwd)/fixtures/fluent.conf
```

You can then build and run with:

```shell
go run ./cmd/forward -t foo.bar
```

This will send two regular `Message`s, one with the timestamp as seconds since
the epoch, the other with the timestamp as seconds.nanoseconds.  Those will
then be written to a file with the same name as the tag supplied as the argument
to the `-t` flag (`foo.bar` in the above example).

It will also send a `ForwardMessage` containing a pair of events - these will be
written to the same file.

It will then send a `PackedForwardMessage` containing a pair of events - these
will be written to `$TAG.packed`.

Last, it will send a `CompressedPackedForwardMessage` with the same pair of events, which should then be written to `$TAG.compressed`.
