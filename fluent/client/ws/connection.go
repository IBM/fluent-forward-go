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

type ReadHandler func(conn Connection, messageType int, p []byte, err error) error

type ConnectionOptions struct {
	CloseDeadline time.Duration
	CloseHandler  func(conn Connection, code int, text string) error
	PingHandler   func(conn Connection, appData string) error
	PongHandler   func(conn Connection, appData string) error
	// TODO: should be a duration and added to `now` before every operation
	ReadDeadline time.Time
	// ReadHandler handles new messages received on the websocket. The read loop
	// will exit automatically if a close message or network error is received. In
	// all other cases, the client MUST call `Close` to exit the read loop.
	ReadHandler ReadHandler
	// TODO: should be a duration and added to `now` before every operation
	WriteDeadline time.Time
}

type ConnState uint8

const (
	ConnStateOpen ConnState = 1 << iota
	ConnStateListening
	ConnStateCloseReceived
	ConnStateCloseSent
	ConnStateClosed
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

// CloseWithMsg sends a close message to the peer
func (wsc *connection) CloseWithMsg(closeCode int, msg string) error {
	if wsc.hasConnState(ConnStateCloseSent) {
		return errors.New("multiple close calls")
	}

	wsc.unsetConnState(ConnStateOpen)

	err := wsc.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(
			closeCode, msg,
		),
	)

	wsc.setConnState(ConnStateCloseSent)

	if err == nil && wsc.hasConnState(ConnStateListening) && !wsc.hasConnState(ConnStateCloseReceived) {
		select {
		case <-time.After(wsc.closeDeadline):
			// sent a close, but never heard back, close anyway
			err = errors.New("close deadline expired")
		case <-wsc.closedSig:
		}
	}

	wsc.setConnState(ConnStateClosed)

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

func (wsc *connection) handleClose(_ Connection, code int, text string) error {
	wsc.setConnState(ConnStateCloseReceived)

	var err error

	// If the peer initiated the close, then this client must send a message
	// confirming receipt. If this client sent the initial close message, then
	// the closing handshake is complete and no further action is required.
	if !wsc.hasConnState(ConnStateCloseSent) {
		// respond with close
		err = wsc.Close()
	}

	return err
}

type connMsg struct {
	mt      int
	message []byte
	err     error
}

func (wsc *connection) Listen() error {
	if wsc.hasConnState(ConnStateListening) {
		return errors.New("already listening on this connection")
	}

	wsc.setConnState(ConnStateListening)

	nextMsg := make(chan connMsg)

	go func() {
		defer func() {
			wsc.unsetConnState(ConnStateListening)
			wsc.closedSig <- struct{}{}
		}()

		msg := connMsg{}

		for {
			msg.mt, msg.message, msg.err = wsc.Conn.ReadMessage()

			if msg.err == net.ErrClosed {
				break
			}

			nextMsg <- msg

			if _, ok := msg.err.(net.Error); ok {
				break
			}

			if wsc.hasAnyConnState(ConnStateCloseReceived, ConnStateClosed) {
				break
			}
		}

		close(nextMsg)
	}()

	var err error

	for msg := range nextMsg {
		if wsc.readHandler == nil {
			continue
		}

		// TODO error handling in this loop still needs work; passing a close message
		// to the read handler may not be necessary
		if rerr := wsc.readHandler(wsc, msg.mt, msg.message, msg.err); rerr != nil {
			// set error only if it is something other than a normal close
			if !websocket.IsCloseError(rerr, websocket.CloseNormalClosure) {
				err = rerr
			}
		}
	}

	// TODO return a default error, eg `ErrConnectionClosed`, in the same way
	// http.Server.Listen and websocket.Close do
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
