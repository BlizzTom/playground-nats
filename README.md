# NATS Playground

Just a playground for NATS things. This is just for experimentations and figuring things out, and shouldn't be used in production or as a guide to do anything useful.

This repository uses [forego](https://github.com/ddollar/forego), however, the commands can be found in the [Procfile](./Procfile).

Nats CLI: 0.0.33
Nats Server: 2.8.4

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

