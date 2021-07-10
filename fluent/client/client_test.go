package client_test

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmihailenco/msgpack/v5"

	. "github.ibm.com/Observability/fluent-forward-go/fluent/client"
	"github.ibm.com/Observability/fluent-forward-go/fluent/protocol"
)

var _ = Describe("Client", func() {
	var (
		client *Client
	)

	BeforeEach(func() {
		client = &Client{Timeout: 2 * time.Second}
	})

	XDescribe("Network Management", func() {
		var (
			server     net.Listener
			serverErr  error
			hostname   string
			port       int
			handleConn func(net.Conn)
		)

		BeforeEach(func() {
			server, serverErr = net.Listen("tcp", ":0")
			Expect(serverErr).NotTo(HaveOccurred())

			port = server.Addr().(*net.TCPAddr).Port

			handleConn = func(net.Conn) {
				fmt.Fprintln(os.Stderr, "HANDLING CONNECTION")
				return
			}
		})

		JustBeforeEach(func() {
			go func() {
				defer GinkgoRecover()
				for {
					conn, err := server.Accept()
					if err != nil {
						Fail("Failure accepting connection")
						return
					}
					defer conn.Close()
					handleConn(conn)
					return
				}
			}()
		})

		AfterEach(func() {
			client.Disconnect()
			err := server.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("Connect", func() {
			BeforeEach(func() {
				handleConn = func(net.Conn) {
					return
				}
			})

			It("Connects without erroring", func() {
				Expect(client.Connect(hostname, port, AuthInfo{})).NotTo(HaveOccurred())
			})

			Context("When connecting fails", func() {
				BeforeEach(func() {
					hostname = "notavalidhost"
				})

				It("Returns an error", func() {
					Expect(client.Connect(hostname, port, AuthInfo{})).To(HaveOccurred())
				})
			})
		})

		XDescribe("Reconnect", func() {
			var (
				connCountChan chan bool
			)

			BeforeEach(func() {
				connCountChan = make(chan bool, 2)
				// handleConn = func(net.Conn) {
				// 	// connCountChan <- true
				// 	return
				// }
			})

			It("Reconnects to the same endpoint", func() {
				client.Connect(hostname, port, AuthInfo{})
				client.Reconnect()
				i := 0
				for _ = range connCountChan {
					i++
				}
				Expect(i).To(Equal(2))
			})
		})
	})

	Describe("Handshake", func() {
		var (
			clientConn, serverConn net.Conn
		)

		BeforeEach(func() {
			clientConn, serverConn = net.Pipe()

			client.Session = &Session{
				Connection: clientConn,
			}
		})

		Context("When a HELO is received", func() {
			var (
				nonce []byte
			)

			BeforeEach(func() {
				nonce = make([]byte, 16)
				rand.Read(nonce)

				h, err := protocol.PackedHelo(&protocol.HeloOpts{Nonce: nonce})
				Expect(err).NotTo(HaveOccurred())

				_, err = serverConn.Write(h)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Sends a PING back", func() {
				client.Handshake()
				buf := make([]byte, 1024)
				for {
					serverConn.SetReadDeadline(time.Now().Add(2 * time.Second))
					_, err := serverConn.Read(buf)
					if err == io.EOF {
						break
					}

					if err != nil {
						Fail("Error reading from server connection")
					}
				}

				var p protocol.Ping
				err := msgpack.Unmarshal(buf, &p)
				Expect(err).NotTo(HaveOccurred())

				Expect(p.MessageType).To(Equal(protocol.MSGTYPE_PING))
			})
		})
	})
})
