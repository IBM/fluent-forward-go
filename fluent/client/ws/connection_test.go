package ws_test

import (
	"bytes"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"time"

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
		fakeConn                    *extfakes.FakeConn
		opts                        ws.ConnectionOptions
		svrReadMsgs, clientReadMsgs chan message
		rhCallCt, rcvdMsgCt         int32
		onFail                      func(err error)
		echo                        func(w http.ResponseWriter, r *http.Request)
	)

	var upgrader = websocket.Upgrader{}

	BeforeEach(func() {
		rhCallCt = 0
		rcvdMsgCt = 0
		svrReadMsgs = make(chan message)
		clientReadMsgs = make(chan message)
		fakeConn = &extfakes.FakeConn{}

		echo = func(w http.ResponseWriter, r *http.Request) {
			xopts := ws.ConnectionOptions{
				CloseDeadline: 100 * time.Millisecond,
				ReadHandler: func(conn ws.Connection, msgType int, p []byte, err error) error {
					atomic.AddInt32(&rcvdMsgCt, 1)
					msg := message{
						mt:  msgType,
						msg: p,
						err: err,
					}

					svrReadMsgs <- msg

					if err != nil {
						conn.Close()
					}

					return err
				},
			}

			wc, _ := upgrader.Upgrade(w, r, nil)
			c, err := ws.NewConnection(wc, xopts)

			if err != nil {
				return
			}

			defer c.Close()
			c.Listen()
		}

		opts = ws.ConnectionOptions{
			CloseDeadline: 100 * time.Millisecond,
			ReadHandler: func(conn ws.Connection, msgType int, p []byte, err error) error {
				atomic.AddInt32(&rhCallCt, 1)
				msg := message{
					mt:  msgType,
					msg: p,
					err: err,
				}
				clientReadMsgs <- msg
				if err != nil {
					conn.Close()
				}
				return err
			},
		}

		onFail = nil

	})

	AfterEach(func() {
		close(svrReadMsgs)
		close(clientReadMsgs)
	})

	Describe("NewConnection", func() {
		It("works", func() {
			c, err := ws.NewConnection(fakeConn, ws.ConnectionOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(c).ToNot(BeNil())
			// TODO: test options are set correctly
		})
	})

	Describe("Connection", func() {
		var (
			connection ws.Connection
			doClose    bool
			svr        *httptest.Server
		)

		BeforeEach(func() {
			doClose = true

			svr = httptest.NewServer(http.HandlerFunc(echo))

			u := "ws" + strings.TrimPrefix(svr.URL, "http")

			// Connect to the server
			conn, _, err := websocket.DefaultDialer.Dial(u, nil)
			Expect(err).ToNot(HaveOccurred())

			connection, err = ws.NewConnection(conn, opts)
			Expect(err).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			startSig := make(chan struct{}, 1)

			go func() {
				defer GinkgoRecover()
				startSig <- struct{}{}
				if err := connection.Listen(); err != nil {
					if onFail == nil {
						defer GinkgoRecover()
						Fail(err.Error())
					}
					onFail(err)
				}
			}()

			// wait for Listen loop to start
			<-startSig
			time.Sleep(time.Millisecond)
		})

		AfterEach(func() {
			if doClose {
				Expect(connection.Close()).ToNot(HaveOccurred())
				Eventually(connection.Closed).Should(BeTrue())
			}
			svr.Close()
		})

		Describe("WriteMessage", func() {
			When("everything is copacetic", func() {
				FIt("writes messages to the connection", func() {
					err := connection.WriteMessage(1, []byte("oi"))
					Expect(err).ToNot(HaveOccurred())
					err = connection.WriteMessage(1, []byte("koi"))
					Expect(err).ToNot(HaveOccurred())

					x := <-svrReadMsgs
					Expect(x.msg).To(Equal([]byte("oi")))
					x = <-svrReadMsgs
					Expect(x.msg).To(Equal([]byte("koi")))

					Consistently(svrReadMsgs).ShouldNot(Receive())
				})
			})

			When("an error occurs", func() {
				BeforeEach(func() {
					doClose = false
					connection.Close()
				})

				It("returns an error", func() {
					Expect(connection.WriteMessage(1, []byte("oi"))).To(MatchError(net.ErrClosed))
				})
			})
		})

		Describe("ReadMessage", func() {
			When("everything is copacetic", func() {
				It("reads a message from the connection and calls the read handler", func() {
					//svrReadMsgs <- message{1, []byte("oi"), nil}
					var gt int32
					Expect(rcvdMsgCt).To(Equal(gt))
					err := connection.WriteMessage(1, []byte("oi"))
					Expect(err).ToNot(HaveOccurred())
					Eventually(func() int32 { return rcvdMsgCt }).Should(BeNumerically(">", gt))
					//Eventually(func() int32 { return rmCallCt }).Should(BeNumerically(">", gt))
				})
			})

			When("an error occurs", func() {
				var callCt int
				BeforeEach(func() {
					callCt = 0
					fakeConn.ReadMessageReturns(1, nil, &websocket.CloseError{})
					onFail = func(e error) {
						defer GinkgoRecover()
						callCt++
						if e == nil {
							Fail("the BOOMing, where is it?")
						}
					}
				})

				It("enqueues the error", func() {
					Eventually(func() int { return int(rhCallCt) }).Should(BeNumerically(">=", 1))
					Eventually(func() int { return callCt }).Should(BeNumerically("==", 1))
				})

				When("the error is a normal close", func() {
					BeforeEach(func() {
						fakeConn.ReadMessageReturns(1, nil, &websocket.CloseError{Code: websocket.CloseNormalClosure})
					})

					It("does not enqueue the error", func() {
						Eventually(func() int { return int(rhCallCt) }).Should(BeNumerically(">=", 1))
						Consistently(func() int { return callCt }).Should(BeNumerically("==", 0))
					})
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
				It("errors", func() {
					Expect(connection.Close()).ToNot(HaveOccurred())
					Expect(connection.Close()).To(MatchError("multiple close calls"))
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
