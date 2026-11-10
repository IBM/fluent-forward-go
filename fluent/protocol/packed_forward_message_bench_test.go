package protocol_test

import (
	"testing"
	"time"

	. "github.com/IBM/fluent-forward-go/fluent/protocol"
)

func benchEntries(n int) EntryList {
	el := make(EntryList, n)
	for i := range el {
		el[i] = EntryExt{
			Timestamp: EventTime{Time: time.Unix(1_441_588_984, 0)},
			Record: map[string]interface{}{
				"message": "hello world, this is a benchmark log line",
				"value":   42,
			},
		}
	}

	return el
}

func BenchmarkMarshalPacked(b *testing.B) {
	for _, tc := range []struct {
		name string
		n    int
	}{{"1", 1}, {"100", 100}} {
		el := benchEntries(tc.n)

		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if _, err := el.MarshalPacked(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkNewPackedForwardMessage(b *testing.B) {
	el := benchEntries(100)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := NewPackedForwardMessage("foo", el); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewCompressedPackedForwardMessage(b *testing.B) {
	el := benchEntries(100)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := NewCompressedPackedForwardMessage("foo", el); err != nil {
			b.Fatal(err)
		}
	}
}
