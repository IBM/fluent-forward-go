package ws

import (
	"errors"
	"io"
	"log"
	"net"
	"runtime/debug"
	"sync"
	"time"

	ext "github.com/IBM/fluent-forward-go/fluent/client/ws/ext"
	"github.com/gorilla/websocket"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	DefaultCloseDeadline = 5 * time.Second
)

type ReadHandler func(conn Connection, messageType int, p []byte, err error) error

type ConnectionOptions struct {
	CloseDeadline time.Duration
	CloseHandler  func(conn Connection, code int, text string) error
	PingHandler   func(conn Connection, appData string) error
	PongHandler   func(conn Connection, appData string) error
	// TODO: should be a duration and added to `now` before every operation
	ReadDeadline time.Time
	// ReadHandler handles new messages received on the websocket. If the handler receives
	// an error, the client MUST close the connection and return an error.
	ReadHandler ReadHandler
	// TODO: should be a duration and added to `now` before every operation
	WriteDeadline time.Time
}

type ConnState uint8

const (
	ConnStateOpen ConnState = 1 << iota
	ConnStateCloseReceived
	ConnStateCloseSent
	ConnStateClosed
)

//counterfeiter:generate . Connection
type Connection interface {
	ext.Conn
	CloseWithMsg(closeCode int, msg string) error
	Closed() bool
	Listen() error
	ReadHandler() ReadHandler
	SetReadHandler(rh ReadHandler)
	Write(data []byte) (int, error)
}

type connection struct {
	ext.Conn
	writeLock     sync.Mutex
	stateLock     sync.RWMutex
	readHandler   ReadHandler
	closedSig     chan struct{}
	connState     ConnState
	closeDeadline time.Duration
}

func NewConnection(conn ext.Conn, opts ConnectionOptions) (Connection, error) {
	wsc := &connection{
		Conn:      conn,
		closedSig: make(chan struct{}),
		connState: ConnStateOpen,
	}

	if opts.CloseHandler == nil {
		opts.CloseHandler = wsc.HandleClose
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

func (wsc *connection) hasConnState(cs ConnState) bool {
	wsc.stateLock.RLock()
	defer wsc.stateLock.RUnlock()

	return wsc.connState&cs != 0
}

func (wsc *connection) hasAnyConnState(cs ...ConnState) bool {
	wsc.stateLock.RLock()
	defer wsc.stateLock.RUnlock()

	for i := 0; i < len(cs); i++ {
		if wsc.connState&cs[i] != 0 {
			return true
		}
	}

	return false
}

func (wsc *connection) setConnState(cs ConnState) {
	wsc.stateLock.Lock()
	defer wsc.stateLock.Unlock()

	wsc.connState |= cs
}

func (wsc *connection) unsetConnState(cs ConnState) {
	if !wsc.hasConnState(cs) {
		return
	}

	wsc.stateLock.Lock()
	defer wsc.stateLock.Unlock()

	wsc.connState ^= cs
}

func (wsc *connection) HandleClose(_ Connection, code int, text string) error {
	if wsc.hasConnState(ConnStateClosed) {
		// already closed, nothing else to do
		return nil
	}

	log.Println("received close message")
	wsc.setConnState(ConnStateCloseReceived)

	if !wsc.hasConnState(ConnStateCloseSent) {
		// respond with close
		return wsc.Close()
	}

	// sent a close and received one, signal shutdown
	log.Println("closing handshake complete")
	wsc.setConnState(ConnStateClosed)

	return nil
}

// CloseConn return true
func (wsc *connection) CloseWithMsg(closeCode int, msg string) error {
	if wsc.hasConnState(ConnStateCloseSent) {
		return errors.New("multiple close calls")
	}

	wsc.unsetConnState(ConnStateOpen)

	log.Println("sending close message")

	err := wsc.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(
			closeCode, msg,
		),
	)

	wsc.setConnState(ConnStateCloseSent)

	if err != nil && err != websocket.ErrCloseSent {
		log.Println("write close failed", err)
	}

	select {
	case <-time.After(wsc.closeDeadline):
		log.Println("close deadline expired")
		// sent a close, but never heard back, close anyway
		wsc.setConnState(ConnStateClosed)
	case <-wsc.closedSig:
	}

	// spec says that only server should close the network connection,
	// but consensus is that it doesn't matter
	return wsc.Conn.Close()
}

func (wsc *connection) Close() error {
	return wsc.CloseWithMsg(websocket.CloseNormalClosure, "so long and thanks for all the fish")
}

func (wsc *connection) Closed() bool {
	return !wsc.hasConnState(ConnStateOpen)
}

type connMsg struct {
	mt      int
	message []byte
	err     error
}

func (wsc *connection) Listen() error {
	// TODO prevent call if already listening
	nextMsg := make(chan connMsg)

	defer func() {
		log.Println("signaling closed")

		wsc.closedSig <- struct{}{}

		if r := recover(); r != nil {
			debug.PrintStack()
			log.Println("panic:", r)
		}
	}()

	go func() {
		msg := connMsg{}

		for {
			msg.mt, msg.message, msg.err = wsc.Conn.ReadMessage()
			nextMsg <- msg

			if wsc.hasAnyConnState(ConnStateCloseReceived, ConnStateClosed) ||
				msg.err == net.ErrClosed {
				close(nextMsg)
				return
			}
		}
	}()

	var closeErr error

	for msg := range nextMsg {
		if err := wsc.readHandler(wsc, msg.mt, msg.message, msg.err); err != nil {
			// enqueue error only if it is something other than a normal close
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				log.Println("readhandler error:", err)
				closeErr = err
			}
		}
	}

	return closeErr
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
