#!/bin/bash

set -e
export RSYNC_RSH="ssh -q"

export RSYNC_OPTS="-r --times --no-owner --delete --compress --itemize-changes "

echo "Building app"
pushd service
go build -ldflags="-s -w" main.go
popd

echo "Transferring assets"
rsync $RSYNC_OPTS ./{named_zones,public_html} proxy:

echo "Transferring service definition"
rsync $RSYNC_OPTS devops/dog-tracking.service proxy:/etc/systemd/system/dog-tracking.service

echo "Reloading systemctl daemon"
ssh -q proxy systemctl daemon-reload

echo "Stopping service"
ssh -q proxy 'systemctl stop dog-tracking.service'

echo "Transferring application binary"
scp -C service/main proxy:.

echo "Starting service"
ssh -q proxy 'systemctl start dog-tracking.service'
