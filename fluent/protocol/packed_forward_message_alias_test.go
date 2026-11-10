package protocol_test

import (
	"testing"
	"time"
	"unsafe"

	. "github.com/IBM/fluent-forward-go/fluent/protocol"
)

func aliasEntries(v string) EntryList {
	return EntryList{{
		Timestamp: EventTime{Time: time.Unix(1_441_588_984, 0)},
		Record:    map[string]interface{}{"tenant": v},
	}}
}

func decodeTenant(t *testing.T, msg *PackedForwardMessage) string {
	t.Helper()

	var el EntryList
	if _, err := el.UnmarshalPacked(msg.EventStream); err != nil {
		t.Fatal(err)
	}

	if len(el) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(el))
	}

	return el[0].Record.(map[string]interface{})["tenant"].(string)
}

func TestMarshalPackedDoesNotAliasPooledBuffer(t *testing.T) {
	a, err := aliasEntries("AAAA").MarshalPacked()
	if err != nil {
		t.Fatal(err)
	}

	want := string(a)

	if _, err = aliasEntries("BBBB").MarshalPacked(); err != nil {
		t.Fatal(err)
	}

	if string(a) != want {
		t.Fatalf("first MarshalPacked result was overwritten by second call")
	}
}

func TestNewPackedForwardMessageDoesNotAlias(t *testing.T) {
	msgA, err := NewPackedForwardMessage("tenant.A", aliasEntries("AAAA"))
	if err != nil {
		t.Fatal(err)
	}

	msgB, err := NewPackedForwardMessage("tenant.B", aliasEntries("BBBB"))
	if err != nil {
		t.Fatal(err)
	}

	if unsafe.SliceData(msgA.EventStream) == unsafe.SliceData(msgB.EventStream) {
		t.Fatal("messages share a backing array")
	}

	if got := decodeTenant(t, msgA); got != "AAAA" {
		t.Fatalf("tenant.A message carries %q", got)
	}

	if got := decodeTenant(t, msgB); got != "BBBB" {
		t.Fatalf("tenant.B message carries %q", got)
	}
}

func TestNewCompressedPackedForwardMessageDoesNotAlias(t *testing.T) {
	msgA, err := NewCompressedPackedForwardMessage("tenant.A", aliasEntries("AAAA"))
	if err != nil {
		t.Fatal(err)
	}

	want := string(msgA.EventStream)

	msgB, err := NewCompressedPackedForwardMessage("tenant.B", aliasEntries("BBBB"))
	if err != nil {
		t.Fatal(err)
	}

	if unsafe.SliceData(msgA.EventStream) == unsafe.SliceData(msgB.EventStream) {
		t.Fatal("messages share a backing array")
	}

	if string(msgA.EventStream) != want {
		t.Fatal("compressed tenant.A stream was overwritten by tenant.B")
	}
}

// A failed MarshalPacked must not leave stale bytes in the pooled writer that
// would leak into a later, unrelated call.
func TestMarshalPackedErrorDoesNotPoisonPool(t *testing.T) {
	bad := EntryList{{
		Timestamp: EventTime{Time: time.Unix(1_441_588_984, 0)},
		Record:    map[string]interface{}{"k": make(chan int)},
	}}

	for i := 0; i < 10; i++ {
		if _, err := bad.MarshalPacked(); err == nil {
			t.Fatal("expected error encoding unsupported type")
		}

		got, err := aliasEntries("AAAA").MarshalPacked()
		if err != nil {
			t.Fatal(err)
		}

		var el EntryList
		if rest, err := el.UnmarshalPacked(got); err != nil || len(rest) != 0 || len(el) != 1 {
			t.Fatalf("corrupt output after failed call: err=%v rest=%d entries=%d", err, len(rest), len(el))
		}
	}
}
