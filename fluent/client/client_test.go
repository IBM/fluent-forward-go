package client_test

import (
	"errors"
	// "fmt"
	// "io"
	"math/rand"
	"net"

	// // "os"
	"time"

	. "github.com/IBM/fluent-forward-go/fluent/client"
	"github.com/IBM/fluent-forward-go/fluent/client/clientfakes"
	"github.com/IBM/fluent-forward-go/fluent/protocol"
	. "github.com/onsi/ginkgo"
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
		client = &Client{
			ConnectionFactory: factory,
			Timeout:           2 * time.Second,
		}

		clientSide, _ = net.Pipe()

		Expect(factory.NewCallCount()).To(Equal(0))
		Expect(client.Session).To(BeNil())
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
			Expect(err).To(BeNil())
			Expect(factory.NewCallCount()).To(Equal(1))
		})

		It("Stores the connection in the Session", func() {
			client.Connect()
			Expect(client.Session).NotTo(BeNil())
			Expect(client.Session.Connection).To(Equal(clientSide))
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

	Describe("SendMessage", func() {
		var (
			serverSide net.Conn
			msg        protocol.MessageExt
		)

		BeforeEach(func() {
			clientSide, serverSide = net.Pipe()
			msg = protocol.MessageExt{
				Tag:       "foo.bar",
				Timestamp: protocol.EventTime{time.Now()},
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
			go func() {
				client.SendMessage(&msg)
			}()
			var recvd protocol.MessageExt
			recvd.DecodeMsg(msgp.NewReader(serverSide))

			Expect(recvd.Tag).To(Equal(msg.Tag))
			Expect(recvd.Options).To(Equal(msg.Options))
			Expect(recvd.Timestamp.Equal(msg.Timestamp.Time)).To(BeTrue())
			Expect(recvd.Record["first"]).To(Equal(msg.Record["first"]))
			Expect(recvd.Record["last"]).To(Equal(msg.Record["last"]))
		})

		Context("When the Session is not yet in Transport phase (handshake not performed)", func() {
			JustBeforeEach(func() {
				client.Session.TransportPhase = false
			})

			It("Returns an error", func() {
				Expect(client.SendMessage(&msg)).To(HaveOccurred())
			})

			// TODO: We need a test that no message is sent
		})
	})

	XDescribe("Handshake", func() {
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
			go func() {
				defer GinkgoRecover()
				err := client.Handshake()
				Expect(err).NotTo(HaveOccurred())
				Expect(client.Session.TransportPhase).To(BeTrue())
			}()

			helo.EncodeMsg(serverWriter)
			serverWriter.Flush()

			ping.DecodeMsg(serverReader)
			Expect(ping.MessageType).To(Equal(protocol.MSGTYPE_PING))
			Expect(protocol.ValidatePingDigest(&ping, sharedKey, nonce)).NotTo(HaveOccurred())

			pong, err := protocol.NewPong(true, "", "", sharedKey, helo, &ping)
			Expect(pong).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())

			pong.EncodeMsg(serverWriter)
			serverWriter.Flush()
		})

		Context("When the client is not currently connected", func() {
			JustBeforeEach(func() {
				client.Disconnect()
				Expect(client.Session).To(BeNil())
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

					helo.EncodeMsg(serverWriter)
					serverWriter.Flush()

					ping.DecodeMsg(serverReader)
					Expect(ping.MessageType).To(Equal(protocol.MSGTYPE_PING))
					Expect(protocol.ValidatePingDigest(&ping, sharedKey, nonce)).To(HaveOccurred())
				})
			})
		})
	})
})
