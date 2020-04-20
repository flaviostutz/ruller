#!/bin/sh
set -e
set -x

echo "Starting Ruller Sample..."
ruller-sample \
    --log-level=$LOG_LEVEL \
    --listen-port=$LISTEN_PORT \
    --listen-address=$LISTEN_ADDRESS \
    --geolite2-db=$GEOLITE2_DB \
    --city-state-db=$CITY_STATE_DB
    
    