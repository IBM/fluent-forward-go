package protocol

import (
	"bytes"
	"compress/gzip"
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

type Record map[string]interface{}

// EntryExt is the basic representation of an individual event, but using the
// msgpack extension format for the timestamp.
//msgp:tuple EntryExt
type EntryExt struct {
	// Timestamp can contain the timestamp in either seconds or nanoseconds
	Timestamp EventTime `msg:"eventTime,extension"`
	// Record is the actual event record - key-value pairs, keys are strings.
	Record Record
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

	return matches == len(e)
}

// EntryExt is the basic representation of an individual event.  The timestamp
// is an int64 representing seconds since the epoch (UTC).  The initial creator
// of the entry is responsible for converting to UTC.
//msgp:tuple Entry
type Entry struct {
	// Timestamp can contain the timestamp in either seconds or nanoseconds
	Timestamp int64
	// Record is the actual event record.
	Record Record
}

type MessageOptions struct {
	Size       int    `msg:"size"`
	Chunk      string `msg:"chunk,omitempty"`
	Compressed string `msg:"compressed,omitempty"`
}

// TODO: This is not working correctly yet
func NewCompressedPackedForwardMessage(
	tag string, entries []EntryExt, opts *MessageOptions,
) *PackedForwardMessage {
	// TODO: create buffer and writer pool
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	// TODO: capture and return error
	_, _ = zw.Write(eventStream(entries))
	zw.Close()

	opts.Size = len(entries)
	opts.Compressed = "gzip"

	// TODO:
	//   NewCompressedPackedForwardMessageFromBytes

	return NewPackedForwardMessageFromBytes(tag, buf.Bytes(), opts)
}
