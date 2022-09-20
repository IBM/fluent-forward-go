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

package protocol_test

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/IBM/fluent-forward-go/fluent/protocol"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tinylib/msgp/msgp"
	// . "github.com/IBM/fluent-forward-go/fluent/protocol"
)

var _ = Describe("ForwardMessage", func() {
	var (
		fwdmsg *protocol.ForwardMessage
	)

	BeforeEach(func() {
		bits, err := ioutil.ReadFile("protocolfakes/forwarded_records.msgpack.bin")
		Expect(err).ToNot(HaveOccurred())
		fwdmsg = &protocol.ForwardMessage{}
		_, err = fwdmsg.UnmarshalMsg(bits)
		Expect(err).NotTo(HaveOccurred())
		entries := []protocol.EntryExt{
			{
				Timestamp: protocol.EventTime{time.Now()},
				Record: map[string]interface{}{
					"foo":    "bar",
					"george": "jungle",
				},
			},
			{
				Timestamp: protocol.EventTime{time.Now()},
				Record: map[string]interface{}{
					"foo":    "kablooie",
					"george": "frank",
				},
			},
		}

		fwdmsg = protocol.NewForwardMessage("foo", entries)
		Expect(*fwdmsg.Options.Size).To(Equal(len(entries)))
	})

	Describe("Unmarshaling", func() {
		testMarshalling := func(msg *protocol.ForwardMessage, opts *protocol.MessageOptions) {
			msg.Options = opts
			b, err := msg.MarshalMsg(nil)
			Expect(err).NotTo(HaveOccurred())

			var unmfwd protocol.ForwardMessage
			_, err = unmfwd.UnmarshalMsg(b)
			Expect(err).NotTo(HaveOccurred())

			if opts == nil {
				Expect(unmfwd.Options).To(BeNil())
			} else {
				Expect(unmfwd.Options).ToNot(BeNil())
			}
			Expect(unmfwd.Tag).To(Equal("foo"))
			Expect(unmfwd.Entries[0].Timestamp.Time.Equal(msg.Entries[0].Timestamp.Time)).To(BeTrue())
			Expect(unmfwd.Entries[0].Record).To(HaveKeyWithValue("foo", "bar"))
			Expect(unmfwd.Entries[0].Record).To(HaveKeyWithValue("george", "jungle"))
			Expect(unmfwd.Entries[1].Timestamp.Time.Equal(msg.Entries[1].Timestamp.Time)).To(BeTrue())
			Expect(unmfwd.Entries[1].Record).To(HaveKeyWithValue("foo", "kablooie"))
			Expect(unmfwd.Entries[1].Record).To(HaveKeyWithValue("george", "frank"))
		}

		It("Marshals and unmarshals correctly", func() {
			testMarshalling(fwdmsg, nil)
			testMarshalling(fwdmsg, &protocol.MessageOptions{})
		})

		testEncodingDecoding := func(msg *protocol.ForwardMessage, opts *protocol.MessageOptions) {
			var buf bytes.Buffer
			en := msgp.NewWriter(&buf)

			msg.Options = opts
			err := msg.EncodeMsg(en)
			Expect(err).NotTo(HaveOccurred())
			en.Flush()

			var unmfwd protocol.ForwardMessage
			re := msgp.NewReader(&buf)
			err = unmfwd.DecodeMsg(re)
			Expect(err).NotTo(HaveOccurred())

			if opts == nil {
				Expect(unmfwd.Options).To(BeNil())
			} else {
				Expect(unmfwd.Options).ToNot(BeNil())
			}
			Expect(unmfwd.Tag).To(Equal("foo"))
			Expect(unmfwd.Entries[0].Timestamp.Time.Equal(msg.Entries[0].Timestamp.Time)).To(BeTrue())
			Expect(unmfwd.Entries[0].Record).To(HaveKeyWithValue("foo", "bar"))
			Expect(unmfwd.Entries[0].Record).To(HaveKeyWithValue("george", "jungle"))
			Expect(unmfwd.Entries[1].Timestamp.Time.Equal(msg.Entries[1].Timestamp.Time)).To(BeTrue())
			Expect(unmfwd.Entries[1].Record).To(HaveKeyWithValue("foo", "kablooie"))
			Expect(unmfwd.Entries[1].Record).To(HaveKeyWithValue("george", "frank"))
		}

		It("Encodes and decodes correctly", func() {
			testEncodingDecoding(fwdmsg, nil)
			testEncodingDecoding(fwdmsg, &protocol.MessageOptions{})
		})

		It("Properly deserializes real fluentbit messages with no options", func() {
			bits, err := ioutil.ReadFile("protocolfakes/forwarded_records.msgpack.bin")
			Expect(err).ToNot(HaveOccurred())

			fwdmsg := protocol.ForwardMessage{}
			_, err = fwdmsg.UnmarshalMsg(bits)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
