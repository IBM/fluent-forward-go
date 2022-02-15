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
	"time"

	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp

// Message is used to send a single event at a time
//msgp:tuple Message
//msgp:decode ignore Message
//msgp:unmarshal ignore Message
//msgp:size ignore Message
//msgp:test ignore Message
type Message struct {
	// Tag is a dot-delimited string used to categorize events
	Tag       string
	Timestamp int64
	Record    interface{}
	// Options - used to control server behavior.
	Options *MessageOptions
}

// NewMessage creates a Message from the supplied
// tag and record. The record object must be a map or
// struct. Objects that implement the msgp.Encodable
// interface will be the most performant. Timestamp is
// set to time.Now().UTC() and marshaled with second
// precision.
func NewMessage(
	tag string,
	record interface{},
) *Message {
	msg := &Message{
		Tag:       tag,
		Timestamp: time.Now().UTC().Unix(),
		Record:    record,
	}

	return msg
}

func (msg *Message) DecodeMsg(dc *msgp.Reader) error {
	sz, err := dc.ReadArrayHeader()
	if err != nil {
		return msgp.WrapError(err, "Array Header")
	}

	if msg.Tag, err = dc.ReadString(); err != nil {
		return msgp.WrapError(err, "Tag")
	}

	if msg.Timestamp, err = dc.ReadInt64(); err != nil {
		return msgp.WrapError(err, "Timestamp")
	}

	if msg.Record, err = dc.ReadIntf(); err != nil {
		return msgp.WrapError(err, "Record")
	}

	// has four elements only when options are included
	if sz == 4 {
		if t, err := dc.NextType(); t == msgp.NilType || err != nil {
			if err != nil {
				return msgp.WrapError(err, "Options")
			}

			return dc.ReadNil()
		}

		msg.Options = &MessageOptions{}
		if err = msg.Options.DecodeMsg(dc); err != nil {
			return msgp.WrapError(err, "Options")
		}
	}

	return nil
}

func (msg *Message) UnmarshalMsg(bits []byte) ([]byte, error) {
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

	if msg.Timestamp, bits, err = msgp.ReadInt64Bytes(bits); err != nil {
		return bits, msgp.WrapError(err, "Timestamp")
	}

	if msg.Record, bits, err = msgp.ReadIntfBytes(bits); err != nil {
		return bits, msgp.WrapError(err, "Record")
	}

	// has four elements only when options are included
	if sz == 4 {
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

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (msg *Message) Msgsize() (s int) {
	s = 1 + msgp.StringPrefixSize + len(msg.Tag) + msgp.Int64Size + msgp.GuessSize(msg.Record)
	if msg.Options != nil {
		s += msg.Options.Msgsize()
	}

	return s
}

func (msg *Message) Chunk() (string, error) {
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

// MessageExt
//msgp:tuple MessageExt
//msgp:decode ignore MessageExt
//msgp:unmarshal ignore MessageExt
//msgp:size ignore MessageExt
//msgp:test ignore MessageExt
type MessageExt struct {
	Tag       string
	Timestamp EventTime `msg:"eventTime,extension"`
	Record    interface{}
	Options   *MessageOptions
}

// NewMessageExt creates a MessageExt from the supplied
// tag and record. The record object must be a map or
// struct. Objects that implement the msgp.Encodable
// interface will be the most performant. Timestamp is
// set to time.Now().UTC() and marshaled with subsecond
// precision.
func NewMessageExt(
	tag string,
	record interface{},
) *MessageExt {
	msg := &MessageExt{
		Tag:       tag,
		Timestamp: EventTimeNow(),
		Record:    record,
	}

	return msg
}

func (msg *MessageExt) DecodeMsg(dc *msgp.Reader) error {
	sz, err := dc.ReadArrayHeader()
	if err != nil {
		return msgp.WrapError(err, "Array Header")
	}

	if msg.Tag, err = dc.ReadString(); err != nil {
		return msgp.WrapError(err, "Tag")
	}

	if err = dc.ReadExtension(&msg.Timestamp); err != nil {
		return msgp.WrapError(err, "Timestamp")
	}

	if msg.Record, err = dc.ReadIntf(); err != nil {
		return msgp.WrapError(err, "Record")
	}

	// has four elements only when options are included
	if sz == 4 {
		if t, err := dc.NextType(); t == msgp.NilType || err != nil {
			if err != nil {
				return msgp.WrapError(err, "Options")
			}

			return dc.ReadNil()
		}

		msg.Options = &MessageOptions{}
		if err = msg.Options.DecodeMsg(dc); err != nil {
			return msgp.WrapError(err, "Options")
		}
	}

	return nil
}

func (msg *MessageExt) UnmarshalMsg(bits []byte) ([]byte, error) {
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

	if bits, err = msgp.ReadExtensionBytes(bits, &msg.Timestamp); err != nil {
		return bits, msgp.WrapError(err, "Timestamp")
	}

	if msg.Record, bits, err = msgp.ReadIntfBytes(bits); err != nil {
		return bits, msgp.WrapError(err, "Record")
	}

	// has four elements only when options are included
	if sz == 4 {
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

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (msg *MessageExt) Msgsize() (s int) {
	s = 1 + msgp.StringPrefixSize + len(msg.Tag) + msgp.ExtensionPrefixSize + msg.Timestamp.Len() + msgp.GuessSize(msg.Record)
	if msg.Options != nil {
		s += msg.Options.Msgsize()
	}

	return
}

func (msg *MessageExt) Chunk() (string, error) {
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
