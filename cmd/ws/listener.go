package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/IBM/fluent-forward-go/fluent/client/ws"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type HttpServer interface {
	Shutdown(ctx context.Context) error
	ListenAndServe() error
	Close() error
}

type Listener struct {
	upgrader       *websocket.Upgrader
	server         HttpServer
	shutdown       chan struct{}
	exited         chan struct{}
	connectionLock sync.Mutex
	wsopts         ws.ConnectionOptions
}

func NewListener(server HttpServer, wsopts ws.ConnectionOptions) *Listener {
	return &Listener{
		&websocket.Upgrader{},
		server,
		make(chan struct{}, 1),
		make(chan struct{}, 1),
		sync.Mutex{},
		wsopts,
	}
}

func (s *Listener) Connect(w http.ResponseWriter, r *http.Request) {
	var (
		c   *websocket.Conn
		err error
	)

	if c, err = s.upgrader.Upgrade(w, r, nil); err == nil {
		connection, _ := ws.NewConnection(c, s.wsopts)

		defer func() {
			if err := connection.Close(); err != nil && err != websocket.ErrCloseSent {
				log.Println("close error:", err)
			}
		}()

		if err = connection.Listen(); err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			log.Println("run error:", err)
		}
	}
}

var (
	once   sync.Once
	router *mux.Router
)

func (s *Listener) ListenAndServe() error {
	defer func() { s.exited <- struct{}{} }()

	once.Do(func() {
		// TODO comment explaining this weirdness
		router = mux.NewRouter()
		http.Handle("/", router)
	})

	router.HandleFunc("/", s.Connect)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("ListenAndServe error: " + err.Error())
		}
	}()

	<-s.shutdown
	log.Println("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var err error
	if err = s.server.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		log.Println("shutdown server error:", err)
		if err = s.server.Close(); err != nil {
			log.Println("close server error:", err)
		}
	}

	if ctx.Err() != nil {
		log.Println("close context error:", ctx.Err())
	}

	s.exited <- struct{}{}
	log.Println("signaled exit")

	return err
}

func (s *Listener) Shutdown() {
	s.shutdown <- struct{}{}
	<-s.exited
}
