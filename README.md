# fluent-forward-go
Golang client-side implementation of the fluent Forward protocol

This library implements the [fluent Forward protocol v1](https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1) in golang.  It provides only a
client-side implementation.

To test against fluent-bit, start up fluent-bit in a docker container with:
```
$ docker pull fluent/fluent-bit:1.8.2
$ docker run -p 127.0.0.1:24224:24224/tcp -v `pwd`:`pwd` -w `pwd` -ti fluent/fluent-bit:1.8.2 /fluent-bit/bin/fluent-bit -c $(pwd)/fixtures/fluent.conf
```

You can then build and run `cmd/send_test_msg` with:
```
$ (cd cmd; go build -o send_test_msg)
$ cmd/send_test_msg -t foo.bar
```
This will send two regular `Message`s, one with the timestamp as seconds since
the epoch, the other with the timestamp as seconds.nanoseconds.  Those will
then be written to a file with the same name as the tag supplied as the argument
to the `-t` flag (`foo.bar` in the above example).

It will also send a `ForwardMessage` containing a pair of events - these will be
written to the same file.

It will then send a `PackedForwardMessage` containing a pair of events - these
will be written to `$TAG.packed`.

Last, it will send a `CompressedPackedForwardMessage` with the same pair of events, which should then be written to `$TAG.compressed`.  Unfortunately, this
message mode is not yet working correctly (but we're working on that).
