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
	"encoding/binary"
	"errors"
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp

// =========
// TRANSPORT
// =========

const (
	OptSize       string = "size"
	OptChunk      string = "chunk"
	OptCompressed string = "compressed"
	OptValGZIP    string = "gzip"

	extensionType int8 = 0
	eventTimeLen  int  = 8
)

var (
	compressorPool  sync.Pool
	chunkReaderPool sync.Pool
	bufferPool      sync.Pool
)

func init() {
	uuid.EnableRandPool()

	msgp.RegisterExtension(extensionType, func() msgp.Extension {
		return new(EventTime)
	})

	compressorPool.New = func() interface{} {
		return new(GzipCompressor)
	}

	chunkReaderPool.New = func() interface{} {
		return new(ChunkReader)
	}

	bufferPool.New = func() interface{} {
		return new(bytes.Buffer)
	}
}

// EventTime is the fluent-forward representation of a timestamp
type EventTime struct {
	time.Time
}

// EventTimeNow returns an EventTime set to time.Now().UTC().
func EventTimeNow() EventTime {
	return EventTime{
		Time: time.Now().UTC(),
	}
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
	binary.BigEndian.PutUint32(b, uint32(utc.Unix()))           /* #nosec G115 */
	binary.BigEndian.PutUint32(b[4:], uint32(utc.Nanosecond())) /* #nosec G115 */

	return nil
}

// UnmarshalBinary implements the Extension interface for unmarshaling
// into an EventTime object.
func (et *EventTime) UnmarshalBinary(timeBytes []byte) error {
	if len(timeBytes) != eventTimeLen {
		return errors.New("invalid length")
	}

	seconds := binary.BigEndian.Uint32(timeBytes)
	nanoseconds := binary.BigEndian.Uint32(timeBytes[4:])

	et.Time = time.Unix(int64(seconds), int64(nanoseconds))

	return nil
}

// EntryExt is the basic representation of an individual event, but using the
// msgpack extension format for the timestamp.
//
//msgp:tuple EntryExt
//msgp:decode ignore EntryExt
//msgp:unmarshal ignore EntryExt
type EntryExt struct {
	// Timestamp can contain the timestamp in either seconds or nanoseconds
	Timestamp EventTime `msg:"eventTime,extension"`
	// Record is the actual event record. The object must be a map or
	// struct. Objects that implement the msgp.Encodable interface will
	// be the most performant.
	Record interface{}
}

// This decodes msgpack data from a a msgp.Reader into EntryExt
// - [timestamp, record]
// - where timestamp can be an integer or [integer, {metadata}] (we ignore metadata)
func (z *EntryExt) DecodeMsg(r *msgp.Reader) error {
	// read the array header to get the number of elements
	// expected 2 elements as per Forward Specification
	arrayLen, err := r.ReadArrayHeader()
	if err != nil {
		return msgp.WrapError(err, "Array Header")
	}

	// validate Forward Protocol format: must have 2 elements
	if arrayLen != 2 {
		return msgp.ArrayError{Wanted: 2, Got: arrayLen}
	}

	// peek at the next msgpack type to determine timestamp format
	nextType, err := r.NextType()
	if err != nil {
		return msgp.WrapError(err, "Timestamp peek")
	}

	switch nextType {
	case msgp.ExtensionType:
		// extension format
		err = r.ReadExtension(&z.Timestamp)
		if err != nil {
			return msgp.WrapError(err, "Timestamp extension")
		}
	case msgp.ArrayType:
		// timestamp with metadata format: [timestamp, metadata]
		var arrSize uint32
		arrSize, err = r.ReadArrayHeader()
		if err != nil {
			return msgp.WrapError(err, "Timestamp array header")
		}

		if arrSize != 2 {
			return msgp.WrapError(err, "Invalid timestamp array size")
		}

		nt, err := r.NextType()
		if err != nil {
			return msgp.WrapError(err, "Timestamp peek")
		}

		switch nt {
		case msgp.ExtensionType:
			err = r.ReadExtension(&z.Timestamp)
			if err != nil {
				return msgp.WrapError(err, "Timestamp extension from array")
			}
		case msgp.IntType:
			// read the timestamp value
			var seconds int64
			seconds, err = r.ReadInt64()
			if err != nil {
				return msgp.WrapError(err, "Timestamp from array")
			}

			z.Timestamp = EventTime{time.Unix(seconds, 0).UTC()}
		}

		// Skip metadata elements (all elements after the first)
		// We iterate starting from 1 because element 0 (timestamp) is already read
		for i := uint32(1); i < arrSize; i++ {
			// Skip() advances past one complete msgpack object in the byte slice
			err = r.Skip()
			if err != nil {
				return msgp.WrapError(err, "Skip metadata element")
			}
		}

	case msgp.IntType:
		// this is simple timestamp format: unix seconds as integer
		var sec int64
		sec, err = r.ReadInt64()
		if err != nil {
			return msgp.WrapError(err, "Timestamp int")
		}
		z.Timestamp = EventTime{time.Unix(sec, 0).UTC()}

	case msgp.UintType:
		// unsigned integer timestamp
		var usec uint64
		usec, err = r.ReadUint64()
		if err != nil {
			return msgp.WrapError(err, "Timestamp uint")
		}
		if usec > math.MaxInt64 {
			return msgp.WrapError(err, "timestamp overflow exceeds maximum int64 value")
		}
		z.Timestamp = EventTime{time.Unix(int64(usec), 0).UTC()} /* #nosec G115 */

	default:
		return msgp.WrapError(err, "unsupported timestamp type: expected int or array")
	}

	// rad the record (second element)
	z.Record, err = r.ReadIntf()
	if err != nil {
		return msgp.WrapError(err, "Record")
	}

	return nil
}

// This decodes msgpack data from a byte slice into EntryExt
// - [timestamp, record]
// - where timestamp can be an integer or [integer, {metadata}] (we ignore metadata)
func (z *EntryExt) UnmarshalMsg(bts []byte) ([]byte, error) {
	// read array header to get number of elements
	arrayLen, bts, err := msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Array Header")
		return bts, err
	}

	// validate Forward Protocol format: must have 2 elements
	if arrayLen != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: arrayLen}
		return bts, err
	}

	// peek at the next msgpack type to determine timestamp format
	nextType := msgp.NextType(bts)

	switch nextType {
	case msgp.ExtensionType:
		// extension format
		bts, err = msgp.ReadExtensionBytes(bts, &z.Timestamp)
		if err != nil {
			return bts, msgp.WrapError(err, "Timestamp extension")
		}

	case msgp.ArrayType:
		// timestamp with metadata format: [timestamp, metadata]
		var arrSize uint32
		arrSize, bts, err = msgp.ReadArrayHeaderBytes(bts)
		if err != nil {
			return bts, msgp.WrapError(err, "Timestamp array header")
		}

		if arrSize != 2 {
			return bts, msgp.WrapError(err, "Timestamp array size", arrSize)
		}

		nt := msgp.NextType(bts)
		switch nt {
		case msgp.ExtensionType:
			bts, err = msgp.ReadExtensionBytes(bts, &z.Timestamp)
			if err != nil {
				return bts, msgp.WrapError(err, "Timestamp extension from array")
			}
		case msgp.IntType:
			// read the timestamp value
			var seconds int64
			seconds, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return bts, msgp.WrapError(err, "Timstamp from array")
			}

			z.Timestamp = EventTime{time.Unix(seconds, 0).UTC()}
		}

		// Skip metadata elements (all elements after the first)
		// We iterate starting from 1 because element 0 (timestamp) is already read
		for i := uint32(1); i < arrSize; i++ {
			// Skip() advances past one complete msgpack object in the byte slice
			bts, err = msgp.Skip(bts)
			if err != nil {
				return bts, msgp.WrapError(err, "Skip metadata element")
			}
		}

	case msgp.IntType:
		// this is simple timestamp format: unix seconds as integer
		var sec int64
		sec, bts, err = msgp.ReadInt64Bytes(bts)
		if err != nil {
			return bts, msgp.WrapError(err, "Timestamp int")
		}
		z.Timestamp = EventTime{time.Unix(sec, 0).UTC()}

	case msgp.UintType:
		// unsigned integer timestamp
		var usec uint64
		usec, bts, err = msgp.ReadUint64Bytes(bts)
		if err != nil {
			return bts, msgp.WrapError(err, "Timestamp uint")
		}
		if usec > math.MaxInt64 {
			return bts, msgp.WrapError(err, "timestamp overflow, exceeds maximum int64 value")
		}
		z.Timestamp = EventTime{time.Unix(int64(usec), 0).UTC()}

	default:
		return bts, msgp.WrapError(err, "unsupported timestamp type: expected int or array")
	}

	// rad the record (second element)
	z.Record, bts, err = msgp.ReadIntfBytes(bts)
	if err != nil {
		return bts, msgp.WrapError(err, "Record")
	}

	return bts, nil
}

type EntryList []EntryExt

func (el *EntryList) UnmarshalPacked(bits []byte) ([]byte, error) {
	var (
		entry EntryExt
		err   error
	)

	*el = (*el)[:0]

	for len(bits) > 0 {
		if bits, err = entry.UnmarshalMsg(bits); err != nil {
			break
		}

		*el = append(*el, entry)
	}

	return bits, err
}

func (el EntryList) MarshalPacked() ([]byte, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	defer func() {
		bufferPool.Put(buf)
	}()

	for _, e := range el {
		if err := msgp.Encode(buf, e); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// Equal compares two EntryList objects and returns true if they have
// exactly the same elements, false otherwise.
func (el EntryList) Equal(e2 EntryList) bool {
	if len(el) != len(e2) {
		return false
	}

	first := make(EntryList, len(el))

	copy(first, el)

	second := make(EntryList, len(e2))

	copy(second, e2)

	matches := 0

	for _, ea := range first {
		for _, eb := range second {
			if ea.Timestamp.Equal(eb.Timestamp.Time) {
				// Timestamps equal, check the record
				if reflect.DeepEqual(ea.Record, eb.Record) {
					matches++
				}
			}
		}
	}

	return matches == len(el)
}

// EntryExt is the basic representation of an individual event.  The timestamp
// is an int64 representing seconds since the epoch (UTC).  The initial creator
// of the entry is responsible for converting to UTC.
//
//msgp:tuple Entry
type Entry struct {
	// Timestamp can contain the timestamp in either seconds or nanoseconds
	Timestamp int64
	// Record is the actual event record.
	Record interface{}
}

type MessageOptions struct {
	Size       *int   `msg:"size,omitempty"`
	Chunk      string `msg:"chunk,omitempty"`
	Compressed string `msg:"compressed,omitempty"`
}

type AckMessage struct {
	Ack string `msg:"ack"`
}

// RawMessage is a ChunkEncoder wrapper for []byte.
//
//msgp:encode ignore RawMessage
type RawMessage []byte

func (rm RawMessage) EncodeMsg(w *msgp.Writer) error {
	if len(rm) == 0 {
		return w.WriteNil()
	}

	_, err := w.Write([]byte(rm))

	return err
}

// Chunk searches the message for the chunk ID. In the case of RawMessage,
// Chunk is read-only. It returns an error if the chunk is not found.
func (rm RawMessage) Chunk() (string, error) {
	return GetChunk([]byte(rm))
}
