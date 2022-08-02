package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

var (
	address  = flag.String("address", nats.DefaultURL, "NATS address")
	subject  = flag.String("subject", "orders.received", "subject to subscribe to")
	stream   = flag.String("stream", "ORDERS", "stream to use")
	consumer = flag.String("consumer", "TEST", "consumer group to use")
	creds    = flag.String("creds", "./nsc/store/creds/local/APP/pubsub.creds", "Credentials")
)

func main() {
	flag.Parse()

	nkeyBits, err := os.ReadFile(*creds)
	if err != nil {
		printE(err)
	}

	token, err := nkeys.ParseDecoratedJWT(nkeyBits)
	if err != nil {
		printE(err)
	}
	nkey, err := nkeys.ParseDecoratedNKey(nkeyBits)
	if err != nil {
		printE(err)
	}

	nkeyPublic, err := nkey.PublicKey()
	if err != nil {
		printE(err)
	}

	nkeySeed, err := nkey.Seed()
	if err != nil {
		printE(err)
	}

	print("Public Key: %s", nkeyPublic)
	print("Private Key: %s", string(nkeySeed))

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
			print("Slow subscriber %q", sub.Subject)
		},
		ClosedCB: func(conn *nats.Conn) {
			print("Connection to server %q closed", conn.ConnectedServerName())
		},
		DisconnectedErrCB: func(conn *nats.Conn, err error) {
			if err != nil {
				print("Disconnected from server %q: %v", conn.ConnectedServerName(), err)
			}
		},
		DiscoveredServersCB: func(conn *nats.Conn) {
			print("Discovered new server: %q", conn.ConnectedServerName())
		},
		UserJWT: func() (string, error) {
			return token, nil
		},
		ReconnectedCB: func(conn *nats.Conn) {
			print("Reconnected to server %q", conn.ConnectedServerName())
		},
		SignatureCB: func(data []byte) ([]byte, error) {
			key, err := nkeys.FromSeed(nkeySeed)
			if err != nil {
				printE(err)
			}

			return key.Sign(data)
		},
	}

	nc, err := opts.Connect()
	if err != nil {
		printE(err)
	}
	defer func() { _ = nc.Drain() }()
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		printE(err)
	}

	print("%v to Server %s", nc.Status(), nc.ConnectedServerName())
	subOpts := []nats.SubOpt{}

	if *consumer != "" {
		print("Joining Consumer %s", *consumer)
		subOpts = append(subOpts, nats.Bind(*stream, *consumer))
	}

	sub, err := js.Subscribe(*subject, func(m *nats.Msg) {
		print("Received a message: %s\n%s\n", m.Subject, string(m.Data))
		if err := m.AckSync(); err != nil {
			print("Failed to ack message: %v", err)
		}
		print("Message Acked")
	}, subOpts...)

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
