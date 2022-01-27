package main

//go test -benchmem -run=^$ -bench ^Benchmark.*$ github.com/IBM/fluent-forward-go/cmd/bm/fluent_logger_golang

import (
	"testing"
	"time"

	"github.com/IBM/fluent-forward-go/cmd/bm"
	"github.com/fluent/fluent-logger-golang/fluent"
)

func Benchmark_Fluent_Logger_Golang_SingleMessage(b *testing.B) {
	logger, err := fluent.New(fluent.Config{
		SubSecondPrecision: false,
	})

	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	tag := "foo"
	data := bm.MakeRecord(12)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = logger.Post(tag, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Fluent_Logger_Golang_SingleMessageAck(b *testing.B) {
	logger, err := fluent.New(fluent.Config{
		Timeout:            3 * time.Second,
		RequestAck:         true,
		SubSecondPrecision: false,
	})

	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	tag := "foo"
	data := bm.MakeRecord(12)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = logger.Post(tag, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
