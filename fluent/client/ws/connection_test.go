package ws_test

import (
	"bytes"
	"errors"
	"fmt"
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

	// Describe("NewConnection", func() {
	// 	It("works", func() {
	// 		c, err := ws.NewConnection(fakeConn, ws.ConnectionOptions{})
	// 		Expect(err).ToNot(HaveOccurred())
	// 		Expect(c).ToNot(BeNil())
	// 		// TODO: test options are set correctly
	// 	})
	// })

	Describe("Connection", func() {
		var (
			connection, svrConnection   ws.Connection
			doClose                     bool
			svr                         *httptest.Server
			fakeConn                    *extfakes.FakeConn
			opts                        *ws.ConnectionOptions
			svrRcvdMsgs, clientRcvdMsgs chan message
			rhCallCt, rcvdMsgCt         *int32
			onListenFail                func(err error)
			listenErrCt                 int
		)

		var makeOpts = func(msgChan chan message, counter *int32) *ws.ConnectionOptions {
			return &ws.ConnectionOptions{
				CloseDeadline: 500 * time.Millisecond,
				ReadHandler: func(conn ws.Connection, msgType int, p []byte, err error) error {

					atomic.AddInt32(counter, 1)
					msg := message{
						mt:  msgType,
						msg: p,
						err: err,
					}

					msgChan <- msg

					if err != nil {
						conn.Close()
					}

					return err
				},
			}
		}

		BeforeEach(func() {
			rh := int32(0)
			rc := rh
			doClose = true
			rhCallCt = &rh
			rcvdMsgCt = &rc
			svrRcvdMsgs = make(chan message)

			svr = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()
				svrOpts := makeOpts(svrRcvdMsgs, rcvdMsgCt)

				var upgrader websocket.Upgrader
				wc, _ := upgrader.Upgrade(w, r, nil)
				var err error
				svrConnection, err = ws.NewConnection(wc, svrOpts)

				if err != nil {
					return
				}

				defer svrConnection.Close()
				Expect(svrConnection.Listen()).ToNot(HaveOccurred())
			}))

			clientRcvdMsgs = make(chan message, 1)
			fakeConn = &extfakes.FakeConn{}
			listenErrCt = 0
			opts = makeOpts(clientRcvdMsgs, rhCallCt)
			onListenFail = nil

		})

		JustBeforeEach(func() {
			u := "ws" + strings.TrimPrefix(svr.URL, "http")
			conn, _, err := websocket.DefaultDialer.Dial(u, nil)
			Expect(err).ToNot(HaveOccurred())

			connection, err = ws.NewConnection(conn, opts)
			Expect(err).ToNot(HaveOccurred())
			startSig := make(chan struct{}, 1)

			go func() {
				defer GinkgoRecover()
				startSig <- struct{}{}
				fmt.Println("here")
				if err := connection.Listen(); err != nil {
					fmt.Println("chere")
					if onListenFail == nil {
						Fail(err.Error())
					}
					onListenFail(err)
				}
				fmt.Println("there")
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
			close(svrRcvdMsgs)
			close(clientRcvdMsgs)
		})

		Describe("WriteMessage", func() {
			When("everything is copacetic", func() {
				It("writes messages to the connection", func() {
					err := connection.WriteMessage(1, []byte("oi"))
					Expect(err).ToNot(HaveOccurred())
					err = connection.WriteMessage(1, []byte("koi"))
					Expect(err).ToNot(HaveOccurred())

					m := <-svrRcvdMsgs
					Expect(m.msg).To(Equal([]byte("oi")))
					m = <-svrRcvdMsgs
					Expect(m.msg).To(Equal([]byte("koi")))

					Consistently(svrRcvdMsgs).ShouldNot(Receive())
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
				FIt("reads a message from the connection and calls the read handler", func() {
					Expect(len(svrRcvdMsgs)).To(Equal(0))

					err := connection.WriteMessage(1, []byte("oi"))
					Expect(err).ToNot(HaveOccurred())

					m := <-svrRcvdMsgs
					Expect(m.err).ToNot(HaveOccurred())

					Consistently(svrRcvdMsgs).ShouldNot(Receive())
				})
			})

			When("an error occurs", func() {
				// var listenErrCt int
				JustBeforeEach(func() {
					doClose = false
					onListenFail = func(e error) {
						listenErrCt++
						if e == nil {
							Fail("the BOOMing, where is it?")
						}
					}
				})

				It("enqueues the error", func() {
					// u := "ws" + strings.TrimPrefix(svr.URL, "http")
					// conn, _, err := websocket.DefaultDialer.Dial(u, nil)
					// Expect(err).ToNot(HaveOccurred())
					// connection, err = ws.NewConnection(conn, opts)
					// Expect(err).ToNot(HaveOccurred())
					//	Expect(connection.UnderlyingConn().Close()).ToNot(HaveOccurred())
					//svrConnection.WriteMessage(websocket.TextMessage, []byte("oi"))

					err := svrConnection.CloseWithMsg(websocket.ClosePolicyViolation, "meh")
					Expect(err).ToNot(HaveOccurred())
					// fmt.Println(<-svrRcvdMsgs)
					// fmt.Println(<-svrRcvdMsgs)
					//	Eventually(svrRcvdMsgs).Should(Receive())
					//Eventually(func() int { return int(rhCallCt) }).Should(BeNumerically("==", 1))
					Eventually(func() int { return int(*rhCallCt) }).Should(BeNumerically("==", 1))
				})

				When("the error is a normal close", func() {
					BeforeEach(func() {
						fakeConn.ReadMessageReturns(1, nil, &websocket.CloseError{Code: websocket.CloseNormalClosure})
					})

					It("does not enqueue the error", func() {
						Eventually(func() int { return int(*rhCallCt) }).Should(BeNumerically(">=", 1))
						Consistently(func() int { return listenErrCt }).Should(BeNumerically("==", 0))
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
