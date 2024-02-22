/*
MIT License

Copyright contributors to the fluent-forward-go project

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package ws_test

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/aanujj/fluent-forward-go/fluent/client/ws"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

type message struct {
	mt  int
	msg []byte
	err error
}

var _ = Describe("Connection", func() {
	var (
		checkClose, checkSvrClose       bool
		connection, svrConnection       ws.Connection
		svr                             *httptest.Server
		opts                            ws.ConnectionOptions
		svrRcvdMsgs                     chan message
		listenErrs                      chan error
		exitConnState, svrExitConnState ws.ConnState
		logBuffer                       *gbytes.Buffer
	)

	var makeOpts = func(logBuffer *gbytes.Buffer, msgChan chan message, name string) ws.ConnectionOptions {
		logFlags := log.Lshortfile | log.LUTC | log.Lmicroseconds
		logger := log.New(logBuffer, name+"> ", logFlags)

		return ws.ConnectionOptions{
			CloseDeadline: 500 * time.Millisecond,
			ReadHandler: func(conn ws.Connection, msgType int, p []byte, err error) error {
				if msgChan != nil {
					msgChan <- message{
						mt:  msgType,
						msg: p,
						err: err,
					}
				}

				if err != nil {
					logger.Println("ReadHandler received error:", err)
					_ = conn.Close()
				}

				return err
			},
			Logger: logger,
		}
	}

	newHandler := func(logBuffer *gbytes.Buffer, svrRcvdMsgs chan message) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			svrOpts := makeOpts(logBuffer, svrRcvdMsgs, "server")

			var upgrader websocket.Upgrader
			wc, _ := upgrader.Upgrade(w, r, nil)

			var err error
			svrConnection, err = ws.NewConnection(wc, svrOpts)
			if err != nil {
				return
			}

			svrConnection.Listen()
			log.Println("exit server handler")
		})
	}

	BeforeEach(func() {
		logBuffer = gbytes.NewBuffer()

		exitConnState = ws.ConnStateCloseReceived | ws.ConnStateCloseSent | ws.ConnStateClosed
		svrExitConnState = ws.ConnStateCloseReceived | ws.ConnStateCloseSent | ws.ConnStateClosed

		checkSvrClose = true
		svrRcvdMsgs = make(chan message, 1)
		svr = httptest.NewServer(newHandler(logBuffer, svrRcvdMsgs))

		checkClose = true
		opts = makeOpts(logBuffer, nil, "client")
	})

	JustBeforeEach(func() {
		u := "ws" + strings.TrimPrefix(svr.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(u, nil)
		Expect(err).ToNot(HaveOccurred())

		connection, err = ws.NewConnection(conn, opts)
		Expect(err).ToNot(HaveOccurred())

		listenErrs = make(chan error, 1)

		go func() {
			defer GinkgoRecover()

			Expect(connection.ConnState()).To(Equal(ws.ConnStateOpen))

			if err := connection.Listen(); err != nil {
				listenErrs <- err
			}
		}()

		// wait for Listen loop to start
		time.Sleep(10 * time.Millisecond)
		Expect(connection.Closed()).To(BeFalse())
	})

	AfterEach(func() {
		defer func() {
			fmt.Print(string(logBuffer.Contents()))
			logBuffer.Close()
		}()

		if !connection.Closed() {
			err := connection.Close()
			if checkClose {
				Expect(err).ToNot(HaveOccurred())
			}
			Eventually(connection.Closed).Should(BeTrue())
		}

		if !svrConnection.Closed() {
			err := svrConnection.Close()
			if checkSvrClose {
				Expect(err).ToNot(HaveOccurred())
			}
			Eventually(svrConnection.Closed).Should(BeTrue())
		}

		svr.Close()

		if checkClose {
			Eventually(connection.ConnState).Should(Equal(exitConnState))
		}
		if checkSvrClose {
			Eventually(svrConnection.ConnState).Should(Equal(svrExitConnState))
		}
	})

	Describe("NewConnection", func() {
		BeforeEach(func() {
			opts.ReadHandler = nil
		})

		When("no ReadHandler is set", func() {
			It("sets a default handler that handles the closing handshake", func() {
				Expect(connection.Close()).ToNot(HaveOccurred())
				closeMsg := <-svrRcvdMsgs
				Expect(closeMsg.err.Error()).To(MatchRegexp("closing connection"))
				Eventually(logBuffer.Contents).Should(MatchRegexp("Default ReadHandler error.+closing connection"))
			})
		})
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
			It("returns an error", func() {
				Expect(connection.Close()).ToNot(HaveOccurred())
				Expect(connection.WriteMessage(1, nil).Error()).To(MatchRegexp("close sent"))
			})
		})
	})

	Describe("Listen", func() {
		When("everything is copacetic", func() {
			It("reads a message from the connection and calls the read handler", func() {
				Expect(len(svrRcvdMsgs)).To(Equal(0))

				err := connection.WriteMessage(1, []byte("oi"))
				Expect(err).ToNot(HaveOccurred())

				m := <-svrRcvdMsgs
				Expect(m.err).ToNot(HaveOccurred())
				Expect(bytes.Equal(m.msg, []byte("oi"))).To(BeTrue())
			})
		})

		When("already listening", func() {
			It("errors", func() {
				Expect(connection.Listen().Error()).To(MatchRegexp("already listening on this connection"))
				Expect(connection.Listen().Error()).To(MatchRegexp("already listening on this connection"))
			})
		})

		When("a network error occurs", func() {
			JustBeforeEach(func() {
				checkClose = false
				checkSvrClose = false
				exitConnState = ws.ConnStateClosed | ws.ConnStateError
				svrExitConnState = exitConnState
				connection.UnderlyingConn().Close()
			})

			It("returns a network error", func() {
				err := <-listenErrs
				Expect(err.Error()).To(MatchRegexp("use of closed network connection"))
			})
		})

		When("a close error occurs", func() {
			It("returns abnormal closures", func() {
				err := svrConnection.CloseWithMsg(websocket.ClosePolicyViolation, "meh")
				Expect(err).ToNot(HaveOccurred())
				err = <-listenErrs
				Expect(err.Error()).To(MatchRegexp("meh"))
			})

			It("does not return normal closures", func() {
				Expect(svrConnection.Close()).ToNot(HaveOccurred())
				Consistently(listenErrs).ShouldNot(Receive())
			})
		})
	})

	Describe("CloseWithMsg", func() {
		When("everything is copacetic", func() {
			It("sends a signal", func() {
				Expect(connection.CloseWithMsg(1000, "oi")).ToNot(HaveOccurred())
				Expect(connection.Closed()).To(BeTrue())

				closeMsg := <-svrRcvdMsgs
				Expect(closeMsg.err.Error()).To(MatchRegexp("oi"))
			})
		})
	})

	Describe("Close and Closed", func() {
		JustBeforeEach(func() {
			Expect(connection.Closed()).To(BeFalse())
		})

		AfterEach(func() {
			Expect(connection.Closed()).To(BeTrue())
		})

		When("everything is copacetic", func() {
			It("signals close", func() {
				Expect(connection.Close()).ToNot(HaveOccurred())
				closeMsg := <-svrRcvdMsgs
				Expect(closeMsg.err.Error()).To(MatchRegexp("closing connection"))
			})
		})

		When("called multiple times", func() {
			It("errors", func() {
				Expect(connection.Close()).ToNot(HaveOccurred())
				Expect(connection.Close().Error()).To(MatchRegexp("multiple close calls"))
			})
		})

		When("the connection errors on close", func() {
			BeforeEach(func() {
				opts.ReadHandler = func(conn ws.Connection, msgType int, p []byte, err error) error {
					// This is kinda cheating, but there is a race condition where the default test
					// read handler occasionally calls Close before the test does. In that case,
					// `Close` returns the "multiple close calls" error.
					return nil
				}
			})

			JustBeforeEach(func() {
				checkClose = false
				checkSvrClose = false
				exitConnState = ws.ConnStateClosed | ws.ConnStateError
				svrExitConnState = exitConnState
				connection.UnderlyingConn().Close()
			})

			AfterEach(func() {
				Expect(connection.ConnState() & exitConnState).To(BeNumerically(">", 1))
				Expect(connection.ConnState() & svrExitConnState).To(BeNumerically(">", 1))
			})

			It("returns an error", func() {
				Expect(connection.Close().Error()).To(MatchRegexp("use of closed network connection"))
			})
		})
	})
})
