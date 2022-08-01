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

### STUCK

Currently this is the step where I get stuck, as the servers start throwing a ton of errors:

```
2022/08/01 14:08:16.195212 [ERR] Received out of order remote server update from: "<SERVER ID>"
2022/08/01 14:08:16.197120 [ERR] SYSTEM - Processing our own account connection event message: ignored
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

**This currently causes an issue where the hub will get 7 messages and the spoke 3 from that one message**