package protocol_test

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.ibm.com/Observability/fluent-forward-go/fluent/protocol"
)

var _ = Describe("Transport", func() {
	Describe("EventTime", func() {
		var (
			ent Entry
		)

		BeforeEach(func() {
			ent = Entry{
				Timestamp: EventTime{
					Time: time.Unix(int64(1257894000), int64(12340000)),
				},
			}
		})

		// This covers both MarshalBinaryTo() and UnmarshalBinary()
		It("Marshals and unmarshals correctly", func() {
			b, err := ent.MarshalMsg(nil)

			Expect(err).NotTo(HaveOccurred())

			// This is the msgpack fixext8 encoding for the timestamp
			// per the fluent-forward spec:
			// D7 == fixext8
			// 00 == type 0
			// 4AF9F070 == 1257894000
			// 00BC4B20 == 12340000
			Expect(
				strings.Contains(fmt.Sprintf("%X", b), "D7004AF9F07000BC4B20"),
			).To(BeTrue())

			var unment Entry
			_, err = unment.UnmarshalMsg(b)
			Expect(err).NotTo(HaveOccurred())

			Expect(unment.Timestamp.Time.Equal(ent.Timestamp.Time)).To(BeTrue())
		})
	})

	Describe("NewCompressedPackedForwardMessage", func() {
		var (
			tag     string
			entries []Entry
			opts    MessageOptions
		)

		BeforeEach(func() {
			tag = "foo.bar"
			entries = []Entry{
				{
					Timestamp: EventTime{time.Now()},
					Record: map[string]string{
						"foo":    "bar",
						"george": "jungle",
					},
				},
				{
					Timestamp: EventTime{time.Now()},
					Record: map[string]string{
						"foo":    "kablooie",
						"george": "frank",
					},
				},
			}
			opts = MessageOptions{}
		})

		It("Returns a message with a gzip-compressed event stream", func() {
			msg := NewCompressedPackedForwardMessage(tag, entries, opts)
			Expect(msg).NotTo(BeNil())
		})
	})
})
