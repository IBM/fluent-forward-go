package ws

import (
	"errors"
	"io"
	"log"
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

func NewConnection(conn ext.Conn, opts *ConnectionOptions) (Connection, error) {
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
	var err error

	if wsc.hasConnState(ConnStateClosed) {
		// already closed, nothing else to do
		return nil
	}

	log.Println("received close message", code, text)
	wsc.setConnState(ConnStateCloseReceived)

	// If the peer initiated the close, then this client must send a message
	// confirming receipt. If this client sent the initial close message, then
	// the closing handshake is complete and no further action is required.
	if !wsc.hasConnState(ConnStateCloseSent) {
		log.Println("finalizing handshake", code, text)

		// respond with close; Gorilla doesn't return the error when sending
		// the close response, not sure why
		err = wsc.Close()
	}

	// sent a close and received one, all done
	log.Println("handshake finalized", err)

	return err
}

// CloseWithMsg sends a close message to the peer
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

	if !wsc.hasConnState(ConnStateCloseReceived) {
		select {
		case <-time.After(wsc.closeDeadline):
			log.Println("close deadline expired")
			// sent a close, but never heard back, close anyway
		case <-wsc.closedSig:
		}
	}

	log.Println("closing ws connection")
	wsc.setConnState(ConnStateClosed)

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
	defer func() {
		log.Println("exit Listen")
		wsc.closedSig <- struct{}{}
	}()

	nextMsg := make(chan connMsg)

	go func() {
		msg := connMsg{}

		for {
			msg.mt, msg.message, msg.err = wsc.Conn.ReadMessage()
			nextMsg <- msg

			if msg.err == net.ErrClosed {
				log.Println(msg.err.Error())
				break
			}

			if wsc.hasAnyConnState(ConnStateCloseReceived, ConnStateClosed) {
				log.Println(wsc.connState)
				break
			}
		}

		close(nextMsg)
	}()

	var err error

	for msg := range nextMsg {
		//log.Println("readhandler error:", msg)
		if rerr := wsc.readHandler(wsc, msg.mt, msg.message, msg.err); rerr != nil {
			// enqueue error only if it is something other than a normal close
			if websocket.IsUnexpectedCloseError(rerr, websocket.CloseNormalClosure) {
				log.Println("xreadhandler error:", rerr)
				err = rerr
				return err
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
