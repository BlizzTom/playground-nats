# NATS Playground

Just a playground for NATS things. This is just for experimentations and figuring things out, and shouldn't be used in production or as a guide to do anything useful.

This repository uses [forego](https://github.com/ddollar/forego), however, the commands can be found in the [Procfile](./Procfile).

Nats CLI: 0.0.33
Nats Server: 2.8.4

## Authentication/Authorizations

This repo contains a subfolder `nsc` that contains the entire keystore for the clusters, Intentionally the keys are stored in here without the `.gitignore` for purposes of the demo, don't do this!!!!

### Working with NSC in this repo

To interact with `nsc`, you will need to setup some environment values:

```bash
export NSC_NO_GIT_IGNORE=true
export NKEYS_PATH=./nsc/keys
export NSC_HOME=./nsc/config
nsc env -s ./nsc/store
```

Use `nsc describe account` to verify it is working and displays information about the `APP` account.

### Regenerate all keys

To generate a new set of keys:

```bash
./bin/generate-accounts.sh
```

Then need to update the [./conf/leaf-remotes.conf](./conf/leaf-remotes.conf) with the correct account NKey Public keys for SYS and APP accounts. These can be found using `nsc describe account --name=<name>` and use the `Account ID` value.


## Jetstream Standalone

This will startup a standalone jetstream enabled NATS server

```bash
nats-server --config ./conf/standalone.conf
```

After the server is up, you need to run the `nsc push -A -u nats://localhost:4222` command. This will push the accounts and users to the nats server.

## Jetstream Cluster with Leaf Nodes (Hub+Spoke)

This will startup a NATS Cluster (`hub`) with Jetstream running, and a Leaf Node (`spoke`) with Jetstream running connected to the `hub`

### Hub Nodes

```bash
nats-server --config ./conf/hub.1.conf
nats-server --config ./conf/hub.2.conf
nats-server --config ./conf/hub.3.conf
```

After the server is up, you need to run the `nsc push -A -u nats://localhost:4222` command. This will push the accounts and users to the nats server.

### Spoke Nodes

```bash
nats-server --config ./conf/spoke.1.conf
nats-server --config ./conf/spoke.2.conf
nats-server --config ./conf/spoke.3.conf
```

After the server is up, you need to run the `nsc push -A -u nats://localhost:4222` command. This will push the accounts and users to the nats server.

### Create some contexts

```
nats context save --server nats://localhost:4222 --creds ./nsc/keys/creds/local/APP/pubsub.creds local-app
nats context save --server nats://localhost:4222 --creds ./nsc/keys/creds/local/SYS/sys.creds local-sys
```

### Validate the servers are clustered

```
nats context select local-sys
nats server ls
```

### Creating the streams

```bash
nats context select local-app
nats --js-domain hub stream add --config ./data/orders.json
nats --js-domain spoke stream add --config ./data/orders.json
```

### Create the consumer

```bash
nats consumer add ORDERS TEST --target="orders.received" --ack explicit --deliver all --max-deliver=-1 --sample 100 --replay=instant --filter="" --max-pending=0 --no-headers-only --backoff=none --deliver-group="" --heartbeat=-1 --no-flow-control
```

### Publish a message

```bash
nats pub order.1 "order 1"
```

