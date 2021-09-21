package protocol

import (
	"bytes"
	"compress/gzip"

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
) *PackedForwardMessage {
	pfm := NewPackedForwardMessageFromBytes(tag, eventStream(entries))
	pfm.Options = &MessageOptions{
		Size: len(entries),
	}

	return pfm
}

// NewPackedForwardMessageFromBytes creates a PackedForwardMessage from the
// supplied tag, bytes, and MessageOptions. This function does not set
// opts[OPT_SIZE] to the length of the entry list.
func NewPackedForwardMessageFromBytes(
	tag string,
	entries []byte,
) *PackedForwardMessage {
	return &PackedForwardMessage{
		Tag:         tag,
		EventStream: entries,
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
		if t, _ := dc.NextType(); t == msgp.NilType {
			return dc.ReadNil()
		}

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
		if t := msgp.NextType(bits); t == msgp.NilType {
			return msgp.ReadNilBytes(bits)
		}

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

//msgp:ignore GzipCompressor
type GzipCompressor struct {
	Buffer     *bytes.Buffer
	GzipWriter *gzip.Writer
}

func NewGzipCompressor() *GzipCompressor {
	buf := new(bytes.Buffer)
	zw := gzip.NewWriter(buf)

	return &GzipCompressor{buf, zw}
}

// Write writes to the compression stream.
func (mc *GzipCompressor) Write(bits []byte) error {
	_, err := mc.GzipWriter.Write(bits)

	if cerr := mc.GzipWriter.Close(); err == nil {
		err = cerr
	}

	return err
}

// Reset resets the buffer to be empty, but it retains the
// underlying storage for use by future writes.
func (mc *GzipCompressor) Reset() {
	mc.Buffer.Reset()
	mc.GzipWriter.Reset(mc.Buffer)
}

// Bytes returns the gzip-compressed byte stream.
func (mc *GzipCompressor) Bytes() []byte {
	return mc.Buffer.Bytes()
}

// NewCompressedPackedForwardMessage returns a PackedForwardMessage with a gzip-compressed byte stream.
func NewCompressedPackedForwardMessage(
	tag string, entries []EntryExt,
) (*PackedForwardMessage, error) {
	return newCompressedPackedForwardMessageFromBytes(tag, eventStream(entries), len(entries))
}

// NewCompressedPackedForwardMessageFromBytes returns a PackedForwardMessage with a gzip-compressed byte stream.
func NewCompressedPackedForwardMessageFromBytes(
	tag string, entries []byte,
) (*PackedForwardMessage, error) {
	return newCompressedPackedForwardMessageFromBytes(tag, entries, 0)
}

func newCompressedPackedForwardMessageFromBytes(
	tag string, entries []byte, sz int,
) (*PackedForwardMessage, error) {
	mc := compressorPool.Get().(*GzipCompressor)
	defer func() {
		mc.Reset()
		compressorPool.Put(mc)
	}()

	if err := mc.Write(entries); err != nil {
		return nil, err
	}

	pfm := NewPackedForwardMessageFromBytes(tag, mc.Bytes())
	pfm.Options = &MessageOptions{
		Size: sz, Compressed: "gzip",
	}

	return pfm, nil
}
