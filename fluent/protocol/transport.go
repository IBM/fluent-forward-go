package protocol

import (
	"encoding/binary"
	"errors"
	"time"

	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp

// =========
// TRANSPORT
// =========

const (
	OPT_SIZE       string = "size"
	OPT_CHUNK      string = "chunk"
	OPT_COMPRESSED string = "compressed"
	OPT_VAL_GZIP   string = "gzip"

	extensionType int8 = 0
	eventTimeLen  int  = 10
)

func init() {
	msgp.RegisterExtension(extensionType, func() msgp.Extension {
		return new(EventTime)
	})
}

// EventTime is the fluent-forward representation of a timestamp
type EventTime struct {
	time.Time
}

func (et *EventTime) ExtensionType() int8 {
	return extensionType
}

func (et *EventTime) Len() int {
	return eventTimeLen
}

// MarshalBinaryTo implements the Extension interface for marshaling an
// EventTime into a byte slice.
func (et *EventTime) MarshalBinaryTo(b []byte) error {
	utc := et.UTC()

	b[0] = 0xD7
	b[1] = 0x00
	binary.BigEndian.PutUint32(b[2:], uint32(utc.Unix()))
	binary.BigEndian.PutUint32(b[6:], uint32(utc.Nanosecond()))

	return nil
}

// UnmarshalBinary implements the Extension interface for unmarshaling
// into an EventTime object.
func (et *EventTime) UnmarshalBinary(timeBytes []byte) error {
	if len(timeBytes) != eventTimeLen {
		return errors.New("Invalid length")
	}

	seconds := binary.BigEndian.Uint32(timeBytes[2:])
	nanoseconds := binary.BigEndian.Uint32(timeBytes[6:])

	et.Time = time.Unix(int64(seconds), int64(nanoseconds))
	return nil
}

// Entry is the basic representation of an individual event.
//msgp:tuple Entry
type Entry struct {
	// Timestamp can contain the timestamp in either seconds or nanoseconds
	Timestamp EventTime `msg:"eventTime,extension"`
	// Record is the actual event record - key-value pairs, keys are strings.
	// At this point, we're using strings for the values as well, but we may
	// want to make those interface{} or something more generic at some point.
	Record map[string]string
}

// Message is used to send a single event at a time
//msgp:tuple Message
type Message struct {
	// Tag is a dot-delimited string used to categorize events
	Tag string
	Entry
	// Options - used to control server behavior.  Same as above, may need to
	// switch to interface{} or similar at some point.
	Options map[string]string
}

// ForwardMessage is used in Forward mode to send multiple events in a single
// msgpack array within a single request.
//msgp:tuple ForwardMessage
type ForwardMessage struct {
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
//msgp:tuple PackedForwardMessage
type PackedForwardMessage struct {
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
//msgp:tuple CompressedPackedForwardMessage
type CompressedPackedForwardMessage struct {
	PackedForwardMessage
}

func NewCompressedPackedForwardMessage(
	tag string, entries []Entry, opts map[string]string,
) *CompressedPackedForwardMessage {
	return nil
}
