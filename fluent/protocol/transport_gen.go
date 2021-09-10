package protocol

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Entry) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0001 uint32
	zb0001, err = dc.ReadArrayHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zb0001}
		return
	}
	z.Timestamp, err = dc.ReadInt64()
	if err != nil {
		err = msgp.WrapError(err, "Timestamp")
		return
	}
	var zb0002 uint32
	zb0002, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err, "Record")
		return
	}
	if z.Record == nil {
		z.Record = make(Record, zb0002)
	} else if len(z.Record) > 0 {
		for key := range z.Record {
			delete(z.Record, key)
		}
	}
	for zb0002 > 0 {
		zb0002--
		var za0001 string
		var za0002 interface{}
		za0001, err = dc.ReadString()
		if err != nil {
			err = msgp.WrapError(err, "Record")
			return
		}
		za0002, err = dc.ReadIntf()
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
		z.Record[za0001] = za0002
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Entry) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 2
	err = en.Append(0x92)
	if err != nil {
		return
	}
	err = en.WriteInt64(z.Timestamp)
	if err != nil {
		err = msgp.WrapError(err, "Timestamp")
		return
	}
	err = en.WriteMapHeader(uint32(len(z.Record)))
	if err != nil {
		err = msgp.WrapError(err, "Record")
		return
	}
	for za0001, za0002 := range z.Record {
		err = en.WriteString(za0001)
		if err != nil {
			err = msgp.WrapError(err, "Record")
			return
		}
		err = en.WriteIntf(za0002)
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Entry) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// array header, size 2
	o = append(o, 0x92)
	o = msgp.AppendInt64(o, z.Timestamp)
	o = msgp.AppendMapHeader(o, uint32(len(z.Record)))
	for za0001, za0002 := range z.Record {
		o = msgp.AppendString(o, za0001)
		o, err = msgp.AppendIntf(o, za0002)
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Entry) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zb0001}
		return
	}
	z.Timestamp, bts, err = msgp.ReadInt64Bytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Timestamp")
		return
	}
	var zb0002 uint32
	zb0002, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Record")
		return
	}
	if z.Record == nil {
		z.Record = make(Record, zb0002)
	} else if len(z.Record) > 0 {
		for key := range z.Record {
			delete(z.Record, key)
		}
	}
	for zb0002 > 0 {
		var za0001 string
		var za0002 interface{}
		zb0002--
		za0001, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, "Record")
			return
		}
		za0002, bts, err = msgp.ReadIntfBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
		z.Record[za0001] = za0002
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Entry) Msgsize() (s int) {
	s = 1 + msgp.Int64Size + msgp.MapHeaderSize
	if z.Record != nil {
		for za0001, za0002 := range z.Record {
			_ = za0002
			s += msgp.StringPrefixSize + len(za0001) + msgp.GuessSize(za0002)
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *EntryExt) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0001 uint32
	zb0001, err = dc.ReadArrayHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zb0001}
		return
	}
	err = dc.ReadExtension(&z.Timestamp)
	if err != nil {
		err = msgp.WrapError(err, "Timestamp")
		return
	}
	var zb0002 uint32
	zb0002, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err, "Record")
		return
	}
	if z.Record == nil {
		z.Record = make(Record, zb0002)
	} else if len(z.Record) > 0 {
		for key := range z.Record {
			delete(z.Record, key)
		}
	}
	for zb0002 > 0 {
		zb0002--
		var za0001 string
		var za0002 interface{}
		za0001, err = dc.ReadString()
		if err != nil {
			err = msgp.WrapError(err, "Record")
			return
		}
		za0002, err = dc.ReadIntf()
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
		z.Record[za0001] = za0002
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *EntryExt) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 2
	err = en.Append(0x92)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Timestamp)
	if err != nil {
		err = msgp.WrapError(err, "Timestamp")
		return
	}
	err = en.WriteMapHeader(uint32(len(z.Record)))
	if err != nil {
		err = msgp.WrapError(err, "Record")
		return
	}
	for za0001, za0002 := range z.Record {
		err = en.WriteString(za0001)
		if err != nil {
			err = msgp.WrapError(err, "Record")
			return
		}
		err = en.WriteIntf(za0002)
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *EntryExt) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// array header, size 2
	o = append(o, 0x92)
	o, err = msgp.AppendExtension(o, &z.Timestamp)
	if err != nil {
		err = msgp.WrapError(err, "Timestamp")
		return
	}
	o = msgp.AppendMapHeader(o, uint32(len(z.Record)))
	for za0001, za0002 := range z.Record {
		o = msgp.AppendString(o, za0001)
		o, err = msgp.AppendIntf(o, za0002)
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *EntryExt) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 2 {
		err = msgp.ArrayError{Wanted: 2, Got: zb0001}
		return
	}
	bts, err = msgp.ReadExtensionBytes(bts, &z.Timestamp)
	if err != nil {
		err = msgp.WrapError(err, "Timestamp")
		return
	}
	var zb0002 uint32
	zb0002, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Record")
		return
	}
	if z.Record == nil {
		z.Record = make(Record, zb0002)
	} else if len(z.Record) > 0 {
		for key := range z.Record {
			delete(z.Record, key)
		}
	}
	for zb0002 > 0 {
		var za0001 string
		var za0002 interface{}
		zb0002--
		za0001, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, "Record")
			return
		}
		za0002, bts, err = msgp.ReadIntfBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
		z.Record[za0001] = za0002
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *EntryExt) Msgsize() (s int) {
	s = 1 + msgp.ExtensionPrefixSize + z.Timestamp.Len() + msgp.MapHeaderSize
	if z.Record != nil {
		for za0001, za0002 := range z.Record {
			_ = za0002
			s += msgp.StringPrefixSize + len(za0001) + msgp.GuessSize(za0002)
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *EntryList) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0004 uint32
	zb0004, err = dc.ReadArrayHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if cap((*z)) >= int(zb0004) {
		(*z) = (*z)[:zb0004]
	} else {
		(*z) = make(EntryList, zb0004)
	}
	for zb0001 := range *z {
		var zb0005 uint32
		zb0005, err = dc.ReadArrayHeader()
		if err != nil {
			err = msgp.WrapError(err, zb0001)
			return
		}
		if zb0005 != 2 {
			err = msgp.ArrayError{Wanted: 2, Got: zb0005}
			return
		}
		err = dc.ReadExtension(&(*z)[zb0001].Timestamp)
		if err != nil {
			err = msgp.WrapError(err, zb0001, "Timestamp")
			return
		}
		var zb0006 uint32
		zb0006, err = dc.ReadMapHeader()
		if err != nil {
			err = msgp.WrapError(err, zb0001, "Record")
			return
		}
		if (*z)[zb0001].Record == nil {
			(*z)[zb0001].Record = make(Record, zb0006)
		} else if len((*z)[zb0001].Record) > 0 {
			for key := range (*z)[zb0001].Record {
				delete((*z)[zb0001].Record, key)
			}
		}
		for zb0006 > 0 {
			zb0006--
			var zb0002 string
			var zb0003 interface{}
			zb0002, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, zb0001, "Record")
				return
			}
			zb0003, err = dc.ReadIntf()
			if err != nil {
				err = msgp.WrapError(err, zb0001, "Record", zb0002)
				return
			}
			(*z)[zb0001].Record[zb0002] = zb0003
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z EntryList) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteArrayHeader(uint32(len(z)))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0007 := range z {
		// array header, size 2
		err = en.Append(0x92)
		if err != nil {
			return
		}
		err = en.WriteExtension(&z[zb0007].Timestamp)
		if err != nil {
			err = msgp.WrapError(err, zb0007, "Timestamp")
			return
		}
		err = en.WriteMapHeader(uint32(len(z[zb0007].Record)))
		if err != nil {
			err = msgp.WrapError(err, zb0007, "Record")
			return
		}
		for zb0008, zb0009 := range z[zb0007].Record {
			err = en.WriteString(zb0008)
			if err != nil {
				err = msgp.WrapError(err, zb0007, "Record")
				return
			}
			err = en.WriteIntf(zb0009)
			if err != nil {
				err = msgp.WrapError(err, zb0007, "Record", zb0008)
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z EntryList) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendArrayHeader(o, uint32(len(z)))
	for zb0007 := range z {
		// array header, size 2
		o = append(o, 0x92)
		o, err = msgp.AppendExtension(o, &z[zb0007].Timestamp)
		if err != nil {
			err = msgp.WrapError(err, zb0007, "Timestamp")
			return
		}
		o = msgp.AppendMapHeader(o, uint32(len(z[zb0007].Record)))
		for zb0008, zb0009 := range z[zb0007].Record {
			o = msgp.AppendString(o, zb0008)
			o, err = msgp.AppendIntf(o, zb0009)
			if err != nil {
				err = msgp.WrapError(err, zb0007, "Record", zb0008)
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *EntryList) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0004 uint32
	zb0004, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if cap((*z)) >= int(zb0004) {
		(*z) = (*z)[:zb0004]
	} else {
		(*z) = make(EntryList, zb0004)
	}
	for zb0001 := range *z {
		var zb0005 uint32
		zb0005, bts, err = msgp.ReadArrayHeaderBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, zb0001)
			return
		}
		if zb0005 != 2 {
			err = msgp.ArrayError{Wanted: 2, Got: zb0005}
			return
		}
		bts, err = msgp.ReadExtensionBytes(bts, &(*z)[zb0001].Timestamp)
		if err != nil {
			err = msgp.WrapError(err, zb0001, "Timestamp")
			return
		}
		var zb0006 uint32
		zb0006, bts, err = msgp.ReadMapHeaderBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, zb0001, "Record")
			return
		}
		if (*z)[zb0001].Record == nil {
			(*z)[zb0001].Record = make(Record, zb0006)
		} else if len((*z)[zb0001].Record) > 0 {
			for key := range (*z)[zb0001].Record {
				delete((*z)[zb0001].Record, key)
			}
		}
		for zb0006 > 0 {
			var zb0002 string
			var zb0003 interface{}
			zb0006--
			zb0002, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, zb0001, "Record")
				return
			}
			zb0003, bts, err = msgp.ReadIntfBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, zb0001, "Record", zb0002)
				return
			}
			(*z)[zb0001].Record[zb0002] = zb0003
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z EntryList) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize
	for zb0007 := range z {
		s += 1 + msgp.ExtensionPrefixSize + z[zb0007].Timestamp.Len() + msgp.MapHeaderSize
		if z[zb0007].Record != nil {
			for zb0008, zb0009 := range z[zb0007].Record {
				_ = zb0009
				s += msgp.StringPrefixSize + len(zb0008) + msgp.GuessSize(zb0009)
			}
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *EventTime) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "Time":
			z.Time, err = dc.ReadTime()
			if err != nil {
				err = msgp.WrapError(err, "Time")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z EventTime) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "Time"
	err = en.Append(0x81, 0xa4, 0x54, 0x69, 0x6d, 0x65)
	if err != nil {
		return
	}
	err = en.WriteTime(z.Time)
	if err != nil {
		err = msgp.WrapError(err, "Time")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z EventTime) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "Time"
	o = append(o, 0x81, 0xa4, 0x54, 0x69, 0x6d, 0x65)
	o = msgp.AppendTime(o, z.Time)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *EventTime) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "Time":
			z.Time, bts, err = msgp.ReadTimeBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Time")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z EventTime) Msgsize() (s int) {
	s = 1 + 5 + msgp.TimeSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MessageOptions) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "size":
			z.Size, err = dc.ReadInt()
			if err != nil {
				err = msgp.WrapError(err, "Size")
				return
			}
		case "chunk":
			z.Chunk, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Chunk")
				return
			}
		case "compressed":
			z.Compressed, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Compressed")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z MessageOptions) EncodeMsg(en *msgp.Writer) (err error) {
	// omitempty: check for empty values
	zb0001Len := uint32(3)
	var zb0001Mask uint8 /* 3 bits */
	if z.Size == 0 {
		zb0001Len--
		zb0001Mask |= 0x1
	}
	if z.Chunk == "" {
		zb0001Len--
		zb0001Mask |= 0x2
	}
	if z.Compressed == "" {
		zb0001Len--
		zb0001Mask |= 0x4
	}
	// variable map header, size zb0001Len
	err = en.Append(0x80 | uint8(zb0001Len))
	if err != nil {
		return
	}
<<<<<<< HEAD
	if z.Record == nil {
		z.Record = make(Record, zb0002)
	} else if len(z.Record) > 0 {
		for key := range z.Record {
			delete(z.Record, key)
		}
=======
	if zb0001Len == 0 {
		return
>>>>>>> add custom marshalling for Message and PackedMessage
	}
	if (zb0001Mask & 0x1) == 0 { // if not empty
		// write "size"
		err = en.Append(0xa4, 0x73, 0x69, 0x7a, 0x65)
		if err != nil {
			return
		}
		err = en.WriteInt(z.Size)
		if err != nil {
			err = msgp.WrapError(err, "Size")
			return
		}
	}
	if (zb0001Mask & 0x2) == 0 { // if not empty
		// write "chunk"
		err = en.Append(0xa5, 0x63, 0x68, 0x75, 0x6e, 0x6b)
		if err != nil {
			return
		}
		err = en.WriteString(z.Chunk)
		if err != nil {
<<<<<<< HEAD
			err = msgp.WrapError(err, "Options")
			return
		}
		for zb0003 > 0 {
			zb0003--
			field, err = dc.ReadMapKeyPtr()
			if err != nil {
				err = msgp.WrapError(err, "Options")
				return
			}
			switch msgp.UnsafeString(field) {
			case "size":
				z.Options.Size, err = dc.ReadInt()
				if err != nil {
					err = msgp.WrapError(err, "Options", "Size")
					return
				}
			case "chunk":
				z.Options.Chunk, err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "Options", "Chunk")
					return
				}
			case "compressed":
				z.Options.Compressed, err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "Options", "Compressed")
					return
				}
			default:
				err = dc.Skip()
				if err != nil {
					err = msgp.WrapError(err, "Options")
					return
				}
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Message) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 4
	err = en.Append(0x94)
	if err != nil {
		return
	}
	err = en.WriteString(z.Tag)
	if err != nil {
		err = msgp.WrapError(err, "Tag")
		return
	}
	err = en.WriteInt64(z.Timestamp)
	if err != nil {
		err = msgp.WrapError(err, "Timestamp")
		return
	}
	err = en.WriteMapHeader(uint32(len(z.Record)))
	if err != nil {
		err = msgp.WrapError(err, "Record")
		return
	}
	for za0001, za0002 := range z.Record {
		err = en.WriteString(za0001)
		if err != nil {
			err = msgp.WrapError(err, "Record")
			return
		}
		err = en.WriteIntf(za0002)
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
	}
	if z.Options == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		// map header, size 3
		// write "size"
		err = en.Append(0x83, 0xa4, 0x73, 0x69, 0x7a, 0x65)
		if err != nil {
			return
		}
		err = en.WriteInt(z.Options.Size)
		if err != nil {
			err = msgp.WrapError(err, "Options", "Size")
			return
		}
		// write "chunk"
		err = en.Append(0xa5, 0x63, 0x68, 0x75, 0x6e, 0x6b)
		if err != nil {
			return
		}
		err = en.WriteString(z.Options.Chunk)
		if err != nil {
			err = msgp.WrapError(err, "Options", "Chunk")
			return
		}
		// write "compressed"
		err = en.Append(0xaa, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x65, 0x64)
		if err != nil {
			return
		}
		err = en.WriteString(z.Options.Compressed)
		if err != nil {
			err = msgp.WrapError(err, "Options", "Compressed")
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Message) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// array header, size 4
	o = append(o, 0x94)
	o = msgp.AppendString(o, z.Tag)
	o = msgp.AppendInt64(o, z.Timestamp)
	o = msgp.AppendMapHeader(o, uint32(len(z.Record)))
	for za0001, za0002 := range z.Record {
		o = msgp.AppendString(o, za0001)
		o, err = msgp.AppendIntf(o, za0002)
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
	}
	if z.Options == nil {
		o = msgp.AppendNil(o)
	} else {
		// map header, size 3
		// string "size"
		o = append(o, 0x83, 0xa4, 0x73, 0x69, 0x7a, 0x65)
		o = msgp.AppendInt(o, z.Options.Size)
		// string "chunk"
		o = append(o, 0xa5, 0x63, 0x68, 0x75, 0x6e, 0x6b)
		o = msgp.AppendString(o, z.Options.Chunk)
		// string "compressed"
		o = append(o, 0xaa, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x65, 0x64)
		o = msgp.AppendString(o, z.Options.Compressed)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Message) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 4 {
		err = msgp.ArrayError{Wanted: 4, Got: zb0001}
		return
	}
	z.Tag, bts, err = msgp.ReadStringBytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Tag")
		return
	}
	z.Timestamp, bts, err = msgp.ReadInt64Bytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Timestamp")
		return
	}
	var zb0002 uint32
	zb0002, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Record")
		return
	}
	if z.Record == nil {
		z.Record = make(Record, zb0002)
	} else if len(z.Record) > 0 {
		for key := range z.Record {
			delete(z.Record, key)
		}
	}
	for zb0002 > 0 {
		var za0001 string
		var za0002 interface{}
		zb0002--
		za0001, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, "Record")
			return
		}
		za0002, bts, err = msgp.ReadIntfBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
		z.Record[za0001] = za0002
	}
	if msgp.IsNil(bts) {
		bts, err = msgp.ReadNilBytes(bts)
		if err != nil {
			return
		}
		z.Options = nil
	} else {
		if z.Options == nil {
			z.Options = new(MessageOptions)
		}
		var field []byte
		_ = field
		var zb0003 uint32
		zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, "Options")
			return
		}
		for zb0003 > 0 {
			zb0003--
			field, bts, err = msgp.ReadMapKeyZC(bts)
			if err != nil {
				err = msgp.WrapError(err, "Options")
				return
			}
			switch msgp.UnsafeString(field) {
			case "size":
				z.Options.Size, bts, err = msgp.ReadIntBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Options", "Size")
					return
				}
			case "chunk":
				z.Options.Chunk, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Options", "Chunk")
					return
				}
			case "compressed":
				z.Options.Compressed, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Options", "Compressed")
					return
				}
			default:
				bts, err = msgp.Skip(bts)
				if err != nil {
					err = msgp.WrapError(err, "Options")
					return
				}
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Message) Msgsize() (s int) {
	s = 1 + msgp.StringPrefixSize + len(z.Tag) + msgp.Int64Size + msgp.MapHeaderSize
	if z.Record != nil {
		for za0001, za0002 := range z.Record {
			_ = za0002
			s += msgp.StringPrefixSize + len(za0001) + msgp.GuessSize(za0002)
		}
	}
	if z.Options == nil {
		s += msgp.NilSize
	} else {
		s += 1 + 5 + msgp.IntSize + 6 + msgp.StringPrefixSize + len(z.Options.Chunk) + 11 + msgp.StringPrefixSize + len(z.Options.Compressed)
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MessageExt) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0001 uint32
	zb0001, err = dc.ReadArrayHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 4 {
		err = msgp.ArrayError{Wanted: 4, Got: zb0001}
		return
	}
	z.Tag, err = dc.ReadString()
	if err != nil {
		err = msgp.WrapError(err, "Tag")
		return
	}
	err = dc.ReadExtension(&z.Timestamp)
	if err != nil {
		err = msgp.WrapError(err, "Timestamp")
		return
	}
	var zb0002 uint32
	zb0002, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err, "Record")
		return
	}
	if z.Record == nil {
		z.Record = make(Record, zb0002)
	} else if len(z.Record) > 0 {
		for key := range z.Record {
			delete(z.Record, key)
		}
	}
	for zb0002 > 0 {
		zb0002--
		var za0001 string
		var za0002 interface{}
		za0001, err = dc.ReadString()
		if err != nil {
			err = msgp.WrapError(err, "Record")
			return
		}
		za0002, err = dc.ReadIntf()
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
		z.Record[za0001] = za0002
	}
	if dc.IsNil() {
		err = dc.ReadNil()
		if err != nil {
			err = msgp.WrapError(err, "Options")
			return
		}
		z.Options = nil
	} else {
		if z.Options == nil {
			z.Options = new(MessageOptions)
		}
		var field []byte
		_ = field
		var zb0003 uint32
		zb0003, err = dc.ReadMapHeader()
		if err != nil {
			err = msgp.WrapError(err, "Options")
			return
		}
		for zb0003 > 0 {
			zb0003--
			field, err = dc.ReadMapKeyPtr()
			if err != nil {
				err = msgp.WrapError(err, "Options")
				return
			}
			switch msgp.UnsafeString(field) {
			case "size":
				z.Options.Size, err = dc.ReadInt()
				if err != nil {
					err = msgp.WrapError(err, "Options", "Size")
					return
				}
			case "chunk":
				z.Options.Chunk, err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "Options", "Chunk")
					return
				}
			case "compressed":
				z.Options.Compressed, err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "Options", "Compressed")
					return
				}
			default:
				err = dc.Skip()
				if err != nil {
					err = msgp.WrapError(err, "Options")
					return
				}
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *MessageExt) EncodeMsg(en *msgp.Writer) (err error) {
	// array header, size 4
	err = en.Append(0x94)
	if err != nil {
		return
	}
	err = en.WriteString(z.Tag)
	if err != nil {
		err = msgp.WrapError(err, "Tag")
		return
	}
	err = en.WriteExtension(&z.Timestamp)
	if err != nil {
		err = msgp.WrapError(err, "Timestamp")
		return
	}
	err = en.WriteMapHeader(uint32(len(z.Record)))
	if err != nil {
		err = msgp.WrapError(err, "Record")
		return
	}
	for za0001, za0002 := range z.Record {
		err = en.WriteString(za0001)
		if err != nil {
			err = msgp.WrapError(err, "Record")
			return
		}
		err = en.WriteIntf(za0002)
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
	}
	if z.Options == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		// map header, size 3
		// write "size"
		err = en.Append(0x83, 0xa4, 0x73, 0x69, 0x7a, 0x65)
		if err != nil {
			return
		}
		err = en.WriteInt(z.Options.Size)
		if err != nil {
			err = msgp.WrapError(err, "Options", "Size")
			return
		}
		// write "chunk"
		err = en.Append(0xa5, 0x63, 0x68, 0x75, 0x6e, 0x6b)
		if err != nil {
			return
		}
		err = en.WriteString(z.Options.Chunk)
		if err != nil {
			err = msgp.WrapError(err, "Options", "Chunk")
=======
			err = msgp.WrapError(err, "Chunk")
>>>>>>> add custom marshalling for Message and PackedMessage
			return
		}
	}
	if (zb0001Mask & 0x4) == 0 { // if not empty
		// write "compressed"
		err = en.Append(0xaa, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x65, 0x64)
		if err != nil {
			return
		}
		err = en.WriteString(z.Compressed)
		if err != nil {
			err = msgp.WrapError(err, "Compressed")
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z MessageOptions) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
<<<<<<< HEAD
	// array header, size 4
	o = append(o, 0x94)
	o = msgp.AppendString(o, z.Tag)
	o, err = msgp.AppendExtension(o, &z.Timestamp)
	if err != nil {
		err = msgp.WrapError(err, "Timestamp")
		return
	}
	o = msgp.AppendMapHeader(o, uint32(len(z.Record)))
	for za0001, za0002 := range z.Record {
		o = msgp.AppendString(o, za0001)
		o, err = msgp.AppendIntf(o, za0002)
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
	}
	if z.Options == nil {
		o = msgp.AppendNil(o)
	} else {
		// map header, size 3
		// string "size"
		o = append(o, 0x83, 0xa4, 0x73, 0x69, 0x7a, 0x65)
		o = msgp.AppendInt(o, z.Options.Size)
		// string "chunk"
		o = append(o, 0xa5, 0x63, 0x68, 0x75, 0x6e, 0x6b)
		o = msgp.AppendString(o, z.Options.Chunk)
		// string "compressed"
		o = append(o, 0xaa, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x65, 0x64)
		o = msgp.AppendString(o, z.Options.Compressed)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MessageExt) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 4 {
		err = msgp.ArrayError{Wanted: 4, Got: zb0001}
		return
	}
	z.Tag, bts, err = msgp.ReadStringBytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Tag")
		return
	}
	bts, err = msgp.ReadExtensionBytes(bts, &z.Timestamp)
	if err != nil {
		err = msgp.WrapError(err, "Timestamp")
		return
	}
	var zb0002 uint32
	zb0002, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "Record")
		return
	}
	if z.Record == nil {
		z.Record = make(Record, zb0002)
	} else if len(z.Record) > 0 {
		for key := range z.Record {
			delete(z.Record, key)
		}
	}
	for zb0002 > 0 {
		var za0001 string
		var za0002 interface{}
		zb0002--
		za0001, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, "Record")
			return
		}
		za0002, bts, err = msgp.ReadIntfBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, "Record", za0001)
			return
		}
		z.Record[za0001] = za0002
	}
	if msgp.IsNil(bts) {
		bts, err = msgp.ReadNilBytes(bts)
		if err != nil {
			return
		}
		z.Options = nil
	} else {
		if z.Options == nil {
			z.Options = new(MessageOptions)
		}
		var field []byte
		_ = field
		var zb0003 uint32
		zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, "Options")
			return
		}
		for zb0003 > 0 {
			zb0003--
			field, bts, err = msgp.ReadMapKeyZC(bts)
			if err != nil {
				err = msgp.WrapError(err, "Options")
				return
			}
			switch msgp.UnsafeString(field) {
			case "size":
				z.Options.Size, bts, err = msgp.ReadIntBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Options", "Size")
					return
				}
			case "chunk":
				z.Options.Chunk, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Options", "Chunk")
					return
				}
			case "compressed":
				z.Options.Compressed, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Options", "Compressed")
					return
				}
			default:
				bts, err = msgp.Skip(bts)
				if err != nil {
					err = msgp.WrapError(err, "Options")
					return
				}
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *MessageExt) Msgsize() (s int) {
	s = 1 + msgp.StringPrefixSize + len(z.Tag) + msgp.ExtensionPrefixSize + z.Timestamp.Len() + msgp.MapHeaderSize
	if z.Record != nil {
		for za0001, za0002 := range z.Record {
			_ = za0002
			s += msgp.StringPrefixSize + len(za0001) + msgp.GuessSize(za0002)
		}
	}
	if z.Options == nil {
		s += msgp.NilSize
	} else {
		s += 1 + 5 + msgp.IntSize + 6 + msgp.StringPrefixSize + len(z.Options.Chunk) + 11 + msgp.StringPrefixSize + len(z.Options.Compressed)
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MessageOptions) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "size":
			z.Size, err = dc.ReadInt()
			if err != nil {
				err = msgp.WrapError(err, "Size")
				return
			}
		case "chunk":
			z.Chunk, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Chunk")
				return
			}
		case "compressed":
			z.Compressed, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Compressed")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
=======
	// omitempty: check for empty values
	zb0001Len := uint32(3)
	var zb0001Mask uint8 /* 3 bits */
	if z.Size == 0 {
		zb0001Len--
		zb0001Mask |= 0x1
>>>>>>> add custom marshalling for Message and PackedMessage
	}
	if z.Chunk == "" {
		zb0001Len--
		zb0001Mask |= 0x2
	}
	if z.Compressed == "" {
		zb0001Len--
		zb0001Mask |= 0x4
	}
	// variable map header, size zb0001Len
	o = append(o, 0x80|uint8(zb0001Len))
	if zb0001Len == 0 {
		return
	}
	if (zb0001Mask & 0x1) == 0 { // if not empty
		// string "size"
		o = append(o, 0xa4, 0x73, 0x69, 0x7a, 0x65)
		o = msgp.AppendInt(o, z.Size)
	}
	if (zb0001Mask & 0x2) == 0 { // if not empty
		// string "chunk"
		o = append(o, 0xa5, 0x63, 0x68, 0x75, 0x6e, 0x6b)
		o = msgp.AppendString(o, z.Chunk)
	}
	if (zb0001Mask & 0x4) == 0 { // if not empty
		// string "compressed"
		o = append(o, 0xaa, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x65, 0x64)
		o = msgp.AppendString(o, z.Compressed)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MessageOptions) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "size":
			z.Size, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Size")
				return
			}
		case "chunk":
			z.Chunk, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Chunk")
				return
			}
		case "compressed":
			z.Compressed, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Compressed")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z MessageOptions) Msgsize() (s int) {
	s = 1 + 5 + msgp.IntSize + 6 + msgp.StringPrefixSize + len(z.Chunk) + 11 + msgp.StringPrefixSize + len(z.Compressed)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Record) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0003 uint32
	zb0003, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if (*z) == nil {
		(*z) = make(Record, zb0003)
	} else if len((*z)) > 0 {
		for key := range *z {
			delete((*z), key)
		}
	}
	for zb0003 > 0 {
		zb0003--
		var zb0001 string
		var zb0002 interface{}
		zb0001, err = dc.ReadString()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		zb0002, err = dc.ReadIntf()
		if err != nil {
			err = msgp.WrapError(err, zb0001)
			return
		}
		(*z)[zb0001] = zb0002
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Record) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0004, zb0005 := range z {
		err = en.WriteString(zb0004)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		err = en.WriteIntf(zb0005)
		if err != nil {
			err = msgp.WrapError(err, zb0004)
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Record) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendMapHeader(o, uint32(len(z)))
	for zb0004, zb0005 := range z {
		o = msgp.AppendString(o, zb0004)
		o, err = msgp.AppendIntf(o, zb0005)
		if err != nil {
			err = msgp.WrapError(err, zb0004)
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Record) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0003 uint32
	zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if (*z) == nil {
		(*z) = make(Record, zb0003)
	} else if len((*z)) > 0 {
		for key := range *z {
			delete((*z), key)
		}
	}
	for zb0003 > 0 {
		var zb0001 string
		var zb0002 interface{}
		zb0003--
		zb0001, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		zb0002, bts, err = msgp.ReadIntfBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, zb0001)
			return
		}
		(*z)[zb0001] = zb0002
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Record) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for zb0004, zb0005 := range z {
			_ = zb0005
			s += msgp.StringPrefixSize + len(zb0004) + msgp.GuessSize(zb0005)
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Record) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0003 uint32
	zb0003, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if (*z) == nil {
		(*z) = make(Record, zb0003)
	} else if len((*z)) > 0 {
		for key := range *z {
			delete((*z), key)
		}
	}
	for zb0003 > 0 {
		zb0003--
		var zb0001 string
		var zb0002 interface{}
		zb0001, err = dc.ReadString()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		zb0002, err = dc.ReadIntf()
		if err != nil {
			err = msgp.WrapError(err, zb0001)
			return
		}
		(*z)[zb0001] = zb0002
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Record) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteMapHeader(uint32(len(z)))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0004, zb0005 := range z {
		err = en.WriteString(zb0004)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		err = en.WriteIntf(zb0005)
		if err != nil {
			err = msgp.WrapError(err, zb0004)
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Record) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendMapHeader(o, uint32(len(z)))
	for zb0004, zb0005 := range z {
		o = msgp.AppendString(o, zb0004)
		o, err = msgp.AppendIntf(o, zb0005)
		if err != nil {
			err = msgp.WrapError(err, zb0004)
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Record) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0003 uint32
	zb0003, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if (*z) == nil {
		(*z) = make(Record, zb0003)
	} else if len((*z)) > 0 {
		for key := range *z {
			delete((*z), key)
		}
	}
	for zb0003 > 0 {
		var zb0001 string
		var zb0002 interface{}
		zb0003--
		zb0001, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		zb0002, bts, err = msgp.ReadIntfBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, zb0001)
			return
		}
		(*z)[zb0001] = zb0002
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Record) Msgsize() (s int) {
	s = msgp.MapHeaderSize
	if z != nil {
		for zb0004, zb0005 := range z {
			_ = zb0005
			s += msgp.StringPrefixSize + len(zb0004) + msgp.GuessSize(zb0005)
		}
	}
	return
}
