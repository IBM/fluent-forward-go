package protocol

import "github.com/tinylib/msgp/msgp"

// ForwardMessage is used in Forward mode to send multiple events in a single
// msgpack array within a single request.
//msgp:tuple ForwardMessage
//msgp:decode ignore ForwardMessage
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
