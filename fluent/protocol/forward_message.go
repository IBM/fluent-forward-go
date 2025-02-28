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
	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp

// ForwardMessage is used in Forward mode to send multiple events in a single
// msgpack array within a single request.
//
//msgp:tuple ForwardMessage
//mgsp:test ignore ForwardMessage
//msgp:encode ignore ForwardMessage
//msgp:decode ignore ForwardMessage
//msgp:marshal ignore ForwardMessage
//msgp:unmarshal ignore ForwardMessage
//msgp:size ignore ForwardMessage
//msgp:test ignore ForwardMessage
type ForwardMessage struct {
	// Tag is a dot-delimted string used to categorize events
	Tag string
	// Entries is the set of event objects to be carried in this message
	Entries EntryList
	// Options - used to control server behavior.  Same as above, may need to
	// switch to interface{} or similar at some point.
	Options *MessageOptions
}

// NewForwardMessage creates a ForwardMessage from the supplied
// tag, EntryList, and MessageOptions. this function will set
// Options.Size to the length of the entry list.
func NewForwardMessage(
	tag string,
	entries EntryList,
) *ForwardMessage {
	lenEntries := len(entries)

	pfm := &ForwardMessage{
		Tag:     tag,
		Entries: entries,
	}

	pfm.Options = &MessageOptions{
		Size: &lenEntries,
	}

	return pfm
}

func (fm *ForwardMessage) EncodeMsg(dc *msgp.Writer) error {
	size := 2
	if fm.Options != nil {
		size = 3
	}

	err := dc.WriteArrayHeader(uint32(size)) //#nosec:G115
	if err != nil {
		return msgp.WrapError(err, "Array Header")
	}

	err = dc.WriteString(fm.Tag)
	if err != nil {
		return msgp.WrapError(err, "Tag")
	}

	err = fm.Entries.EncodeMsg(dc)
	if err != nil {
		return err
	}

	// if the options were included, inlcude them in our encoded message
	if size == 3 {
		err = fm.Options.EncodeMsg(dc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fm *ForwardMessage) DecodeMsg(dc *msgp.Reader) error {
	sz, err := dc.ReadArrayHeader()
	if err != nil {
		return msgp.WrapError(err, "Array Header")
	}

	if fm.Tag, err = dc.ReadString(); err != nil {
		return msgp.WrapError(err, "Tag")
	}

	fm.Entries = EntryList{}
	if err = fm.Entries.DecodeMsg(dc); err != nil {
		return msgp.WrapError(err, "Entries")
	}

	// has three elements only when options are included
	if sz == 3 {
		if t, err := dc.NextType(); t == msgp.NilType || err != nil {
			if err != nil {
				return msgp.WrapError(err, "Options")
			}

			return dc.ReadNil()
		}

		fm.Options = &MessageOptions{}
		if err = fm.Options.DecodeMsg(dc); err != nil {
			return msgp.WrapError(err, "Options")
		}
	}

	return nil
}

func (fm *ForwardMessage) MarshalMsg(bits []byte) ([]byte, error) {
	var (
		sz  uint32
		err error
	)

	if fm.Options != nil {
		sz = 3
	} else {
		sz = 2
	}

	bits = msgp.AppendArrayHeader(bits, sz)
	bits = msgp.AppendString(bits, fm.Tag)

	bits, err = fm.Entries.MarshalMsg(bits)
	if err != nil {
		return bits, err
	}

	if sz == 3 {
		bits, err = fm.Options.MarshalMsg(bits)
	}

	return bits, err
}

func (fm *ForwardMessage) UnmarshalMsg(bits []byte) ([]byte, error) {
	var (
		sz  uint32
		err error
	)

	if sz, bits, err = msgp.ReadArrayHeaderBytes(bits); err != nil {
		return bits, msgp.WrapError(err, "Array Header")
	}

	if fm.Tag, bits, err = msgp.ReadStringBytes(bits); err != nil {
		return bits, msgp.WrapError(err, "Tag")
	}

	fm.Entries = EntryList{}
	if bits, err = fm.Entries.UnmarshalMsg(bits); err != nil {
		return bits, err
	}

	// has three elements only when options are included
	if sz == 3 {
		if t := msgp.NextType(bits); t == msgp.NilType {
			return msgp.ReadNilBytes(bits)
		}

		fm.Options = &MessageOptions{}
		if bits, err = fm.Options.UnmarshalMsg(bits); err != nil {
			return bits, msgp.WrapError(err, "Options")
		}
	}

	return bits, err
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (fm *ForwardMessage) Msgsize() (s int) {
	s = 1 + msgp.StringPrefixSize + len(fm.Tag) + fm.Entries.Msgsize()
	if fm.Options != nil {
		s += fm.Options.Msgsize()
	}

	return
}

func (fm *ForwardMessage) Chunk() (string, error) {
	if fm.Options == nil {
		fm.Options = &MessageOptions{}
	}

	if fm.Options.Chunk != "" {
		return fm.Options.Chunk, nil
	}

	chunk, err := makeChunkID()
	fm.Options.Chunk = chunk

	return chunk, err
}
