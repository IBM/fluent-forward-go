package protocol_test

import (
	"bytes"
	"testing"

	"github.com/tinylib/msgp/msgp"

	. "github.com/IBM/fluent-forward-go/fluent/protocol"
)

func TestMarshalUnmarshalPackedForwardMessage(t *testing.T) {
	v := PackedForwardMessage{}
	bts, err := v.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	left, err := v.UnmarshalMsg(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after UnmarshalMsg(): %q", len(left), left)
	}

	left, err = msgp.Skip(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after Skip(): %q", len(left), left)
	}
}

func BenchmarkMarshalMsgPackedForwardMessage(b *testing.B) {
	v := PackedForwardMessage{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := v.MarshalMsg(nil)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkAppendMsgPackedForwardMessage(b *testing.B) {
	v := PackedForwardMessage{}
	bts := make([]byte, 0, v.Msgsize())
	bts, _ = v.MarshalMsg(bts[0:0])
	b.SetBytes(int64(len(bts)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bts, _ = v.MarshalMsg(bts[0:0])
	}
}

func BenchmarkUnmarshalPackedForwardMessage(b *testing.B) {
	v := PackedForwardMessage{}
	bts, _ := v.MarshalMsg(nil)
	b.ReportAllocs()
	b.SetBytes(int64(len(bts)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := v.UnmarshalMsg(bts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestEncodeDecodePackedForwardMessage(t *testing.T) {
	v := PackedForwardMessage{}
	var buf bytes.Buffer
	err := msgp.Encode(&buf, &v)
	if err != nil {
		t.Error(err)
	}

	m := v.Msgsize()
	if buf.Len() > m {
		t.Log("WARNING: TestEncodeDecodePackedForwardMessage Msgsize() is inaccurate")
	}

	vn := PackedForwardMessage{}
	err = msgp.Decode(&buf, &vn)
	if err != nil {
		t.Error(err)
	}

	buf.Reset()
	err = msgp.Encode(&buf, &v)
	if err != nil {
		t.Error(err)
	}
	err = msgp.NewReader(&buf).Skip()
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkEncodePackedForwardMessage(b *testing.B) {
	v := PackedForwardMessage{}
	var buf bytes.Buffer
	err := msgp.Encode(&buf, &v)
	if err != nil {
		b.Error(err)
	}
	b.SetBytes(int64(buf.Len()))
	en := msgp.NewWriter(msgp.Nowhere)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := v.EncodeMsg(en)
		if err != nil {
			b.Error(err)
		}
	}
	en.Flush()
}

func BenchmarkDecodePackedForwardMessage(b *testing.B) {
	v := PackedForwardMessage{}
	var buf bytes.Buffer
	err := msgp.Encode(&buf, &v)
	if err != nil {
		b.Error(err)
	}
	b.SetBytes(int64(buf.Len()))
	rd := msgp.NewEndlessReader(buf.Bytes(), b)
	dc := msgp.NewReader(rd)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := v.DecodeMsg(dc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestEncodeDecodeCompressedPackedForwardMessage(t *testing.T) {
	bits := make([]byte, 1028)
	v, err := NewCompressedPackedForwardMessageFromBytes("foo", bits)
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	err = msgp.Encode(&buf, v)
	if err != nil {
		t.Error(err)
	}

	m := v.Msgsize()
	if buf.Len() > m {
		t.Log("WARNING: TestEncodeDecodePackedForwardMessage Msgsize() is inaccurate")
	}

	vn := PackedForwardMessage{}
	err = msgp.Decode(&buf, &vn)
	if err != nil {
		t.Error(err)
	}

	buf.Reset()
	err = msgp.Encode(&buf, v)
	if err != nil {
		t.Error(err)
	}
	err = msgp.NewReader(&buf).Skip()
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkNCFMFB(b *testing.B) {
	bits := make([]byte, 1024)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewCompressedPackedForwardMessageFromBytes("foo", bits)
		if err != nil {
			b.Error(err)
		}
	}
}
