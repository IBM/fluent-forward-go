package client_test

import (
	"bytes"
	"errors"

	// "fmt"
	// "io"

	// // "os"
	"time"

	. "github.com/IBM/fluent-forward-go/fluent/client"
	"github.com/IBM/fluent-forward-go/fluent/client/clientfakes"
	"github.com/IBM/fluent-forward-go/fluent/client/ws/ext"
	"github.com/IBM/fluent-forward-go/fluent/client/ws/ext/extfakes"
	"github.com/IBM/fluent-forward-go/fluent/client/ws/wsfakes"
	"github.com/IBM/fluent-forward-go/fluent/protocol"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IAMAuthInfo", func() {
	It("gets and sets an IAM token", func() {
		iai := NewIAMAuthInfo("a")
		Expect(iai.IAMToken()).To(Equal("a"))
		iai.SetIAMToken("b")
		Expect(iai.IAMToken()).To(Equal("b"))
	})
})

var _ = Describe("WSClient", func() {
	var (
		factory    *clientfakes.FakeWSConnectionFactory
		client     *WSClient
		clientSide ext.Conn
		conn       *wsfakes.FakeConnection
	)

	BeforeEach(func() {
		factory = &clientfakes.FakeWSConnectionFactory{}
		client = &WSClient{
			ConnectionFactory: factory,
		}
		clientSide = &extfakes.FakeConn{}
		conn = &wsfakes.FakeConnection{}

		Expect(factory.NewCallCount()).To(Equal(0))
		Expect(client.Session).To(BeNil())
	})

	JustBeforeEach(func() {
		factory.NewReturns(clientSide, nil)
	})

	Describe("Connect", func() {
		It("Does not return an error", func() {
			Expect(client.Connect()).ToNot(HaveOccurred())
		})

		It("Gets the connection from the ConnectionFactory", func() {
			client.Connect()
			Expect(factory.NewCallCount()).To(Equal(1))
		})

		It("Stores the connection in the Session", func() {
			client.Connect()
			Expect(client.Session).ToNot(BeNil())
			Expect(client.Session.Connection).ToNot(BeNil())
		})

		When("the factory returns an error", func() {
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

		When("the factory returns an error", func() {
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

	Describe("Disconnect", func() {
		When("the session is not nil", func() {
			JustBeforeEach(func() {
				client.Connect()
				client.Session.Connection = conn
			})

			It("closes the connection", func() {
				Expect(client.Disconnect()).ToNot(HaveOccurred())
				Expect(conn.CloseCallCount()).To(Equal(1))
			})
		})

		When("the session is nil", func() {
			JustBeforeEach(func() {
				client.Session = nil
			})

			It("does not error or panic", func() {
				Ω(func() {
					Expect(client.Disconnect()).ToNot(HaveOccurred())
				}).ShouldNot(Panic())
			})
		})
	})

	Describe("Reconnect", func() {
		var (
			session1 *WSSession
		)

		JustBeforeEach(func() {
			client.Connect()
			client.Session.Connection = conn
			session1 = client.Session
		})

		It("calls Disconnect and creates a new Session", func() {
			Expect(client.Reconnect()).ToNot(HaveOccurred())

			Expect(conn.CloseCallCount()).To(Equal(1))

			Expect(client.Session).ToNot(BeNil())
			Expect(client.Session).ToNot(Equal(session1))
			Expect(client.Session.Connection).ToNot(BeNil())
			Expect(client.Session.Connection).ToNot(Equal(conn))
		})
	})

	Describe("Listen", func() {
		JustBeforeEach(func() {
			client.Connect()
			client.Session.Connection = conn
		})

		It("listens on the Session connection", func() {
			Expect(client.Listen()).ToNot(HaveOccurred())
			Expect(conn.ListenCallCount()).To(Equal(1))
		})
	})

	Describe("SendMessage", func() {
		var (
			msg protocol.MessageExt
		)

		BeforeEach(func() {
			msg = protocol.MessageExt{
				Tag:       "foo.bar",
				Timestamp: protocol.EventTime{time.Now()},
				Record:    map[string]interface{}{},
				Options:   &protocol.MessageOptions{},
			}
		})

		JustBeforeEach(func() {
			err := client.Connect()
			client.Session.Connection = conn
			Expect(err).ToNot(HaveOccurred())
		})

		It("Sends the message", func() {
			bits, _ := msg.MarshalMsg(nil)
			Expect(client.SendMessage(&msg)).ToNot(HaveOccurred())

			writtenbits := conn.WriteArgsForCall(0)
			Expect(bytes.Equal(bits, writtenbits)).To(BeTrue())
		})
	})
})
