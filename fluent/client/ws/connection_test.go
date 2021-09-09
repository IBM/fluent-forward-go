package ws_test

import (
	"bytes"
	"errors"

	"github.com/IBM/fluent-forward-go/fluent/client/ws"
	"github.com/IBM/fluent-forward-go/fluent/client/ws/ext/extfakes"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type message struct {
	mt  int
	msg []byte
	err error
}

var _ = Describe("Connection", func() {
	var (
		fakeConn           *extfakes.FakeConn
		opts               ws.ConnectionOptions
		readMsgs           chan message
		rhCallCt, rmCallCt int
		onFail             func(err error)
	)

	BeforeEach(func() {
		rmCallCt = 0
		rhCallCt = 0
		readMsgs = make(chan message, 1)
		fakeConn = &extfakes.FakeConn{}
		fakeConn.ReadMessageStub = func() (int, []byte, error) {
			rmCallCt++
			m, ok := <-readMsgs
			if !ok {
				return m.mt, m.msg, errors.New("connection")
			}
			return m.mt, m.msg, m.err
		}

		opts = ws.ConnectionOptions{
			ReadHandler: func(_ ws.Connection, msgType int, p []byte, err error) error {
				rhCallCt++
				return err
			},
		}

		onFail = func(e error) { Fail(e.Error()) }
	})

	AfterEach(func() {
		close(readMsgs)
	})

	Describe("NewConnection", func() {
		It("works", func() {
			Expect(ws.NewConnection(fakeConn, ws.ConnectionOptions{})).ToNot(BeNil())
			// TODO: test options are set correctly
		})
	})

	Describe("Connection", func() {
		var (
			connection ws.Connection
			doClose    bool
		)

		BeforeEach(func() {
			doClose = true
			connection = ws.NewConnection(fakeConn, opts)

			go func() {
				defer GinkgoRecover()
				if err := connection.Listen(); err != nil {
					onFail(err)
				}
			}()
		})

		AfterEach(func() {
			if doClose {
				Expect(connection.Close()).ToNot(HaveOccurred())
			}
		})

		Describe("WriteMessage", func() {
			When("everything is copacetic", func() {
				It("writes messages to the connection", func() {
					err := connection.WriteMessage(1, []byte("oi"))
					Expect(err).ToNot(HaveOccurred())
					err = connection.WriteMessage(1, []byte("koi"))
					Expect(err).ToNot(HaveOccurred())

					Eventually(fakeConn.WriteMessageCallCount).Should(Equal(2))
					Consistently(fakeConn.WriteMessageCallCount).Should(BeNumerically("<", 3))

					mt, bmsg := fakeConn.WriteMessageArgsForCall(0)
					Expect(mt).To(Equal(1))
					Expect(bmsg).To(Equal([]byte("oi")))

					mt, bmsg = fakeConn.WriteMessageArgsForCall(1)
					Expect(mt).To(Equal(1))
					Expect(bmsg).To(Equal([]byte("koi")))
				})
			})

			When("an error occurs", func() {
				BeforeEach(func() {
					fakeConn.WriteMessageReturns(errors.New("BOOM"))
				})

				It("returns an error", func() {
					Expect(connection.WriteMessage(1, []byte("oi"))).To(HaveOccurred())
					Expect(fakeConn.WriteMessageCallCount()).To(Equal(1))
				})
			})
		})

		Describe("ReadMessage", func() {
			When("everything is copacetic", func() {
				It("reads a message from the connection and calls the read handler", func() {
					readMsgs <- message{1, []byte("oi"), nil}
					Eventually(func() int { return rhCallCt }).Should(BeNumerically(">", 0))
					Eventually(func() int { return rmCallCt }).Should(BeNumerically(">", 0))
				})
			})

			When("an error occurs", func() {
				BeforeEach(func() {
					fakeConn.ReadMessageReturns(1, nil, errors.New("BOOM"))
					onFail = func(e error) {
						if e == nil {
							Fail("the BOOMing, where is it?")
						}
					}
				})

				It("enqueues the error", func() {
					Consistently(func() int { return rhCallCt }).Should(BeNumerically("==", 0))
				})
			})
		})

		Describe("CloseWithMsg", func() {
			When("everything is copacetic", func() {
				It("sends a signal", func() {
					Expect(connection.CloseWithMsg(1, "a")).ToNot(HaveOccurred())
					a, b := fakeConn.WriteMessageArgsForCall(0)
					Expect(a).To(Equal(8))
					msg := websocket.FormatCloseMessage(1, "a")
					Expect(bytes.Equal(b, msg)).To(BeTrue())
				})
			})
		})

		Describe("Close", func() {
			BeforeEach(func() {
				doClose = false
			})

			When("everything is copacetic", func() {
				It("signals close", func() {
					Expect(connection.Close()).ToNot(HaveOccurred())
					a, b := fakeConn.WriteMessageArgsForCall(0)
					Expect(a).To(Equal(8))
					msg := websocket.FormatCloseMessage(
						websocket.CloseNormalClosure, "so long and thanks for all the fish",
					)
					Expect(bytes.Equal(b, msg)).To(BeTrue())
				})
			})

			When("called multiple times", func() {
				It("doesn't error", func() {
					Expect(connection.Close()).ToNot(HaveOccurred())
					Expect(connection.Close()).ToNot(HaveOccurred())
				})
			})

			When("the connection errors on close", func() {
				BeforeEach(func() {
					fakeConn.WriteMessageReturns(errors.New("BOOM"))
					fakeConn.CloseReturns(errors.New("BOOM"))
				})

				It("returns an error", func() {
					Expect(connection.Close()).To(HaveOccurred())
				})
			})
		})
	})
})
