package protocol

import "github.com/tinylib/msgp/msgp"

//go:generate msgp

// ForwardMessage is used in Forward mode to send multiple events in a single
// msgpack array within a single request.
//msgp:tuple ForwardMessage
//mgsp:test ignore ForwardMessage
//msgp:encode ignore ForwardMessage
//msgp:decode ignore ForwardMessage
//msgp:marshal ignore ForwardMessage
//msgp:unmarshal ignore ForwardMessage
type ForwardMessage struct {
	// Tag is a dot-delimted string used to categorize events
	Tag string
	// Entries is the set of event objects to be carried in this message
	Entries EntryList
	// Options - used to control server behavior.  Same as above, may need to
	// switch to interface{} or similar at some point.
	Options *MessageOptions
}

func (fm *ForwardMessage) EncodeMsg(dc *msgp.Writer) error {
	size := 2
	if fm.Options != nil {
		size = 3
	}

	err := dc.WriteArrayHeader(uint32(size))
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
		return err
	}

	// has three elements only when options are included
	if sz == 3 {
		fm.Options = &MessageOptions{}
		if err = fm.Options.DecodeMsg(dc); err != nil {
			return err
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
		fm.Options = &MessageOptions{}
		if bits, err = fm.Options.UnmarshalMsg(bits); err != nil {
			return bits, err
		}
	}

	return bits, err
}
