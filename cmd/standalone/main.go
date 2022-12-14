package main

import (
	"sync"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	nats "github.com/nats-io/nats.go"
	"gopkg.battle.net/logging"
)

func main() {
	natsServer := natsserver.New(&natsserver.Options{
		ServerName: "embeded",
		DontListen: true,
		JetStream:  true,
	})

	go natsServer.Start()

	if !natsServer.ReadyForConnections(time.Second * 10) {
		panic("not ready in time")
	}

	nc, err := nats.Connect("", nats.InProcessServer(natsServer))
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	sub, err := nc.Subscribe("hello", func(msg *nats.Msg) {
		_ = msg.Ack()
		logging.Infof("Hello, %s!", string(msg.Data))
		wg.Done()
	})

	if err != nil {
		panic(err)
	}

	if err := nc.Publish("hello", []byte("world")); err != nil {
		panic(err)
	}

	wg.Wait()

	_ = sub.Drain()
	nc.Close()

	natsServer.Shutdown()
	natsServer.WaitForShutdown()
}
