package protocol_test

import (
	"encoding/base64"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/IBM/fluent-forward-go/fluent/protocol"
	. "github.com/IBM/fluent-forward-go/fluent/protocol"
)

var _ = Describe("Chunk", func() {
	Describe("GetChunk", func() {

		It("returns the chunk ID for a byte-encoded message", func() {
			msg := protocol.Message{}
			expected, err := msg.Chunk()
			Expect(err).ToNot(HaveOccurred())

			bits, err := msg.MarshalMsg(nil)
			Expect(err).ToNot(HaveOccurred())

			chunk, err := GetChunk(bits)
			Expect(err).ToNot(HaveOccurred())
			Expect(chunk).To(Equal(expected))
		})

		It("returns an error when chunk is not found", func() {
			msg := protocol.Message{}

			bits, err := msg.MarshalMsg(nil)
			Expect(err).ToNot(HaveOccurred())

			_, err = GetChunk(bits)
			Expect(err.Error()).To(ContainSubstring("chunk not found"))
		})
	})

	Describe("Messages", func() {
		When("Chunk is called", func() {
			It("works as expected", func() {
				msg := &protocol.Message{}
				msg.Chunk()
				bits, _ := msg.MarshalMsg(nil)
				raw := protocol.RawMessage(bits)

				for _, ce := range []protocol.ChunkEncoder{raw, &protocol.Message{}, &protocol.MessageExt{}, &protocol.ForwardMessage{}, &protocol.PackedForwardMessage{}} {
					chunk, err := ce.Chunk()
					Expect(err).ToNot(HaveOccurred())
					Expect(chunk).ToNot(BeEmpty())
					chunk2, err := ce.Chunk()
					Expect(err).ToNot(HaveOccurred())
					Expect(chunk).To(Equal(chunk2))

					b, err := base64.StdEncoding.DecodeString(chunk)
					Expect(err).ToNot(HaveOccurred())
					Expect(b).ToNot(BeEmpty())
				}
			})
		})
	})
})
