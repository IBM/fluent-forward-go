package protocol

import (
	"encoding/binary"
	"errors"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

// =========
// TRANSPORT
// =========

func init() {
	msgpack.RegisterExt(0, (*EntryTime)(nil))
}

const (
	OPT_SIZE       string = "size"
	OPT_CHUNK      string = "chunk"
	OPT_COMPRESSED string = "compressed"
	OPT_VAL_GZIP   string = "gzip"
)

var (
	_ msgpack.Marshaler   = (*EntryTime)(nil)
	_ msgpack.Unmarshaler = (*EntryTime)(nil)
)

type EntryTime struct {
	time.Time
}

func (et *EntryTime) MarshalMsgpack() ([]byte, error) {
	timeBytes := make([]byte, 8)

	binary.BigEndian.PutUint32(timeBytes, uint32(et.Unix()))
	binary.BigEndian.PutUint32(timeBytes[4:], uint32(et.Nanosecond()))

	return timeBytes, nil
}

func (et *EntryTime) UnmarshalMsgpack(timeBytes []byte) error {
	if len(timeBytes) != 8 {
		return errors.New("Invalid length")
	}

	seconds := binary.BigEndian.Uint32(timeBytes)
	nanoseconds := binary.BigEndian.Uint32(timeBytes[4:])

	et.Time = time.Unix(int64(seconds), int64(nanoseconds))
	return nil
}

//

// Entry is the basic representation of an individual event.
type Entry struct {
	_msgpack struct{} `msgpack:",as_array"`
	// Timestamp can contain the timestamp in either seconds or nanoseconds
	TimeStamp int64
	// Record is the actual event record - key-value pairs, keys are strings.
	// At this point, we're using strings for the values as well, but we may
	// want to make those interface{} or something more generic at some point.
	Record map[string]string
}

// Message is used to send a single event at a time
type Message struct {
	_msgpack struct{} `msgpack:",as_array"`
	// Tag is a dot-delimited string used to categorize events
	Tag string
	Entry
	// Options - used to control server behavior.  Same as above, may need to
	// switch to interface{} or similar at some point.
	Options map[string]string
}

// ForwardMessage is used in Forward mode to send multiple events in a single
// msgpack array within a single request.
type ForwardMessage struct {
	_msgpack struct{} `msgpack:",as_array"`
	// Tag is a dot-delimted string used to categorize events
	Tag string
	// Entries is the set of event objects to be carried in this message
	Entries []Entry
	// Options - used to control server behavior.  Same as above, may need to
	// switch to interface{} or similar at some point.
	Options map[string]string
}

// PackedForwardMessage is just like ForwardMessage, except that the events
// are carried as a msgpack binary stream
type PackedForwardMessage struct {
	_msgpack struct{} `msgpack:",as_array"`
	// Tag is a dot-delimited string used to categorize events
	Tag string
	// EventStream is the set of events (entries in Fluent-speak) serialized
	// into a msgpack byte stream
	EventStream []byte
	// Options - used to control server behavior.  Same as above, may need to
	// switch to interface{} or similar at some point.
	Options map[string]string
}

// CompressedPackedForwardMode is just like PackedForwardMessage, except that
// the msgpack byte stream containing the events/entries is compressed using
// gzip.  The protocol spec states that the event stream may be formed by
// concatenating multiple gzipped binary strings, but we do not claim to
// support that yet.
// Users of this type MUST set the option "compressed" => "gzip"
type CompressedPackedForwardMessage struct {
	PackedForwardMessage
}

func NewCompressedPackedForwardMessage(
	tag string, entries []Entry, opts map[string]string,
) *CompressedPackedForwardMessage {
	return nil
}
