package client_test

import (
	"crypto/tls"
	"net"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/IBM/fluent-forward-go/fluent/client"
)

var _ = Describe("SocketFactory", func() {

	var (
		network, address string
		server           net.Listener
		serverErr        error
		factory          *SocketFactory
		session          *Session
		tlsConfig        *tls.Config
	)

	BeforeEach(func() {
		network = "tcp"
		address = ":0"
	})

	JustBeforeEach(func() {
		if network == "unix" {
			if err := os.RemoveAll(address); err != nil {
				Fail(err.Error())
			}
		}

		if tlsConfig != nil {
			server, serverErr = tls.Listen(network, address, tlsConfig)
		} else {
			server, serverErr = net.Listen(network, address)
		}
		Expect(serverErr).NotTo(HaveOccurred())

		if network == "tcp" {
			address = server.Addr().(*net.TCPAddr).String()
		}

		factory = &SocketFactory{
			Network:   network,
			Address:   address,
			TLSConfig: tlsConfig,
		}
	})

	AfterEach(func() {
		if session != nil {
			closeErr := session.Connection.Close()
			Expect(closeErr).NotTo(HaveOccurred())
		}

		err := server.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Session", func() {
		testSession := func() {
			ch := make(chan struct{})

			go func() {
				defer GinkgoRecover()
				ch <- struct{}{}
				session, err := factory.Session()
				Expect(err).NotTo(HaveOccurred())
				Expect(session).NotTo(BeNil())
			}()

			<-ch
			conn, err := server.Accept()
			Expect(conn).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
		}

		When("using tcp", func() {
			It("returns an established connection", func() {
				testSession()
			})
		})

		When("using unix socket", func() {
			BeforeEach(func() {
				network = "unix"
				address = "/tmp/test.sock"
			})

			It("returns an established connection", func() {
				testSession()
			})
		})

		When("using tls", func() {
			BeforeEach(func() {
				cer, err := tls.LoadX509KeyPair("clientfakes/server.crt", "clientfakes/server.key")
				Expect(err).ToNot(HaveOccurred())

				tlsConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
			})

			It("returns an established connection", func() {
				testSession()
			})
		})
	})
})
