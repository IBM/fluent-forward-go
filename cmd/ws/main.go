package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/IBM/fluent-forward-go/fluent/client"
	"github.com/IBM/fluent-forward-go/fluent/client/ws"
	"github.com/IBM/fluent-forward-go/fluent/protocol"
)

var (
	tagVar string
	useTLS bool
)

func init() {
	flag.StringVar(&tagVar, "tag", "test.message", "-tag <dot-delimited tag>")
	flag.StringVar(&tagVar, "t", "test.message", "-t <dot-delimited tag> (shorthand for -tag)")
	flag.BoolVar(&useTLS, "s", false, "specify to use tls")
}

//nolint
func listen() *Listener {

	log.Println("Starting server on port 8083")

	s := &http.Server{Addr: ":8083"}
	wo := ws.ConnectionOptions{
		ReadHandler: func(conn ws.Connection, _ int, p []byte, err error) error {
			msg := protocol.Message{}
			msg.UnmarshalMsg(p)

			log.Println("server got a message", msg, err)

			return err
		},
	}

	wsSvr := NewListener(s, wo)

	go func() {
		if err := wsSvr.ListenAndServe(); err != nil {
			panic("ListenAndServe: " + err.Error())
		}
	}()

	return wsSvr
}

func main() {
	flag.Parse()

	var tlsCfg *tls.Config

	url := "ws://127.0.0.1:8083"
	if useTLS {
		url = "wss://127.0.0.1:8083"
		tlsCfg = &tls.Config{InsecureSkipVerify: true} //#nosec
	}

	fmt.Fprintln(os.Stderr, "Connecting to - ", url)

	c := client.NewWS(client.WSConnectionOptions{
		Factory: &client.DefaultWSConnectionFactory{
			URL:       url,
			TLSConfig: tlsCfg,
		},
	})

	wsSvr := listen()

	time.Sleep(time.Second)

	err := c.Connect()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to connect, exiting", err)
		os.Exit(1)
	}

	fmt.Println("Creating message")

	msg := protocol.Message{
		Tag:       tagVar,
		Timestamp: time.Now().UTC().Unix(),
		Record: map[string]interface{}{
			"first": "Sir",
			"last":  "Gawain",
			"enemy": "Green Knight",
		},
	}

	if err := c.Send(&msg); err != nil {
		log.Fatal(err)
	}

	msg = protocol.Message{
		Tag:       tagVar,
		Timestamp: time.Now().UTC().Unix(),
		Record: map[string]interface{}{
			"first": "Sir",
			"last":  "Lancelot",
			"enemy": "Himself",
		},
	}

	if err := c.Send(&msg); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Messages sent")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		if err := c.Disconnect(); err != nil {
			log.Fatal(err)
		}

		wsSvr.Shutdown()
		interrupt <- os.Interrupt
	}()

	<-interrupt

	os.Exit(0)
}
