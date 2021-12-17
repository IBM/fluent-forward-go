package main

//go test -benchmem -benchtime=100x -run=^$ -bench ^Benchmark.*$ github.com/IBM/fluent-forward-go/cmd/bm/fluent_forward_go

import (
	"testing"
	"time"

	"github.com/IBM/fluent-forward-go/fluent/client"
	"github.com/IBM/fluent-forward-go/fluent/protocol"
)

func Benchmark_Fluent_Forward_Go_SendOnly(b *testing.B) {
	tagVar := "bar"

	c := client.New(client.ConnectionOptions{
		ConnectionTimeout: 3 * time.Second,
	})

	err := c.Connect()
	if err != nil {
		b.Fatal(err)
	}

	defer c.Disconnect()

	record := map[string]interface{}{
		"first": "Sir",
		"last":  "Gawain",
		"enemy": "Green Knight",
	}
	mne := protocol.NewMessage(tagVar, record)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = c.SendMessage(mne)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Fluent_Forward_Go_SingleMessage(b *testing.B) {
	tagVar := "bar"

	c := client.New(client.ConnectionOptions{
		ConnectionTimeout: 3 * time.Second,
	})

	err := c.Connect()
	if err != nil {
		b.Fatal(err)
	}

	defer c.Disconnect()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		record := map[string]interface{}{
			"first": "Sir",
			"last":  "Gawain",
			"enemy": "Green Knight",
		}
		mne := protocol.NewMessage(tagVar, record)
		err = c.SendMessage(mne)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Fluent_Forward_Go_SingleMessageAck(b *testing.B) {
	tagVar := "foo"

	c := client.New(client.ConnectionOptions{
		RequireAck:        true,
		ConnectionTimeout: 3 * time.Second,
	})

	err := c.Connect()
	if err != nil {
		b.Fatal(err)
	}

	defer c.Disconnect()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		record := map[string]interface{}{
			"first": "Sir",
			"last":  "Gawain",
			"enemy": "Green Knight",
		}
		mne := protocol.NewMessage(tagVar, record)
		err = c.SendMessage(mne)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Fluent_Forward_Go_Bytes(b *testing.B) {
	tagVar := "foo"

	c := client.New(client.ConnectionOptions{
		ConnectionTimeout: 3 * time.Second,
	})

	err := c.Connect()
	if err != nil {
		b.Fatal(err)
	}

	defer c.Disconnect()

	record := map[string]interface{}{
		"first": "Sir",
		"last":  "Gawain",
		"enemy": "Green Knight",
	}
	mne := protocol.NewMessage(tagVar, record)

	bits, _ := mne.MarshalMsg(nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = c.SendRaw(bits)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Fluent_Forward_Go_BytesAck(b *testing.B) {
	tagVar := "foo"

	c := client.New(client.ConnectionOptions{
		RequireAck:        true,
		ConnectionTimeout: 3 * time.Second,
	})

	err := c.Connect()
	if err != nil {
		b.Fatal(err)
	}

	defer c.Disconnect()

	record := map[string]interface{}{
		"first": "Sir",
		"last":  "Gawain",
		"enemy": "Green Knight",
	}
	mne := protocol.NewMessage(tagVar, record)

	mne.Chunk()
	bits, _ := mne.MarshalMsg(nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = c.SendRaw(bits)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Fluent_Forward_Go_RawMessage(b *testing.B) {
	tagVar := "foo"

	c := client.New(client.ConnectionOptions{
		ConnectionTimeout: 3 * time.Second,
	})

	err := c.Connect()
	if err != nil {
		b.Fatal(err)
	}

	defer c.Disconnect()

	record := map[string]interface{}{
		"first": "Sir",
		"last":  "Gawain",
		"enemy": "Green Knight",
	}
	mne := protocol.NewMessage(tagVar, record)

	bits, _ := mne.MarshalMsg(nil)
	rbits := protocol.RawMessage(bits)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = c.SendMessage(rbits)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Fluent_Forward_Go_RawMessageAck(b *testing.B) {
	tagVar := "foo"

	c := client.New(client.ConnectionOptions{
		RequireAck:        true,
		ConnectionTimeout: 3 * time.Second,
	})

	err := c.Connect()
	if err != nil {
		b.Fatal(err)
	}

	defer c.Disconnect()

	record := map[string]interface{}{
		"first": "Sir",
		"last":  "Gawain",
		"enemy": "Green Knight",
	}
	mne := protocol.NewMessage(tagVar, record)

	mne.Chunk()
	bits, _ := mne.MarshalMsg(nil)
	rbits := protocol.RawMessage(bits)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = c.SendMessage(rbits)
		if err != nil {
			b.Fatal(err)
		}
	}
}
