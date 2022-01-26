package client_test

import (
	"crypto/tls"
	"net"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/IBM/fluent-forward-go/fluent/client"
)

var _ = Describe("ConnFactory", func() {

	var (
		network, address string
		server           net.Listener
		serverErr        error
		factory          *ConnFactory
		tlsConfig        *tls.Config
	)

	BeforeEach(func() {
		network = "tcp"
		address = ":0"
		tlsConfig = nil
	})

	JustBeforeEach(func() {
		svrNetwork := network
		if svrNetwork == "" {
			svrNetwork = "tcp"
		}

		var clientTlsCfg *tls.Config
		if tlsConfig != nil {
			clientTlsCfg = &tls.Config{InsecureSkipVerify: true}
			server, serverErr = tls.Listen(svrNetwork, address, tlsConfig)
		} else {
			server, serverErr = net.Listen(svrNetwork, address)
		}
		Expect(serverErr).NotTo(HaveOccurred())

		if svrNetwork == "tcp" {
			address = server.Addr().(*net.TCPAddr).String()
		}

		factory = &ConnFactory{
			Network:   network,
			Address:   address,
			TLSConfig: clientTlsCfg,
			Timeout:   100 * time.Millisecond,
		}
	})

	AfterEach(func() {
		err := server.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("New", func() {
		testConnection := func(testTls bool) {
			socketConn, err := factory.New()
			Expect(err).NotTo(HaveOccurred())
			Expect(socketConn).NotTo(BeNil())
			time.Sleep(time.Millisecond)

			if testTls {
				tconn := socketConn.(*tls.Conn)
				state := tconn.ConnectionState()
				Expect(state.PeerCertificates).ToNot(BeEmpty())
			}

			tmp := make([]byte, 256)
			n, err := socketConn.Read(tmp)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(1))

			Expect(socketConn.Close()).ToNot(HaveOccurred())
		}

		JustBeforeEach(func() {
			svr := server
			go func() {
				defer GinkgoRecover()
				conn, err := svr.Accept()
				Expect(err).NotTo(HaveOccurred())
				Expect(conn).NotTo(BeNil())
				defer conn.Close()

				n, err := conn.Write([]byte{0x00})
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(1))
			}()

			time.Sleep(3 * time.Millisecond)
		})

		When("connecting with tcp", func() {
			It("returns an established connection", func() {
				testConnection(false)
			})

			When("Network is empty", func() {
				BeforeEach(func() {
					network = ""
				})

				It("defaults to tcp", func() {
					Expect(factory.Network).To(BeEmpty())
					testConnection(false)
					Expect(factory.Network).To(Equal("tcp"))
				})
			})

			When("using tls", func() {
				BeforeEach(func() {
					// go run $GOROOT/src/crypto/tls/generate_cert.go --rsa-bits 1024 \
					//   --host 127.0.0.1,::1,localhost --ca \
					//   --start-date "Jan 1 00:00:00 1970" --duration=1000000h
					cer, err := tls.LoadX509KeyPair("clientfakes/cert.pem", "clientfakes/key.pem")
					Expect(err).ToNot(HaveOccurred())

					tlsConfig = &tls.Config{
						Certificates: []tls.Certificate{cer},
					}
				})

				It("returns an established connection", func() {
					testConnection(true)
				})
			})
		})

		When("using unix socket", func() {
			BeforeEach(func() {
				network = "unix"
				address = "/tmp/test.sock"
			})

			AfterEach(func() {
				if err := os.RemoveAll(address); err != nil {
					Fail(err.Error())
				}
			})

			It("returns an established connection", func() {
				testConnection(false)
			})
		})
	})
})
