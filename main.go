package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	nats "github.com/nats-io/nats.go"
)

var (
	address = flag.String("address", nats.DefaultURL, "NATS address")
	subject = flag.String("subject", "order.>", "subject to subscribe to")
)

func main() {
	flag.Parse()

	opts := nats.Options{
		AllowReconnect:     true,
		MaxReconnect:       -1, // connect forever
		ReconnectWait:      nats.DefaultReconnectWait,
		ReconnectJitter:    nats.DefaultReconnectJitter,
		ReconnectJitterTLS: nats.DefaultReconnectJitterTLS,
		Timeout:            nats.DefaultTimeout,
		PingInterval:       nats.DefaultPingInterval,
		MaxPingsOut:        nats.DefaultMaxPingOut,
		SubChanLen:         nats.DefaultMaxChanLen,
		ReconnectBufSize:   nats.DefaultReconnectBufSize,
		DrainTimeout:       nats.DefaultDrainTimeout,
		Servers:            []string{*address},
		Name:               "Demo Client",
		AsyncErrorCB: func(conn *nats.Conn, sub *nats.Subscription, err error) {
			print("Slow subscriber %s", sub.Subject)
		},
		ClosedCB: func(conn *nats.Conn) {
			print("Connection to server closed")
		},
		DisconnectedErrCB: func(conn *nats.Conn, err error) {
			print("Disconnected from server: %v", err)
		},
		DiscoveredServersCB: func(conn *nats.Conn) {
			print("Discovered new server: %s", conn.ConnectedAddr())
		},
		ReconnectedCB: func(conn *nats.Conn) {
			print("Reconnected to server")
		},
	}

	nc, err := opts.Connect()
	if err != nil {
		printE(err)
	}
	defer func() { _ = nc.Drain() }()
	defer nc.Close()

	js, err := nc.JetStream(nats.PublishAsyncMaxPending(1))
	if err != nil {
		printE(err)
	}

	print("Connected to Server Status: %v", nc.Status())
	sub, err := js.Subscribe(*subject, func(m *nats.Msg) {
		print("Received a message: %s\n%s\n", m.Subject, string(m.Data))
	})

	if err != nil {
		printE(err)
	}

	schan := make(chan os.Signal, 6)
	defer close(schan)

	signal.Notify(schan, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGHUP, syscall.Signal(21))
	defer signal.Stop(schan)

	<-schan

	if err := sub.Drain(); err != nil {
		printE(err)
	}
	if err := sub.Unsubscribe(); err != nil {
		printE(err)
	}
}

func print(m string, args ...any) {
	fmt.Fprintf(os.Stdout, m+"\n", args...)
}

func printE(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	os.Exit(1)
}
