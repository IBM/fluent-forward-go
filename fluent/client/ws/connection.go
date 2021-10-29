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
	ID            string
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
	id            string
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
		// default handler ensure read buffer is emptied
		opts.ReadHandler = func(c Connection, _ int, _ []byte, err error) error {
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

	wsc.id = opts.ID

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

func (wsc *connection) HandleClose(_ Connection, code int, text string) error {
	var err error

	if wsc.hasConnState(ConnStateClosed) {
		// already closed, nothing else to do
		return nil
	}

	log.Println(wsc.id, "received close message", code, text)
	wsc.setConnState(ConnStateCloseReceived)

	// If the peer initiated the close, then this client must send a message
	// confirming receipt. If this client sent the initial close message, then
	// the closing handshake is complete and no further action is required.
	if !wsc.hasConnState(ConnStateCloseSent) {
		log.Println(wsc.id, "finalizing handshake", code, text)

		// respond with close; Gorilla doesn't return the error from close response,
		// not sure why
		err = wsc.Close()
	}

	// sent a close and received one, all done
	log.Println(wsc.id, "handshake finalized", err)

	return err
}

// CloseWithMsg sends a close message to the peer
func (wsc *connection) CloseWithMsg(closeCode int, msg string) error {
	if wsc.hasConnState(ConnStateCloseSent) {
		return errors.New("multiple close calls")
	}

	defer log.Println(wsc.id, "ws connection closed")

	wsc.unsetConnState(ConnStateOpen)

	log.Println(wsc.id, "sending close message")

	err := wsc.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(
			closeCode, msg,
		),
	)

	wsc.setConnState(ConnStateCloseSent)

	if err != nil && err != websocket.ErrCloseSent {
		log.Println(wsc.id, "write close failed", err)
	}

	if err == nil && wsc.hasConnState(ConnStateListening) && !wsc.hasConnState(ConnStateCloseReceived) {
		log.Println(wsc.id, "waiting for close response")

		select {
		case <-time.After(wsc.closeDeadline):
			// sent a close, but never heard back, close anyway
			log.Println(wsc.id, "close deadline expired")

			err = errors.New("close deadline expired")
		case <-wsc.closedSig:
			log.Println(wsc.id, "received close sig")
		}
	}

	log.Println(wsc.id, "closing ws connection")
	wsc.setConnState(ConnStateClosed)

	// spec says that only server should close the network connection,
	// but consensus is that it doesn't matter
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

	nextMsg := make(chan *connMsg)

	go func() {
		defer func() {
			wsc.unsetConnState(ConnStateListening)
			wsc.closedSig <- struct{}{}
			log.Println(wsc.id, "exit ReadMessage loop")
		}()

		for {
			msg := &connMsg{}
			msg.mt, msg.message, msg.err = wsc.Conn.ReadMessage()

			log.Printf("%s next message: %+v", wsc.id, msg)

			if msg.err == net.ErrClosed {
				log.Println(wsc.id, "network connection closed", msg.err.Error())
				break
			}

			nextMsg <- msg

			if err, ok := msg.err.(net.Error); ok {
				log.Println(wsc.id, "net error:", err.Error())
				break
			}

			if wsc.hasAnyConnState(ConnStateCloseReceived, ConnStateClosed) {
				log.Println(wsc.id, "breaking ReadMessage loop with connState: ", wsc.connState)
				break
			}
		}

		close(nextMsg)
	}()

	var err error

	for msg := range nextMsg {
		log.Printf("%s readhandler: %+v", wsc.id, msg)

		if rerr := wsc.readHandler(wsc, msg.mt, msg.message, msg.err); rerr != nil {
			// enqueue error only if it is something other than a normal close
			log.Printf("%s readhandler err: %+v", wsc.id, msg.err)

			if websocket.IsUnexpectedCloseError(rerr, websocket.CloseNormalClosure) {
				err = rerr
			}
		}
	}

	log.Println(wsc.id, "exit Listen with error val:", err)

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
