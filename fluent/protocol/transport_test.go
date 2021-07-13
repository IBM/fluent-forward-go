package protocol_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.ibm.com/Observability/fluent-forward-go/fluent/protocol"

	"github.com/vmihailenco/msgpack/v5"
)

var _ = Describe("Transport", func() {
	Describe("EntryTime", func() {
		Describe("MarshalMsgpack", func() {
			var (
				seconds, nanoseconds int64
				et                   *EntryTime
			)

			BeforeEach(func() {
				// If you change this, you *must* change the byte elements of the
				// returned slice from msgpack.Marshal() to match
				seconds = 1257894000
			})

			JustBeforeEach(func() {
				et = &EntryTime{
					Time: time.Unix(seconds, nanoseconds),
				}
			})

			It("Returns a properly marshaled byte sequence", func() {
				b, _ := msgpack.Marshal(et)
				Expect(b).To(Equal([]byte{0xD7, 0x00,
					0x4A, 0xF9, 0xF0, 0x70,
					0x00, 0x00, 0x00, 0x00}))
			})

			Context("When the timestamp includes nanoseconds", func() {
				BeforeEach(func() {
					nanoseconds = 500
				})

				It("Marshals the nanoseconds value correctly", func() {
					b, _ := msgpack.Marshal(et)
					Expect(b).To(Equal([]byte{0xD7, 0x00,
						0x4A, 0xF9, 0xF0, 0x70,
						0x00, 0x00, 0x01, 0xF4}))
				})
			})
		})

		Describe("UnmarshalMsgpack", func() {
			var (
				timeBytes []byte
			)

			BeforeEach(func() {
				timeBytes = []byte{0xD7, 0x00,
					0x00, 0x00, 0x01, 0xFF,
					0x00, 0x00, 0x10, 0xFF,
				}
			})

			It("Unmarshals to the correct timestamp", func() {
				dest := &EntryTime{}

				err := msgpack.Unmarshal(timeBytes, dest)
				Expect(err).NotTo(HaveOccurred())

				Expect(dest.Unix()).To(Equal(int64(511)))
				Expect(dest.Nanosecond()).To(Equal(4351))
			})

			Context("When the slice is the wrong length", func() {
				BeforeEach(func() {
					timeBytes = []byte{0xD7, 0x00}
				})

				It("Returns an error", func() {
					dest := &EntryTime{}

					err := msgpack.Unmarshal(timeBytes, dest)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
