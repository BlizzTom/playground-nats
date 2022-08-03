#!/bin/bash

# Purge out the jwt tokens
rm -rf ./data/*/jwt
rm -rf ./data/*/*/jwt

# Purge out the jetstream
rm -rf ./data/*/jetstream
rm -rf ./data/*/*/jetstream
