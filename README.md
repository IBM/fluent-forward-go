# fluent-forward-go

`fluent-forward-go` is a fast, memory-efficient implementation of the [Fluent Forward v1 specification](https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1). It allows you to send events to [Fluentd](https://www.fluentd.org/), [Fluent Bit](https://fluentbit.io/), and other endpoints supporting the Fluent protocol. It also includes a websocket client for high-speed proxying of Fluent events over ports such as `80` and `443`.

Features include:

- TCP, TLS, mTLS, and unix socket transport
- shared-key authentication
- support for all [Fluent message modes](https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1#message-modes)
- [`gzip` compression](https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1#compressedpackedforward-mode)
- ability to send byte-encoded messages
- `ack` support
- a websocket client for proxying Fluent messages


## Installation

```shell
go get github.com/IBM/fluent-forward-go
```

## Examples

### Create a TCP client

```go
c := client.New(client.ConnectionOptions{
  Factory: &client.ConnFactory{
    Address: "localhost:24224",
  },
})
if err := c.Connect(); err != nil {
  // ...
}
defer c.Disconnect()
```

### Create a TLS client

```go
c := client.New(client.ConnectionOptions{
  Factory: &client.ConnFactory{
    Address: "localhost:24224",
    TLSConfig: &tls.Config{InsecureSkipVerify: true},
  },
})
if err := c.Connect(); err != nil {
  // ...
}
defer c.Disconnect()
```

### Send a new log message

The `record` object must be a `map` or `struct`. Objects that implement the [`msgp.Encodable`](https://pkg.go.dev/github.com/tinylib/msgp/msgp#Encodable) interface will the be most performant.

```go
record := map[string]interface{}{
  "Hello": "World",
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

### Message confirmation

The client supports `ack` confirmations as specified by the Fluent protocol. When enabled, `SendMessage` returns once the acknowledgement is received or the timeout is reached.

Note: For types other than `RawMessage`, the `SendMessage` function sets the "chunk" option before sending. A `RawMessage` is immutable and must already contain a "chunk" value. The behavior is otherwise identical.

```go
c.RequireAck = true
if err := c.SendMessage(myMsg); err != nil {
  // ...
}
```

## Performance

**tl;dr** `fluent-forward-go` is fast and memory efficient. In some cases it is **70% faster** than the official package.

You can read more about the benchmarks [here](cmd/bm/README.md).

### SendMessage

Run on `localhost`. Does not include message creation.

```shell
Benchmark_Fluent_Forward_Go_SendOnly-16      10000      10847 ns/op      0 B/op      0 allocs/op
```

### Comparisons with `fluent-logger-golang`

The benchmarks below compare `fluent-forward-go` with the official package, [`fluent-logger-golang`](https://github.com/fluent/fluent-logger-golang). The message is a simple map with twelve keys.

The differences in execution times can vary from one test run to another. The differences in memory allocations, however, are constant.

#### Send a single message

```shell
Benchmark_Fluent_Forward_Go_SingleMessage-16    	   10000	     11355 ns/op	      48 B/op	       1 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessage-16      10000	     19687 ns/op	    2169 B/op	      33 allocs/op
```

#### Send single message with confirmation

```shell
Benchmark_Fluent_Forward_Go_SingleMessageAck-16       10000	    768743 ns/op	     185 B/op	       6 allocs/op
Benchmark_Fluent_Logger_Golang_SingleMessageAck-16    10000	    793360 ns/op	    6015 B/op	      47 allocs/op
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
