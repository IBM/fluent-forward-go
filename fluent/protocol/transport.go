package protocol

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"strconv"
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
	eventTimeLen  int  = 8
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

	// b[0] = 0xD7
	// b[1] = 0x00
	binary.BigEndian.PutUint32(b, uint32(utc.Unix()))
	binary.BigEndian.PutUint32(b[4:], uint32(utc.Nanosecond()))

	return nil
}

// UnmarshalBinary implements the Extension interface for unmarshaling
// into an EventTime object.
func (et *EventTime) UnmarshalBinary(timeBytes []byte) error {
	if len(timeBytes) != eventTimeLen {
		return errors.New("Invalid length")
	}

	seconds := binary.BigEndian.Uint32(timeBytes)
	nanoseconds := binary.BigEndian.Uint32(timeBytes[4:])

	et.Time = time.Unix(int64(seconds), int64(nanoseconds))
	return nil
}

// EntryExt is the basic representation of an individual event, but using the
// msgpack extension format for the timestamp.
//msgp:tuple EntryExt
type EntryExt struct {
	// Timestamp can contain the timestamp in either seconds or nanoseconds
	Timestamp EventTime `msg:"eventTime,extension"`
	// Record is the actual event record - key-value pairs, keys are strings.
	// At this point, we're using strings for the values as well, but we may
	// want to make those interface{} or something more generic at some point.
	Record map[string]string
}

// EntryExt is the basic representation of an individual event.  The timestamp
// is an int64 representing seconds since the epoch (UTC).  The initial creator
// of the entry is responsible for converting to UTC.
//msgp:tuple Entry
type Entry struct {
	// Timestamp can contain the timestamp in either seconds or nanoseconds
	Timestamp int64
	// Record is the actual event record - key-value pairs, keys are strings.
	// At this point, we're using strings for the values as well, but we may
	// want to make those interface{} or something more generic at some point.
	Record map[string]string
}

type MessageOptions map[string]string

// Message is used to send a single event at a time
//msgp:tuple Message
type Message struct {
	// Tag is a dot-delimited string used to categorize events
	Tag       string
	Timestamp int64
	Record    map[string]string
	// Options - used to control server behavior.  Same as above, may need to
	// switch to interface{} or similar at some point.
	Options MessageOptions
}

// MessageExt
//msgp:tuple MessageExt
type MessageExt struct {
	Tag       string
	Timestamp EventTime `msg:"eventTime,extension"`
	Record    map[string]string
	Options   MessageOptions
}

// ForwardMessage is used in Forward mode to send multiple events in a single
// msgpack array within a single request.
//msgp:tuple ForwardMessage
type ForwardMessage struct {
	// Tag is a dot-delimted string used to categorize events
	Tag string
	// Entries is the set of event objects to be carried in this message
	Entries []EntryExt
	// Options - used to control server behavior.  Same as above, may need to
	// switch to interface{} or similar at some point.
	Options MessageOptions
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
	Options MessageOptions
}

type EntryList []EntryExt

// Equal compares two EntryList objects and returns true if they have
// exactly the same elements, false otherwise.
func (e EntryList) Equal(e2 EntryList) bool {
	if len(e) != len(e2) {
		return false
	}

	first := make(EntryList, len(e))
	copy(first, e)
	second := make(EntryList, len(e2))
	copy(second, e2)

	matches := 0
	for _, ea := range first {
		for _, eb := range second {
			if ea.Timestamp.Equal(eb.Timestamp.Time) {
				// Timestamps equal, check the record
				if len(ea.Record) == len(eb.Record) {
					// This only works if we have the same number of kv pairs in each record
					for k, v := range ea.Record {
						if eb.Record[k] == v {
							// KV match, so delete key from each record
							delete(ea.Record, k)
							delete(eb.Record, k)
						}
					}
					if len(ea.Record) == 0 && len(eb.Record) == 0 {
						// No more keys left means everything matched
						matches++
					}
				}
			}
		}
	}
	if matches == len(e) {
		return true
	}
	return false
}

// NewPackedForwardMessage creates a PackedForwardMessage from the supplied
// tag, EntryList, and MessageOptions.  Regardless of the options supplied,
// this function will set opts[OPT_SIZE] to the length of the entry list.
func NewPackedForwardMessage(
	tag string,
	entries EntryList,
	opts MessageOptions,
) *PackedForwardMessage {

	msg := &PackedForwardMessage{
		Tag:         tag,
		EventStream: eventStream(entries),
		Options: MessageOptions{
			OPT_SIZE: strconv.Itoa(len(entries)),
		},
	}
	return msg
}

func eventStream(entries EntryList) []byte {
	var buf bytes.Buffer
	w := msgp.NewWriter(&buf)
	for _, e := range entries {
		e.EncodeMsg(w)
	}
	w.Flush()

	return buf.Bytes()
}

// CompressedPackedForwardMode is just like PackedForwardMessage, except that
// the msgpack byte stream containing the events/entries is compressed using
// gzip.  The protocol spec states that the event stream may be formed by
// concatenating multiple gzipped binary strings, but we do not claim to
// support that yet.
// Users of this type MUST set the option "compressed" => "gzip"
//msgp:tuple CompressedPackedForwardMessage
type CompressedPackedForwardMessage struct {
	Tag                   string
	CompressedEventStream []byte
	Options               MessageOptions
}

// TODO: This is not working correctly yet
func NewCompressedPackedForwardMessage(
	tag string, entries []EntryExt, opts MessageOptions,
) *CompressedPackedForwardMessage {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.Write(eventStream(entries))
	zw.Close()
	// TODO: Do something real here.
	return &CompressedPackedForwardMessage{
		Tag:                   tag,
		CompressedEventStream: buf.Bytes(),
		Options: MessageOptions{
			OPT_SIZE:       strconv.Itoa(len(entries)),
			OPT_COMPRESSED: OPT_VAL_GZIP,
		},
	}
}
