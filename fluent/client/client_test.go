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

package client_test

import (
	"errors"
	"math/rand"
	"net"
	"reflect"
	"time"

	. "github.com/IBM/fluent-forward-go/fluent/client"
	"github.com/IBM/fluent-forward-go/fluent/client/clientfakes"
	"github.com/IBM/fluent-forward-go/fluent/protocol"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tinylib/msgp/msgp"
)

var _ = Describe("Client", func() {
	var (
		factory    *clientfakes.FakeConnectionFactory
		client     *Client
		clientSide net.Conn
	)

	BeforeEach(func() {
		factory = &clientfakes.FakeConnectionFactory{}

		opts := ConnectionOptions{
			Factory:     factory,
			DialTimeout: 2 * time.Second,
		}

		client = New(opts)

		clientSide, _ = net.Pipe()

		Expect(factory.NewCallCount()).To(Equal(0))
		Expect(client.TransportPhase()).To(BeFalse())
	})

	JustBeforeEach(func() {
		factory.NewReturns(clientSide, nil)
	})

	Describe("Connect", func() {
		It("Does not return an error", func() {
			Expect(client.Connect()).NotTo(HaveOccurred())
		})

		It("Gets the connection from the ConnectionFactory", func() {
			err := client.Connect()
			Expect(err).NotTo(HaveOccurred())
			Expect(factory.NewCallCount()).To(Equal(1))
		})

		It("Completes the handshake", func() {
			err := client.Connect()
			Expect(err).NotTo(HaveOccurred())
			Expect(client.TransportPhase()).To(BeTrue())
		})

		It("errors if session already exists", func() {
			err := client.Connect()
			Expect(err).NotTo(HaveOccurred())
			err = client.Connect()
			Expect(err.Error()).To(Equal("a session is already active"))
		})

		Context("When the factory returns an error", func() {
			var (
				connectionError error
			)

			JustBeforeEach(func() {
				connectionError = errors.New("Nope")
				factory.NewReturns(nil, connectionError)
			})

			It("Returns an error", func() {
				err := client.Connect()
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeIdenticalTo(connectionError))
			})
		})
	})

	Describe("Reconnect", func() {
		JustBeforeEach(func() {
			err := client.Connect()
			Expect(err).NotTo(HaveOccurred())
		})

		It("calls Disconnect and creates a new Session", func() {
			Expect(client.Reconnect()).ToNot(HaveOccurred())
			Expect(factory.NewCallCount()).To(Equal(2))
		})

		It("works if no active session", func() {
			Expect(client.Disconnect()).ToNot(HaveOccurred())

			Expect(client.Reconnect()).ToNot(HaveOccurred())
			Expect(factory.NewCallCount()).To(Equal(2))
		})
	})

	Describe("Send", func() {
		var (
			serverSide net.Conn
			msg        protocol.MessageExt
		)

		BeforeEach(func() {
			clientSide, serverSide = net.Pipe()
			msg = protocol.MessageExt{
				Tag:       "foo.bar",
				Timestamp: protocol.EventTimeNow(), //nolint
				Record: map[string]interface{}{
					"first": "Eddie",
					"last":  "Van Halen",
				},
				Options: &protocol.MessageOptions{},
			}
		})

		JustBeforeEach(func() {
			err := client.Connect()
			Expect(err).NotTo(HaveOccurred())
		})

		It("Sends the message", func() {
			c := make(chan bool, 1)
			go func() {
				defer GinkgoRecover()

				c <- true
				err := client.Send(&msg)
				Expect(err).NotTo(HaveOccurred())
			}()

			var recvd protocol.MessageExt
			<-c
			err := recvd.DecodeMsg(msgp.NewReader(serverSide))
			Expect(err).NotTo(HaveOccurred())

			Expect(recvd.Tag).To(Equal(msg.Tag))
			Expect(recvd.Options).To(Equal(msg.Options))
			Expect(recvd.Timestamp.Equal(msg.Timestamp.Time)).To(BeTrue())
			Expect(reflect.DeepEqual(recvd.Record, msg.Record)).To(BeTrue())
		})

		Context("When the Session is not yet in Transport phase (handshake not performed)", func() {
			JustBeforeEach(func() {
				client.Disconnect()
			})

			It("Returns an error", func() {
				Expect(client.Send(&msg)).To(HaveOccurred())
			})

			// TODO: We need a test that no message is sent
		})

		Context("When WriteDeadline is set and write takes too long", func() {
			BeforeEach(func() {
				opts := ConnectionOptions{
					Factory:      factory,
					DialTimeout:  2 * time.Second,
					WriteTimeout: 100 * time.Millisecond, // short timeout
					RequireAck:   false,
				}
				client = New(opts)
				clientSide, serverSide = net.Pipe()
				factory.NewReturns(clientSide, nil)
			})

			It("returns a write timeout error", func() {
				go func() {
					defer GinkgoRecover()
					time.Sleep(200 * time.Millisecond) // this is longer than WriteTimeout
					buf := make([]byte, 1024)
					_, err := serverSide.Read(buf)
					Expect(err).NotTo(HaveOccurred())
				}()

				err := client.Send(&msg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("write pipe: i/o timeout"))
			})
		})

		Context("When ReadDeadline is set and acknowledgment is delayed", func() {
			BeforeEach(func() {
				opts := ConnectionOptions{
					Factory:     factory,
					DialTimeout: 2 * time.Second,
					ReadTimeout: 100 * time.Millisecond, // short timeout
					RequireAck:  true,                   // waig for acknowledgment
				}
				client = New(opts)
				clientSide, serverSide = net.Pipe()
				factory.NewReturns(clientSide, nil)
			})

			It("returns a read timeout error", func() {
				done := make(chan bool) // used to synchronize test
				go func() {
					defer GinkgoRecover()
					defer func() {
						done <- true
					}()
					err := client.Send(&msg)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("read pipe: i/o timeout"))
				}()

				// Server reads the message
				rcvd := &protocol.MessageExt{}
				err := rcvd.DecodeMsg(msgp.NewReader(serverSide))
				Expect(err).NotTo(HaveOccurred())

				// delay sending the ack (longer than ReadTimeout)
				time.Sleep(200 * time.Millisecond)

				// send the ack
				chunk := rcvd.Options.Chunk
				ack := &protocol.AckMessage{Ack: chunk}
				err = ack.EncodeMsg(msgp.NewWriter(serverSide))
				Expect(err).NotTo(HaveOccurred())

				<-done
			})
		})

		Context("RequireAck is true", func() {
			var (
				serverSide   net.Conn
				serverWriter *msgp.Writer
				serverReader *msgp.Reader
				msg          protocol.MessageExt
			)

			BeforeEach(func() {
				clientSide, serverSide = net.Pipe()
				msg = protocol.MessageExt{
					Tag: "foo.bar",
				}
				serverWriter = msgp.NewWriter(serverSide)
				serverReader = msgp.NewReader(serverSide)
			})

			JustBeforeEach(func() {
				client.RequireAck = true
			})

			It("chunks the message and waits for the ack", func() {
				done := make(chan bool)
				Expect(msg.Options).To(BeNil())
				go func() {
					defer GinkgoRecover()
					defer func() { done <- true }()
					err := client.Send(&msg)
					Expect(err).ToNot(HaveOccurred())
				}()

				msgext := &protocol.MessageExt{}
				err := msgext.DecodeMsg(serverReader)
				Expect(err).ToNot(HaveOccurred())

				chunk := msgext.Options.Chunk
				Expect(chunk).ToNot(BeEmpty())
				Expect(chunk).To(Equal(msg.Options.Chunk))

				ack := &protocol.AckMessage{Ack: chunk}
				err = ack.EncodeMsg(serverWriter)
				Expect(err).ToNot(HaveOccurred())
				serverWriter.Flush()

				<-done
			})

			It("returns an error when the ack is bad", func() {
				done := make(chan bool)
				Expect(msg.Options).To(BeNil())
				go func() {
					defer GinkgoRecover()
					defer func() { done <- true }()
					err := client.Send(&msg)
					Expect(err.Error()).To(ContainSubstring("Expected chunk"))
				}()

				msgext := &protocol.MessageExt{}
				serverSide.SetReadDeadline(time.Now().Add(time.Second))
				err := msgext.DecodeMsg(serverReader)
				Expect(err).ToNot(HaveOccurred())

				chunk := msgext.Options.Chunk
				Expect(chunk).ToNot(BeEmpty())
				Expect(chunk).To(Equal(msg.Options.Chunk))

				ack := &protocol.AckMessage{Ack: ""}
				err = ack.EncodeMsg(serverWriter)
				Expect(err).ToNot(HaveOccurred())
				serverWriter.Flush()

				<-done
			})
		})
	})

	Describe("Send*", func() {
		type msgSender struct {
			tag     string
			doSend  func() error
			decoder msgp.Decodable
		}
		var (
			serverSide net.Conn
			msg        protocol.MessageExt
			bits       []byte
			el         protocol.EntryList
		)

		BeforeEach(func() {
			clientSide, serverSide = net.Pipe()

			msg = protocol.MessageExt{
				Record: map[string]string{
					"a": "b",
				},
				Tag: "msg",
			}
			var err error
			bits, err = msg.MarshalMsg(nil)
			Expect(err).NotTo(HaveOccurred())

			el = protocol.EntryList{
				{
					Record: map[string]interface{}{
						"foo":    "bar",
						"george": "jungle",
					},
				},
				{
					Record: map[string]interface{}{
						"foo":    "kablooie",
						"george": "frank",
					},
				},
			}
		})

		JustBeforeEach(func() {
			err := client.Connect()
			Expect(err).NotTo(HaveOccurred())
		})

		doTest := func(sndr msgSender) {
			c := make(chan bool, 1)
			go func() {
				defer GinkgoRecover()

				c <- true
				err := sndr.doSend()
				Expect(err).NotTo(HaveOccurred())
			}()

			<-c
			err := sndr.decoder.DecodeMsg(msgp.NewReader(serverSide))
			Expect(err).NotTo(HaveOccurred())
			val := reflect.Indirect(reflect.ValueOf(sndr.decoder))
			Expect(val.FieldByName("Tag").String()).To(Equal(sndr.tag))
		}

		Context("SendForward", func() {
			It("works", func() {
				doTest(msgSender{
					tag:     "fwd",
					decoder: &protocol.ForwardMessage{},
					doSend: func() error {
						return client.SendForward("fwd", el)
					},
				})
			})
		})

		Context("SendCompressed", func() {
			It("works", func() {
				doTest(msgSender{
					tag:     "cmp",
					decoder: &protocol.PackedForwardMessage{},
					doSend: func() error {
						return client.SendCompressed("cmp", el)
					},
				})
			})
		})

		Context("SendCompressedFromBytes", func() {
			It("works", func() {
				doTest(msgSender{
					tag:     "cmpfb",
					decoder: &protocol.PackedForwardMessage{},
					doSend: func() error {
						return client.SendCompressedFromBytes("cmpfb", bits)
					},
				})
			})
		})

		Context("SendMessage", func() {
			It("works", func() {
				doTest(msgSender{
					tag:     "sm",
					decoder: &protocol.Message{},
					doSend: func() error {
						return client.SendMessage("sm", el[0].Record)
					},
				})
			})
		})

		Context("SendMessageExt", func() {
			It("works", func() {
				doTest(msgSender{
					tag:     "ext",
					decoder: &protocol.MessageExt{},
					doSend: func() error {
						return client.SendMessageExt("ext", el[1].Record)
					},
				})
			})
		})

		Context("SendPacked", func() {
			It("works", func() {
				doTest(msgSender{
					tag:     "pkd",
					decoder: &protocol.PackedForwardMessage{},
					doSend: func() error {
						return client.SendPacked("pkd", el)
					},
				})
			})
		})

		Context("SendPackedFromBytes", func() {
			It("works", func() {
				doTest(msgSender{
					tag:     "pkdfb",
					decoder: &protocol.PackedForwardMessage{},
					doSend: func() error {
						return client.SendPackedFromBytes("pkdfb", bits)
					},
				})
			})
		})

		Context("SendRaw", func() {
			It("works", func() {
				doTest(msgSender{
					tag:     "msg",
					decoder: &protocol.MessageExt{},
					doSend: func() error {
						return client.SendRaw(bits)
					},
				})
			})

			Context("When the Session is not yet in Transport phase (handshake not performed)", func() {
				JustBeforeEach(func() {
					client.Disconnect()
				})

				It("Returns an error", func() {
					Expect(client.SendRaw(bits)).To(HaveOccurred())
				})
			})
		})
	})

	Describe("Handshake", func() {
		var (
			serverSide       net.Conn
			serverWriter     *msgp.Writer
			serverReader     *msgp.Reader
			helo             *protocol.Helo
			ping             protocol.Ping
			sharedKey, nonce []byte
		)

		BeforeEach(func() {
			clientSide, serverSide = net.Pipe()
			Expect(clientSide).NotTo(BeNil())
			Expect(serverSide).NotTo(BeNil())
			serverWriter = msgp.NewWriter(serverSide)
			serverReader = msgp.NewReader(serverSide)

			sharedKey = []byte(`thisisasharedkey`)
			client.AuthInfo.SharedKey = sharedKey

			nonce = make([]byte, 16)
			numb, err := rand.Read(nonce)
			Expect(err).NotTo(HaveOccurred())
			Expect(numb).To(Equal(len(nonce)))

			helo = protocol.NewHelo(&protocol.HeloOpts{
				Nonce: nonce,
			})
		})

		JustBeforeEach(func() {
			err := client.Connect()
			Expect(err).NotTo(HaveOccurred())
		})

		It("Completes the handshake", func() {
			hs := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				<-hs
				err := client.Handshake()
				Expect(err).NotTo(HaveOccurred())
				Expect(client.TransportPhase()).To(BeTrue())
				hs <- struct{}{}
			}()

			hs <- struct{}{}
			err := helo.EncodeMsg(serverWriter)
			Expect(err).NotTo(HaveOccurred())
			serverWriter.Flush()

			err = ping.DecodeMsg(serverReader)
			Expect(err).NotTo(HaveOccurred())
			Expect(ping.MessageType).To(Equal(protocol.MsgTypePing))
			Expect(protocol.ValidatePingDigest(&ping, sharedKey, nonce)).NotTo(HaveOccurred())

			pong, err := protocol.NewPong(true, "", "", sharedKey, helo, &ping)
			Expect(pong).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())

			err = pong.EncodeMsg(serverWriter)
			Expect(err).NotTo(HaveOccurred())
			serverWriter.Flush()
			<-hs
		})

		Context("When the client is not currently connected", func() {
			JustBeforeEach(func() {
				err := client.Disconnect()
				Expect(err).NotTo(HaveOccurred())
				Expect(client.TransportPhase()).To(BeFalse())
			})

			It("Returns an error", func() {
				err := client.Handshake()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Authentication failures", func() {
			Context("When the client has the wrong shared key", func() {
				BeforeEach(func() {
					client.AuthInfo.SharedKey = []byte(`thisisthewrongkey`)
				})

				// We'll treat this as an auth failure and have a slightly different behavior
				// - we'll need to detect the auth failure on the client side and return
				// a useful error.
				It("Sends an bad digest", func() {
					go func() {
						defer GinkgoRecover()
						err := client.Handshake()
						Expect(err).NotTo(HaveOccurred())
					}()

					err := helo.EncodeMsg(serverWriter)
					Expect(err).NotTo(HaveOccurred())
					serverWriter.Flush()

					err = ping.DecodeMsg(serverReader)
					Expect(err).NotTo(HaveOccurred())
					Expect(ping.MessageType).To(Equal(protocol.MsgTypePing))
					Expect(protocol.ValidatePingDigest(&ping, sharedKey, nonce)).To(HaveOccurred())
				})
			})
		})
	})
})
