package protocol_test

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/IBM/fluent-forward-go/fluent/protocol"
)

var _ = Describe("Transport", func() {
	Describe("EventTime", func() {
		var (
			ent EventExt
		)

		BeforeEach(func() {
			ent = EventExt{
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

			var unment EventExt
			_, err = unment.UnmarshalMsg(b)
			Expect(err).NotTo(HaveOccurred())

			Expect(unment.Timestamp.Time.Equal(ent.Timestamp.Time)).To(BeTrue())
		})
	})

	Describe("EventStream", func() {
		var (
			e1 EventStream
			et time.Time
		)

		BeforeEach(func() {
			et = time.Now()
			e1 = EventStream{
				{
					Timestamp: EventTime{et},
					Record: map[string]interface{}{
						"foo":    "bar",
						"george": "jungle",
					},
				},
				{
					Timestamp: EventTime{et},
					Record: map[string]interface{}{
						"foo":    "kablooie",
						"george": "frank",
					},
				},
			}
		})

		Describe("Un/MarshalPacked", func() {
			var (
				e2 EventStream
			)

			BeforeEach(func() {
				e2 = EventStream{
					{
						Timestamp: EventTime{et},
						Record: map[string]interface{}{
							"foo":    "bar",
							"george": "jungle",
						},
					},
					{
						Timestamp: EventTime{et},
						Record: map[string]interface{}{
							"foo":    "kablooie",
							"george": "frank",
						},
					},
				}
			})

			It("Can marshal and unmarshal packed entries", func() {
				b, err := e2.MarshalPacked()
				Expect(err).ToNot(HaveOccurred())

				el := EventStream{}
				_, err = el.UnmarshalPacked(b)
				Expect(err).ToNot(HaveOccurred())
				Expect(el.Equal(e2)).To(BeTrue())
			})
		})

		Describe("Equal", func() {
			var (
				e2 EventStream
			)

			BeforeEach(func() {
				e2 = EventStream{
					{
						Timestamp: EventTime{et},
						Record: map[string]interface{}{
							"foo":    "bar",
							"george": "jungle",
						},
					},
					{
						Timestamp: EventTime{et},
						Record: map[string]interface{}{
							"foo":    "kablooie",
							"george": "frank",
						},
					},
				}
			})

			It("Returns true", func() {
				Expect(e1.Equal(e2)).To(BeTrue())
			})

			Context("When the lists have different element counts", func() {
				BeforeEach(func() {
					e2 = e2[:1]
				})

				It("Returns false", func() {
					Expect(e1.Equal(e2)).To(BeFalse())
				})
			})

			Context("When the lists have differing elements", func() {
				BeforeEach(func() {
					e2[0].Timestamp = EventTime{et.Add(5 * time.Second)}
				})

				It("Returns false", func() {
					Expect(e1.Equal(e2)).To(BeFalse())
				})
			})
		})
	})

	Describe("NewPackedForwardMessage", func() {
		var (
			tag     string
			entries EventStream
		)

		BeforeEach(func() {
			tag = "foo.bar"
			entries = EventStream{
				{
					Timestamp: EventTime{time.Now()},
					Record: map[string]interface{}{
						"foo":    "bar",
						"george": "jungle",
					},
				},
				{
					Timestamp: EventTime{time.Now()},
					Record: map[string]interface{}{
						"foo":    "kablooie",
						"george": "frank",
					},
				},
			}
		})

		It("Returns a PackedForwardMessage", func() {
			msg, err := NewPackedForwardMessage(tag, entries)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).NotTo(BeNil())
			Expect(*msg.Options.Size).To(Equal(len(entries)))
			Expect(msg.Options.Compressed).To(BeEmpty())
		})

		It("Correctly encodes the entries into a bytestream", func() {
			msg, err := NewPackedForwardMessage(tag, entries)
			Expect(err).NotTo(HaveOccurred())
			elist := make(EventStream, 2)
			_, err = elist.UnmarshalPacked(msg.EventStream)
			Expect(err).NotTo(HaveOccurred())
			Expect(elist.Equal(entries)).To(BeTrue())
		})
	})

	Describe("NewCompressedPackedForwardMessage", func() {
		var (
			tag     string
			entries []EventExt
		)

		BeforeEach(func() {
			tag = "foo.bar"
			entries = []EventExt{
				{
					Timestamp: EventTime{time.Now()},
					Record: map[string]interface{}{
						"foo":    "bar",
						"george": "jungle",
					},
				},
				{
					Timestamp: EventTime{time.Now()},
					Record: map[string]interface{}{
						"foo":    "kablooie",
						"george": "frank",
					},
				},
			}
		})

		It("Returns a message with a gzip-compressed event stream", func() {
			msg, err := NewCompressedPackedForwardMessage(tag, entries)
			Expect(err).ToNot(HaveOccurred())
			Expect(msg).NotTo(BeNil())
			Expect(*msg.Options.Size).To(Equal(len(entries)))
			Expect(msg.Options.Compressed).To(Equal("gzip"))
		})
	})
})
