package main

import (
	"flag"
	"fmt"
	"os"
	"time"

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
	c := &client.Client{
		ConnectionFactory: &client.TCPConnectionFactory{
			Target: client.ServerAddress{
				Hostname: "localhost",
				Port:     24224,
			},
		},
	}

	err := c.Connect()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to connect, exiting", err)
		os.Exit(-1)
	}

	msg := protocol.Message{
		Tag:       tagVar,
		Timestamp: time.Now().UTC().Unix(),
		Record: map[string]interface{}{
			"first": "Sir",
			"last":  "Gawain",
			"enemy": "Green Knight",
		},
		Options: &protocol.MessageOptions{},
	}

	mne := protocol.MessageExt{
		Tag:       tagVar,
		Timestamp: protocol.EventTime{Time: time.Now().UTC()},
		Record: map[string]interface{}{
			"first": "Sir",
			"last":  "Gawain",
			"enemy": "Green Knight",
		},
		Options: &protocol.MessageOptions{},
	}

	fwd := protocol.ForwardMessage{
		Tag: tagVar,
		Entries: []protocol.EntryExt{
			{
				Timestamp: protocol.EventTime{Time: time.Now().UTC()},
				Record: map[string]interface{}{
					"first": "Edgar",
					"last":  "Winter",
					"enemy": "wimpy music",
				},
			},
			{
				Timestamp: protocol.EventTime{Time: time.Now().UTC()},
				Record: map[string]interface{}{
					"first": "George",
					"last":  "Clinton",
					"enemy": "Sir Nose D Voidoffunk",
				},
			},
		},
		Options: &protocol.MessageOptions{},
	}

	packedFwd := protocol.NewPackedForwardMessage(tagVar+".packed", fwd.Entries)

	compressed, _ := protocol.NewCompressedPackedForwardMessage(tagVar+".compressed",
		fwd.Entries)

	err = c.SendMessage(&msg)
	if err != nil {
		os.Exit(1)
	}

	err = c.SendMessage(&mne)
	if err != nil {
		os.Exit(1)
	}

	err = c.SendMessage(&fwd)
	if err != nil {
		os.Exit(1)
	}

	err = c.SendMessage(packedFwd)
	if err != nil {
		os.Exit(1)
	}

	err = c.SendMessage(compressed)
	if err != nil {
		os.Exit(1)
	}

	fmt.Println("Messages sent")

	os.Exit(0)
}
