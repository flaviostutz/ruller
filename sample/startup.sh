#!/bin/sh
set -e
set -x

if [ -f /opt/Geolite2-City.mmdb ]; then
    export GEOLITE2_DB=/opt/Geolite2-City.mmdb
fi

echo "Starting Ruller Sample..."
ruller-sample \
    --log-level=$LOG_LEVEL \
    --listen-port=3000 \
    --listen-address="0.0.0.0" \
    --geolite2-db="$GEOLITE2_DB" \
    --city-state-db="/opt/city-state.csv"
    
    