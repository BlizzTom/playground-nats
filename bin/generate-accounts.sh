#!/bin/bash

rm -rf ./nsc
mkdir -p ./nsc/{config,data,store}

NSC_CMD="nsc --config-dir ./nsc/config --data-dir ./nsc/data --keystore-dir ./nsc/store"
export NSC_NO_GIT_IGNORE=true

$NSC_CMD add operator --generate-signing-key --sys --name local
$NSC_CMD add account APP
$NSC_CMD edit account APP --sk generate \
    --js-streams -1 \
    --js-consumer -1 \
    --js-disk-storage -1 \
    --js-ha-resources -1 \
    --js-mem-storage 0
$NSC_CMD add user --account APP leaf
$NSC_CMD add user --account APP pubsub

$NSC_CMD generate config --sys-account=SYS --mem-resolver > ./conf/resolver.conf
