package fluent_forward_go

//go test -benchmem -run=^$ -bench ^Benchmark.*$ github.com/IBM/fluent-forward-go/cmd/bm/fluent_logger_golang

import (
	"testing"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
)

func Benchmark_Fluent_Logger_Golang_SendOnly(b *testing.B) {
	logger, err := fluent.New(fluent.Config{
		SubSecondPrecision: true,
	})

	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tag := "foo"
		var data = map[string]string{
			"first": "Sir",
			"last":  "Gawain",
			"enemy": "Green Knight",
		}

		err = logger.Post(tag, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Fluent_Logger_Golang_SingleMessage(b *testing.B) {
	logger, err := fluent.New(fluent.Config{
		SubSecondPrecision: true,
	})

	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tag := "foo"
		var data = map[string]string{
			"first": "Sir",
			"last":  "Gawain",
			"enemy": "Green Knight",
		}

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
		SubSecondPrecision: true,
	})

	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()
	tag := "foo"
	var data = map[string]string{
		"first": "Sir",
		"last":  "Gawain",
		"enemy": "Green Knight",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = logger.Post(tag, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
