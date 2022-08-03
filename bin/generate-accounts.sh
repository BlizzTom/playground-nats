#!/bin/bash

rm -rf ./nsc
mkdir -p ./nsc/{config,keys,store}
mkdir -p ./data/{hub,spoke}/{1,2,3}/jwt

export NSC_NO_GIT_IGNORE=true
export NKEYS_PATH=./nsc/keys
export NSC_HOME=./nsc/config
nsc env -s ./nsc/store

nsc add operator --generate-signing-key --sys --name local
nsc edit operator --service-url nats://localhost:4222

nsc add account APP
nsc edit account APP --sk generate \
    --js-streams -1 \
    --js-consumer -1 \
    --js-disk-storage -1 \
    --js-ha-resources -1 \
    --js-mem-storage 0
nsc add user --account APP leaf
nsc add user --account APP pubsub
