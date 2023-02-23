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

package ws

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	ext "github.com/IBM/fluent-forward-go/fluent/client/ws/ext"
	"github.com/gorilla/websocket"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	DefaultCloseDeadline = 5 * time.Second
)

type Logger interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

type noopLogger struct{}

func (l *noopLogger) Println(v ...interface{}) {}

func (l *noopLogger) Printf(format string, v ...interface{}) {}

type ReadHandler func(conn Connection, messageType int, p []byte, err error) error

type ConnectionOptions struct {
	CloseDeadline time.Duration
	CloseHandler  func(conn Connection, code int, text string) error
	PingHandler   func(conn Connection, appData string) error
	PongHandler   func(conn Connection, appData string) error
	// TODO: should be a duration and added to `now` before every operation
	ReadDeadline time.Time
	// ReadHandler handles new messages received on the websocket. If an error
	// is received the client MUST call `Close`. An error returned by ReadHandler
	// will be retured by `Listen`.
	ReadHandler ReadHandler
	// TODO: should be a duration and added to `now` before every operation
	WriteDeadline time.Time
	// Logger is an optional debug log writer.
	Logger Logger
}

type ConnState uint8

const (
	ConnStateOpen ConnState = 1 << iota
	ConnStateListening
	ConnStateCloseReceived
	ConnStateCloseSent
	ConnStateClosed
	ConnStateError
)

//counterfeiter:generate . Connection
type Connection interface {
	ext.Conn
	CloseWithMsg(closeCode int, msg string) error
	Closed() bool
	ConnState() ConnState
	Listen() error
	ReadHandler() ReadHandler
	SetReadHandler(rh ReadHandler)
	Write(data []byte) (int, error)
}

type connection struct {
	ext.Conn
	logger        Logger
	closeLock     sync.Mutex
	listenLock    sync.Mutex
	writeLock     sync.Mutex
	stateLock     sync.RWMutex
	readHandler   ReadHandler
	done          chan struct{}
	connState     ConnState
	closeDeadline time.Duration
}

func NewConnection(conn ext.Conn, opts ConnectionOptions) (Connection, error) {
	wsc := &connection{
		Conn:      conn,
		done:      make(chan struct{}),
		connState: ConnStateOpen,
		logger:    opts.Logger,
	}

	if wsc.logger == nil {
		wsc.logger = &noopLogger{}
	}

	if opts.CloseHandler == nil {
		opts.CloseHandler = wsc.handleClose
	}

	wsc.SetCloseHandler(func(code int, text string) error {
		return opts.CloseHandler(wsc, code, text)
	})

	if opts.PingHandler != nil {
		wsc.SetPingHandler(func(appData string) error {
			return opts.PingHandler(wsc, appData)
		})
	}

	if opts.PongHandler != nil {
		wsc.SetPongHandler(func(appData string) error {
			return opts.PongHandler(wsc, appData)
		})
	}

	if opts.ReadHandler == nil {
		opts.ReadHandler = func(c Connection, _ int, _ []byte, err error) error {
			if err != nil {
				wsc.logger.Println("Default ReadHandler error:", err)

				_ = c.Close()
			}

			return err
		}
	}

	wsc.SetReadHandler(opts.ReadHandler)

	if opts.CloseDeadline == 0 {
		opts.CloseDeadline = DefaultCloseDeadline
	}

	wsc.closeDeadline = opts.CloseDeadline

	if err := wsc.SetReadDeadline(opts.ReadDeadline); err != nil {
		return nil, err
	}

	if err := wsc.SetWriteDeadline(opts.WriteDeadline); err != nil {
		return nil, err
	}

	return wsc, nil
}

func (wsc *connection) ConnState() ConnState {
	wsc.stateLock.RLock()
	defer wsc.stateLock.RUnlock()

	return wsc.connState
}

func (wsc *connection) hasConnState(cs ConnState) bool {
	wsc.stateLock.RLock()
	defer wsc.stateLock.RUnlock()

	return wsc.connState&cs != 0
}

func (wsc *connection) setConnState(cs ConnState) {
	wsc.stateLock.Lock()
	defer wsc.stateLock.Unlock()

	wsc.connState |= cs
}

func (wsc *connection) unsetConnState(cs ConnState) {
	wsc.stateLock.Lock()
	defer wsc.stateLock.Unlock()

	if wsc.connState&cs == 0 {
		return
	}

	wsc.connState ^= cs
}

// CloseWithMsg sends a close message to the peer
func (wsc *connection) CloseWithMsg(closeCode int, msg string) error {
	wsc.closeLock.Lock()

	if wsc.Closed() {
		wsc.closeLock.Unlock()
		return errors.New("multiple close calls")
	}

	wsc.unsetConnState(ConnStateOpen)
	wsc.closeLock.Unlock()

	var err error

	if !wsc.hasConnState(ConnStateError) {
		wsc.logger.Printf("sending close message: code %d; msg '%s'", closeCode, msg)

		// TODO: currently only the tests check this state to confirm handshake;
		// need to refactor it out
		wsc.setConnState(ConnStateCloseSent)

		err = wsc.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				closeCode, msg,
			),
		)

		// if the close message was sent and the connection is listening for incoming
		// messages, wait N seconds for a response.
		if err == nil && wsc.hasConnState(ConnStateListening) {
			wsc.logger.Println("awaiting peer response")

			select {
			case <-time.After(wsc.closeDeadline):
				// sent a close, but never heard back, close anyway
				err = errors.New("close deadline expired")
			case <-wsc.done:
			}
		}
	}

	wsc.setConnState(ConnStateClosed)

	wsc.logger.Println("closing the connection")

	if cerr := wsc.Conn.Close(); cerr != nil {
		err = cerr
	}

	return err
}

func (wsc *connection) Close() error {
	return wsc.CloseWithMsg(websocket.CloseNormalClosure, "closing connection")
}

func (wsc *connection) Closed() bool {
	return !wsc.hasConnState(ConnStateOpen)
}

func (wsc *connection) handleClose(_ Connection, code int, msg string) error {
	wsc.logger.Printf("close received: code %d; msg '%s'", code, msg)

	// TODO: currently only the tests check this state to confirm handshake;
	// need to refactor it out
	wsc.setConnState(ConnStateCloseReceived)

	return nil
}

type connMsg struct {
	mt      int
	message []byte
	err     error
}

func (wsc *connection) runReadLoop(nextMsg chan connMsg) {
	defer func() {
		wsc.logger.Println("exiting read loop")

		close(nextMsg)
		wsc.unsetConnState(ConnStateListening)
		close(wsc.done)
	}()

	msg := connMsg{}

	for {
		msg.mt, msg.message, msg.err = wsc.Conn.ReadMessage()

		if msg.err != nil {
			if wsc.hasConnState(ConnStateClosed) && errors.Is(msg.err, net.ErrClosed) {
				// healthy close
				break
			}

			var err net.Error
			if errors.As(msg.err, &err) || errors.Is(msg.err, net.ErrClosed) ||
				websocket.IsCloseError(msg.err, websocket.CloseAbnormalClosure) {
				// mark the connection with error state so Close doesn't attempt to
				// send closing message to peer
				wsc.setConnState(ConnStateError)
			}
		}

		nextMsg <- msg

		if msg.err != nil {
			break
		}
	}
}

func (wsc *connection) Listen() error {
	wsc.listenLock.Lock()

	if wsc.hasConnState(ConnStateListening) {
		wsc.listenLock.Unlock()
		return errors.New("already listening on this connection")
	}

	wsc.logger.Println("listening")
	wsc.setConnState(ConnStateListening)
	wsc.listenLock.Unlock()

	nextMsg := make(chan connMsg)
	go wsc.runReadLoop(nextMsg)

	var err error

	for msg := range nextMsg {
		if rerr := wsc.readHandler(wsc, msg.mt, msg.message, msg.err); rerr != nil {
			if msg.err != nil {
				wsc.logger.Println("handler returned error: ", msg.err.Error())
			}

			// set error only if it is something other than a normal close
			if !websocket.IsCloseError(rerr, websocket.CloseNormalClosure) {
				err = rerr
			}
		}
	}

	return err
}

func (wsc *connection) NextReader() (messageType int, r io.Reader, err error) {
	panic("use ReadHandler instead")
}

func (wsc *connection) ReadMessage() (messageType int, p []byte, err error) {
	panic("use ReadHandler instead")
}

func (wsc *connection) SetReadHandler(rh ReadHandler) {
	wsc.readHandler = rh
}

func (wsc *connection) ReadHandler() ReadHandler {
	return wsc.readHandler
}

func (wsc *connection) WriteMessage(messageType int, data []byte) error {
	wsc.writeLock.Lock()
	defer wsc.writeLock.Unlock()

	return wsc.Conn.WriteMessage(messageType, data)
}

func (wsc *connection) Write(data []byte) (int, error) {
	if err := wsc.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return 0, err
	}

	return len(data), nil
}
