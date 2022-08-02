# NATS Playground

Just a playground for NATS things. This is just for experimentations and figuring things out, and shouldn't be used in production or as a guide to do anything useful.

This repository uses [forego](https://github.com/ddollar/forego), however, the commands can be found in the [Procfile](./Procfile).

Nats CLI: 0.0.33
Nats Server: 2.8.4

## Keys

NKeys used, don't use them, that would be silly.

## Jetstream Standalone

This will startup a standalone jetstream enabled NATS server

```bash
nats-server --config ./conf/standalone.conf
```

## Jetstream Cluster with Leaf Nodes (Hub+Spoke)

This will startup a NATS Cluster (`hub`) with Jetstream running, and a Leaf Node (`spoke`) with Jetstream running connected to the `hub`

### Hub Nodes

```bash
nats-server --config ./conf/hub.1.conf
nats-server --config ./conf/hub.2.conf
nats-server --config ./conf/hub.3.conf
```

### Spoke Nodes

```bash
nats-server --config ./conf/spoke.1.conf
nats-server --config ./conf/spoke.2.conf
nats-server --config ./conf/spoke.3.conf
```

### Creating the streams

```bash
nats -s 'nats://localhost:4222' --creds=./nsc/store/creds/local/APP/pubsub.creds --js-domain hub stream add --config ./data/orders.json
nats -s 'nats://localhost:4222' ---creds=./nsc/store/creds/local/APP/pubsub.creds --js-domain spoke stream add --config ./data/orders.json
```

### Create the consumer

```bash
nats -s 'nats://localhost:4222' ---creds=./nsc/store/creds/local/APP/pubsub.creds consumer add ORDERS TEST --target="order.received" --ack explicit --deliver all --max-deliver=-1 --sample 100 --replay=instant --filter="" --max-pending=0 --no-headers-only --backoff=none --deliver-group="" --heartbeat=-1 --no-flow-control
```

### Publish a message

```bash
nats -s 'nats://localhost:4222' ---creds=./nsc/store/creds/local/APP/pubsub.creds pub order.1 "order 1"
```

