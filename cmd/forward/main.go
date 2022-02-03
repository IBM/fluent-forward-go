package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/IBM/fluent-forward-go/fluent/client"
	"github.com/IBM/fluent-forward-go/fluent/protocol"
)

var (
	tagVar string
)

func init() {
	flag.StringVar(&tagVar, "tag", "test.message", "-tag <dot-delimited tag>")
	flag.StringVar(&tagVar, "t", "test.message", "-t <dot-delimited tag> (shorthand for -tag)")
}

func main() {
	flag.Parse()

	c := client.New(client.ConnectionOptions{
		RequireAck: true,
	})

	err := c.Connect()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to connect, exiting", err)
		os.Exit(-1)
	}

	record := map[string]interface{}{
		"first": "Sir",
		"last":  "Gawain",
		"enemy": "Green Knight",
		"equipment": []string{
			"sword",
			"lance",
			"full plate",
		},
	}

	entries := []protocol.EventExt{
		{
			Timestamp: protocol.EventTimeNow(),
			Record: map[string]interface{}{
				"first": "Edgar",
				"last":  "Winter",
				"enemy": "wimpy music",
			},
		},
		{
			Timestamp: protocol.EventTimeNow(),
			Record: map[string]interface{}{
				"first": "George",
				"last":  "Clinton",
				"enemy": "Sir Nose D Voidoffunk",
			},
		},
	}

	msg := protocol.NewMessage(tagVar, record)
	mne := protocol.NewMessageExt(tagVar, record)
	fwd := protocol.NewForwardMessage(tagVar, entries)
	packedFwd, _ := protocol.NewPackedForwardMessage(tagVar+".packed", entries)
	compressed, _ := protocol.NewCompressedPackedForwardMessage(tagVar+".compressed",
		fwd.Entries)

	err = c.Send(msg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = c.Send(mne)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = c.Send(fwd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = c.Send(packedFwd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = c.Send(compressed)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, _ = compressed.Chunk()
	b, _ := compressed.MarshalMsg(nil)
	rm := protocol.RawMessage(b)

	err = c.Send(rm)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Messages sent")

	os.Exit(0)
}
