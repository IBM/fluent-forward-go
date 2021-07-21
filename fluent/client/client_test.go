package client_test

import (
	"errors"
	// "fmt"
	// "io"
	// "math/rand"
	"net"
	// // "os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tinylib/msgp/msgp"
	. "github.ibm.com/Observability/fluent-forward-go/fluent/client"
	"github.ibm.com/Observability/fluent-forward-go/fluent/client/clientfakes"
	"github.ibm.com/Observability/fluent-forward-go/fluent/protocol"
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
			client.Connect()
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
			msg        protocol.Message
		)

		BeforeEach(func() {
			clientSide, serverSide = net.Pipe()
			msg = protocol.Message{
				Tag: "foo.bar",
				Entry: protocol.Entry{
					Timestamp: protocol.EventTime{time.Now()},
					Record: map[string]string{
						"first": "Eddie",
						"last":  "Van Halen",
					},
				},
				Options: protocol.MessageOptions{},
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
			var recvd protocol.Message
			recvd.DecodeMsg(msgp.NewReader(serverSide))

			Expect(recvd.Tag).To(Equal(msg.Tag))
			Expect(recvd.Options).To(Equal(msg.Options))
			Expect(recvd.Entry.Timestamp.Equal(msg.Entry.Timestamp.Time)).To(BeTrue())
			Expect(recvd.Entry.Record["first"]).To(Equal(msg.Entry.Record["first"]))
			Expect(recvd.Entry.Record["last"]).To(Equal(msg.Entry.Record["last"]))
		})
	})

	// XDescribe("Network Management", func() {
	// 	var (
	// 		server     net.Listener
	// 		serverErr  error
	// 		hostname   string
	// 		port       int
	// 		handleConn func(net.Conn)
	// 	)
	//
	// 	BeforeEach(func() {
	// 		server, serverErr = net.Listen("tcp", ":0")
	// 		Expect(serverErr).NotTo(HaveOccurred())
	//
	// 		port = server.Addr().(*net.TCPAddr).Port
	//
	// 		handleConn = func(net.Conn) {
	// 			fmt.Fprintln(os.Stderr, "HANDLING CONNECTION")
	// 			return
	// 		}
	// 	})
	//
	// 	JustBeforeEach(func() {
	// 		go func() {
	// 			defer GinkgoRecover()
	// 			for {
	// 				conn, err := server.Accept()
	// 				if err != nil {
	// 					Fail("Failure accepting connection")
	// 					return
	// 				}
	// 				defer conn.Close()
	// 				handleConn(conn)
	// 				return
	// 			}
	// 		}()
	// 	})
	//
	// 	AfterEach(func() {
	// 		client.Disconnect()
	// 		err := server.Close()
	// 		Expect(err).NotTo(HaveOccurred())
	// 	})
	//
	// 	Describe("Connect", func() {
	// 		BeforeEach(func() {
	// 			handleConn = func(net.Conn) {
	// 				return
	// 			}
	// 		})
	//
	// 		It("Connects without erroring", func() {
	// 			Expect(client.Connect(hostname, port, AuthInfo{})).NotTo(HaveOccurred())
	// 		})
	//
	// 		Context("When connecting fails", func() {
	// 			BeforeEach(func() {
	// 				hostname = "notavalidhost"
	// 			})
	//
	// 			It("Returns an error", func() {
	// 				Expect(client.Connect(hostname, port, AuthInfo{})).To(HaveOccurred())
	// 			})
	// 		})
	// 	})

	// XDescribe("Reconnect", func() {
	// 	var (
	// 		connCountChan chan bool
	// 	)
	//
	// 	BeforeEach(func() {
	// 		connCountChan = make(chan bool, 2)
	// 		// handleConn = func(net.Conn) {
	// 		// 	// connCountChan <- true
	// 		// 	return
	// 		// }
	// 	})
	//
	// 	It("Reconnects to the same endpoint", func() {
	// 		client.Connect(hostname, port, AuthInfo{})
	// 		client.Reconnect()
	// 		i := 0
	// 		for _ = range connCountChan {
	// 			i++
	// 		}
	// 		Expect(i).To(Equal(2))
	// 	})
	// })
	// })

	// XDescribe("Handshake", func() {
	// 	var (
	// 		clientConn, serverConn net.Conn
	// 	)
	//
	// 	BeforeEach(func() {
	// 		clientConn, serverConn = net.Pipe()
	//
	// 		client.Session = &Session{
	// 			Connection: clientConn,
	// 		}
	// 	})
	//
	// 	Context("When a HELO is received", func() {
	// 		var (
	// 			nonce []byte
	// 		)
	//
	// 		BeforeEach(func() {
	// 			nonce = make([]byte, 16)
	// 			rand.Read(nonce)
	//
	// 			h, err := protocol.PackedHelo(&protocol.HeloOpts{Nonce: nonce})
	// 			Expect(err).NotTo(HaveOccurred())
	//
	// 			_, err = serverConn.Write(h)
	// 			Expect(err).NotTo(HaveOccurred())
	// 		})
	//
	// 		It("Sends a PING back", func() {
	// 			// client.Handshake()
	// 			buf := make([]byte, 1024)
	// 			for {
	// 				serverConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	// 				_, err := serverConn.Read(buf)
	// 				if err == io.EOF {
	// 					break
	// 				}
	//
	// 				if err != nil {
	// 					Fail("Error reading from server connection")
	// 				}
	// 			}
	//
	// 			var p protocol.Ping
	// 			err := msgpack.Unmarshal(buf, &p)
	// 			Expect(err).NotTo(HaveOccurred())
	//
	// 			Expect(p.MessageType).To(Equal(protocol.MSGTYPE_PING))
	// 		})
	// 	})
	// })
})
