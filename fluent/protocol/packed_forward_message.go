/*
MIT License

Copyright contributors to the fluent-forward-go project

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package protocol

import (
	"bytes"
	"compress/gzip"

	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp

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
// this function will set Options.Size to the length of the entry list.
func NewPackedForwardMessage(
	tag string,
	entries EntryList,
) (*PackedForwardMessage, error) {
	el := EntryList(entries) //nolint

	bits, err := el.MarshalPacked()
	if err != nil {
		return nil, err
	}

	lenEntries := len(entries)

	pfm := NewPackedForwardMessageFromBytes(tag, bits)
	pfm.Options = &MessageOptions{
		Size: &lenEntries,
	}

	return pfm, nil
}

// NewPackedForwardMessageFromBytes creates a PackedForwardMessage from the
// supplied tag, bytes, and MessageOptions. This function does not set
// Options.Size to the length of the entry list.
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

func (msg *PackedForwardMessage) Chunk() (string, error) {
	if msg.Options == nil {
		msg.Options = &MessageOptions{}
	}

	if msg.Options.Chunk != "" {
		return msg.Options.Chunk, nil
	}

	chunk, err := makeChunkID()
	msg.Options.Chunk = chunk

	return chunk, err
}

//msgp:ignore GzipCompressor
type GzipCompressor struct {
	Buffer     *bytes.Buffer
	GzipWriter *gzip.Writer
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
	if mc.Buffer == nil {
		mc.Buffer = new(bytes.Buffer)
		mc.GzipWriter = gzip.NewWriter(mc.Buffer)

		return
	}

	mc.Buffer.Reset()
	mc.GzipWriter.Reset(mc.Buffer)
}

// Bytes returns the gzip-compressed byte stream.
func (mc *GzipCompressor) Bytes() []byte {
	return mc.Buffer.Bytes()
}

// NewCompressedPackedForwardMessage returns a PackedForwardMessage with a
// gzip-compressed byte stream.
func NewCompressedPackedForwardMessage(
	tag string, entries []EntryExt,
) (*PackedForwardMessage, error) {
	el := EntryList(entries) //nolint

	bits, err := el.MarshalPacked()
	if err != nil {
		return nil, err
	}

	lenEntries := len(entries)

	msg, err := NewCompressedPackedForwardMessageFromBytes(tag, bits)
	if err == nil {
		msg.Options.Size = &lenEntries
	}

	return msg, err
}

// NewCompressedPackedForwardMessageFromBytes returns a PackedForwardMessage
// with a gzip-compressed byte stream.
func NewCompressedPackedForwardMessageFromBytes(
	tag string, entries []byte,
) (*PackedForwardMessage, error) {
	mc := compressorPool.Get().(*GzipCompressor)
	mc.Reset()

	defer func() {
		compressorPool.Put(mc)
	}()

	if err := mc.Write(entries); err != nil {
		return nil, err
	}

	pfm := NewPackedForwardMessageFromBytes(tag, mc.Bytes())
	pfm.Options = &MessageOptions{Compressed: "gzip"}

	return pfm, nil
}
