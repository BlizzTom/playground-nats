# NATS Playground

Just a playground for NATS things. This is just for experimentations and figuring things out, and shouldn't be used in production or as a guide to do anything useful.

## Start servers

This repository uses [forego](https://github.com/ddollar/forego), however, the commands can be found in the [Procfile](./Procfile).

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

### Standalone

```bash
nats-server --config ./conf/standalone.conf
```
