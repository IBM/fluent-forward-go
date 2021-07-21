package client_test

import (
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.ibm.com/Observability/fluent-forward-go/fluent/client"
)

var _ = Describe("Tcp", func() {
	Describe("TCPConnectionFactory", func() {
		var (
			server     net.Listener
			serverErr  error
			hostname   string
			port       int
			factory    *TCPConnectionFactory
			clientConn net.Conn
		)

		BeforeEach(func() {
			server, serverErr = net.Listen("tcp", ":0")
			Expect(serverErr).NotTo(HaveOccurred())

			port = server.Addr().(*net.TCPAddr).Port

			factory = &TCPConnectionFactory{
				Target: ServerAddress{
					Hostname: hostname,
					Port:     port,
				},
			}
		})

		AfterEach(func() {
			if clientConn != nil {
				closeErr := clientConn.Close()
				Expect(closeErr).NotTo(HaveOccurred())
			}

			err := server.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("New", func() {
			It("Creates a connection to the specified target", func() {
				go func() {
					defer GinkgoRecover()
					clientConn, err := factory.New()
					Expect(err).NotTo(HaveOccurred())
					Expect(clientConn).NotTo(BeNil())
				}()

				conn, err := server.Accept()
				Expect(conn).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
