package protocol

import (
	"bytes"

	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp

func eventStream(entries EntryList) []byte {
	var buf bytes.Buffer
	w := msgp.NewWriter(&buf)
	for _, e := range entries {
		// TODO: capture and return error
		_ = e.EncodeMsg(w)
	}
	w.Flush()

	return buf.Bytes()
}

// PackedForwardMessage is just like ForwardMessage, except that the events
// are carried as a msgpack binary stream
//msgp:tuple PackedForwardMessage
//msgp:decode ignore PackedForwardMessage
//msgp:unmarshal ignore PackedForwardMessage
//msgp:size ignore PackedForwardMessage
//msgp:test ignore PackedForwardMessage
type PackedForwardMessage struct {
	// Tag is a dot-delimited string used to categorize events
	Tag string
	// EventStream is the set of events (entries in Fluent-speak) serialized
	// into a msgpack byte stream
	EventStream []byte
	// Options - used to control server behavior.  Same as above, may need to
	// switch to interface{} or similar at some point.
	Options *MessageOptions
}

// NewPackedForwardMessage creates a PackedForwardMessage from the supplied
// tag, EntryList, and MessageOptions.  Regardless of the options supplied,
// this function will set opts[OPT_SIZE] to the length of the entry list.
func NewPackedForwardMessage(
	tag string,
	entries EntryList,
	opts *MessageOptions,
) *PackedForwardMessage {
	// set the options size to be the number of entries
	opts.Size = len(entries)

	return NewPackedForwardMessageFromBytes(tag, eventStream(entries), opts)
}

// NewPackedForwardMessageFromBytes creates a PackedForwardMessage from the
// supplied tag, bytes, and MessageOptions. This function does not set
// opts[OPT_SIZE] to the length of the entry list.
func NewPackedForwardMessageFromBytes(
	tag string,
	entries []byte,
	opts *MessageOptions,
) *PackedForwardMessage {
	return &PackedForwardMessage{
		Tag:         tag,
		EventStream: entries,
		Options:     opts,
	}
}

func (msg *PackedForwardMessage) DecodeMsg(dc *msgp.Reader) error {
	sz, err := dc.ReadArrayHeader()
	if err != nil {
		return msgp.WrapError(err, "Array Header")
	}

	if msg.Tag, err = dc.ReadString(); err != nil {
		return msgp.WrapError(err, "Tag")
	}

	if msg.EventStream, err = dc.ReadBytes(msg.EventStream); err != nil {
		return msgp.WrapError(err, "EventStream")
	}

	// has three elements only when options are included
	if sz == 3 {
		msg.Options = &MessageOptions{}
		if err = msg.Options.DecodeMsg(dc); err != nil {
			return msgp.WrapError(err, "Options")
		}
	}

	return nil
}

func (msg *PackedForwardMessage) UnmarshalMsg(bits []byte) ([]byte, error) {
	var (
		sz  uint32
		err error
	)

	if sz, bits, err = msgp.ReadArrayHeaderBytes(bits); err != nil {
		return bits, msgp.WrapError(err, "Array Header")
	}

	if msg.Tag, bits, err = msgp.ReadStringBytes(bits); err != nil {
		return bits, msgp.WrapError(err, "Tag")
	}

	if msg.EventStream, bits, err = msgp.ReadBytesBytes(bits, msg.EventStream); err != nil {
		return bits, msgp.WrapError(err, "EventStream")
	}

	// has three elements only when options are included
	if sz == 3 {
		msg.Options = &MessageOptions{}
		if bits, err = msg.Options.UnmarshalMsg(bits); err != nil {
			return bits, msgp.WrapError(err, "Options")
		}
	}

	return bits, err
}

func (msg *PackedForwardMessage) Msgsize() (s int) {
	s = 1 + msgp.StringPrefixSize + len(msg.Tag) + msgp.BytesPrefixSize + len(msg.EventStream)
	if msg.Options != nil {
		s += msg.Options.Msgsize()
	}
	return
}
