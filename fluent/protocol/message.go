package protocol

import (
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
	Record    Record
	// Options - used to control server behavior.
	Options *MessageOptions
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

	msg.Record = Record{}
	if err = msg.Record.DecodeMsg(dc); err != nil {
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

	msg.Record = Record{}
	if bits, err = msg.Record.UnmarshalMsg(bits); err != nil {
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
	s = 1 + msgp.StringPrefixSize + len(msg.Tag) + msgp.Int64Size + msg.Record.Msgsize()
	if msg.Options != nil {
		s += msg.Options.Msgsize()
	}

	return
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
	Record    Record
	Options   *MessageOptions
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

	msg.Record = Record{}
	if err = msg.Record.DecodeMsg(dc); err != nil {
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

	msg.Record = Record{}
	if bits, err = msg.Record.UnmarshalMsg(bits); err != nil {
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
	s = 1 + msgp.StringPrefixSize + len(msg.Tag) + msgp.ExtensionPrefixSize + msg.Timestamp.Len() + msg.Record.Msgsize()
	if msg.Options != nil {
		s += msgp.NilSize
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
