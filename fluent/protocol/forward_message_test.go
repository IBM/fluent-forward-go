package protocol_test

import (
	"bytes"
	"io/ioutil"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tinylib/msgp/msgp"

	. "github.ibm.com/Observability/fluent-forward-go/fluent/protocol"
)

var _ = Describe("ForwardMessage", func() {
	var (
		fwdmsg ForwardMessage
	)

	BeforeEach(func() {
		bits, err := ioutil.ReadFile("protocolfakes/forwarded_records.msgpack.bin")
		Expect(err).ToNot(HaveOccurred())
		_, err = fwdmsg.UnmarshalMsg(bits)
		Expect(err).NotTo(HaveOccurred())
		fwdmsg = ForwardMessage{
			Tag: "foo",
			Entries: []EntryExt{
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
			},
		}
	})

	Describe("Unmarshaling", func() {
		testMarshalling := func(msg ForwardMessage, opts *MessageOptions) {
			msg.Options = opts
			b, err := msg.MarshalMsg(nil)
			Expect(err).NotTo(HaveOccurred())

			var unmfwd ForwardMessage
			_, err = unmfwd.UnmarshalMsg(b)
			Expect(err).NotTo(HaveOccurred())

			if opts == nil {
				Expect(unmfwd.Options).To(BeNil())
			} else {
				Expect(unmfwd.Options).ToNot(BeNil())
			}
			Expect(msg.Entries[0].Record).To(HaveKeyWithValue("foo", "bar"))
			Expect(msg.Entries[0].Record).To(HaveKeyWithValue("george", "jungle"))
			Expect(msg.Entries[1].Record).To(HaveKeyWithValue("foo", "kablooie"))
			Expect(msg.Entries[1].Record).To(HaveKeyWithValue("george", "frank"))
		}

		It("Marshals and unmarshals correctly", func() {
			// Uncomment after adding custom encoding
			// testMarshalling(fwdmsg, nil)
			testMarshalling(fwdmsg, &MessageOptions{})
		})

		testEncodingDecoding := func(msg ForwardMessage, opts *MessageOptions) {
			var buf bytes.Buffer
			en := msgp.NewWriter(&buf)

			msg.Options = opts
			err := msg.EncodeMsg(en)
			Expect(err).NotTo(HaveOccurred())
			en.Flush()

			var unmfwd ForwardMessage
			re := msgp.NewReader(&buf)
			err = unmfwd.DecodeMsg(re)
			Expect(err).NotTo(HaveOccurred())

			if opts == nil {
				Expect(unmfwd.Options).To(BeNil())
			} else {
				Expect(unmfwd.Options).ToNot(BeNil())
			}
			Expect(unmfwd.Entries[0].Record).To(HaveKeyWithValue("foo", "bar"))
			Expect(unmfwd.Entries[0].Record).To(HaveKeyWithValue("george", "jungle"))
			Expect(unmfwd.Entries[1].Record).To(HaveKeyWithValue("foo", "kablooie"))
			Expect(unmfwd.Entries[1].Record).To(HaveKeyWithValue("george", "frank"))
		}

		It("Marshals and unmarshals correctly", func() {
			// Uncomment after adding custom encoding
			// testEncodingDecoding(fwdmsg, nil)
			testEncodingDecoding(fwdmsg, &MessageOptions{})
		})
	})

	It("Properly deserializes real fluentbit messages with no options", func() {
		bits, err := ioutil.ReadFile("protocolfakes/forwarded_records.msgpack.bin")
		Expect(err).ToNot(HaveOccurred())

		fwdmsg := ForwardMessage{}
		_, err = fwdmsg.UnmarshalMsg(bits)
		Expect(err).NotTo(HaveOccurred())
	})
})
