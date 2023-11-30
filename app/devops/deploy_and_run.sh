#!/bin/bash

set -e

# Vars
SERVICE_FILE="dog-tracking.service"


# Commands
export RSYNC_RSH="ssh -q"
export RSYNC_OPTS="-r --times --no-owner --exclude=public_html/log.html --delete --compress --itemize-changes "


# Read and check .env file
pushd devops &> /dev/null
if [[ ! -e .env ]]; then
    echo "ERROR: .env file not found. Please copy .env.example to .env and then edit it with your details."
    exit 1
fi

. .env

if [[ -z "$VPS_FQDN" ]]; then
    echo "ERROR: VPS_FQDN not set in .env file."
    exit 1
fi
popd &> /dev/null


# Build app
echo "Building app"
pushd service &> /dev/null
go build -ldflags="-s -w" .
popd &> /dev/null


# Copy everything over to VPS and restart service
echo "Transferring assets"
rsync $RSYNC_OPTS ./{named_zones,public_html} "$VPS_FQDN":

echo "Transferring service definition"
rsync $RSYNC_OPTS devops/"$SERVICE_FILE" "$VPS_FQDN":/etc/systemd/system/"$SERVICE_FILE"

echo "Reloading systemctl daemon"
ssh -q "$VPS_FQDN" systemctl daemon-reload

echo "Stopping service"
ssh -q "$VPS_FQDN" "systemctl stop $SERVICE_FILE"

echo "Transferring application binary"
scp -C service/gps-tags "$VPS_FQDN":.

echo "Starting service"
ssh -q "$VPS_FQDN" "systemctl start $SERVICE_FILE"
