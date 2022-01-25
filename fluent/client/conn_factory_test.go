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
		if tlsConfig != nil {
			server, serverErr = tls.Listen(network, address, tlsConfig)
		} else {
			server, serverErr = net.Listen(network, address)
		}
		Expect(serverErr).NotTo(HaveOccurred())

		if network == "tcp" {
			address = server.Addr().(*net.TCPAddr).String()
		}

		factory = &ConnFactory{
			Network:   network,
			Address:   address,
			TLSConfig: tlsConfig,
			Timeout:   100 * time.Millisecond,
		}
	})

	AfterEach(func() {
		err := server.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("New", func() {
		testConnection := func() {
			socketConn, err := factory.New()
			Expect(err).NotTo(HaveOccurred())
			Expect(socketConn).NotTo(BeNil())
			time.Sleep(time.Millisecond)
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
				testConnection()
			})

			When("using tls", func() {
				BeforeEach(func() {
					cer, err := tls.LoadX509KeyPair("clientfakes/test.crt", "clientfakes/test.key")
					Expect(err).ToNot(HaveOccurred())

					tlsConfig = &tls.Config{Certificates: []tls.Certificate{cer}, InsecureSkipVerify: true}
				})

				It("returns an established connection", func() {
					testConnection()
				})
			})
		})

		When("using unix socket", func() {
			// Note: when this test runs independently, it passes;
			// when run with the full suite, it fails
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
				testConnection()
			})
		})
	})
})
