# NATS Playground

Just a playground for NATS things. This is just for experimentations and figuring things out, and shouldn't be used in production or as a guide to do anything useful.

This repository uses [forego](https://github.com/ddollar/forego), however, the commands can be found in the [Procfile](./Procfile).

## Jetstream Standalone

This will startup a standalone jetstream enabled NATS server

```bash
nats-server --config ./conf/standalone.conf
```

## Jetstream Cluster with Leaf Nodes (Hub+Spoke)

This will startup a NATS Cluster (`hub`) with Jetstream running, and a Leaf Node (`spoke`) with Jetstream running connected to the `hub`

### Cluster

```bash
nats-server --config ./conf/1.conf
nats-server --config ./conf/2.conf
nats-server --config ./conf/3.conf
```

### Leaf Node

```bash
nats-server --config ./conf/leaf.conf
```

### Creating a Stream

```bash
nats -s 'nats://leaf:password@localhost:4222' --js-domain hub stream add --config ./data/orders.json
nats -s 'nats://leaf:password@localhost:4222' --js-domain spoke stream add --config ./data/orders.json
```

### Validate the streams

```bash
nats -s 'nats://leaf:password@localhost:4222' --js-domain hub stream info ORDERS
nats -s 'nats://leaf:password@localhost:4222' --js-domain spoke stream info ORDERS
```

### Publish a message

```bash
nats -s 'nats://leaf:password@localhost:4222' pub order.1 "order 1"
```

**This currently causes an issue where the hub will get 7 messages and the spoke 3 from that one message**