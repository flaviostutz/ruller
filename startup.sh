#!/bin/bash
set -e
set -x

echo "Starting Ruller Sample..."
ruller \
    --log-level=$LOG_LEVEL \
    --listen-port=$LISTEN_PORT \
    --listen-address=$LISTEN_ADDRESS
    