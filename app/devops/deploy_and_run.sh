#!/bin/bash

VPS_FQDN=proxy

set -e
export RSYNC_RSH="ssh -q"

export RSYNC_OPTS="-r --times --no-owner --delete --compress --itemize-changes "

echo "Building app"
pushd service
go build -ldflags="-s -w" main.go
popd

echo "Transferring assets"
rsync $RSYNC_OPTS ./{named_zones,public_html} "$VPS_FQDN":

echo "Transferring service definition"
rsync $RSYNC_OPTS devops/dog-tracking.service "$VPS_FQDN":/etc/systemd/system/dog-tracking.service

echo "Reloading systemctl daemon"
ssh -q "$VPS_FQDN" systemctl daemon-reload

echo "Stopping service"
ssh -q "$VPS_FQDN" 'systemctl stop dog-tracking.service'

echo "Transferring application binary"
scp -C service/main "$VPS_FQDN":.

echo "Starting service"
ssh -q "$VPS_FQDN" 'systemctl start dog-tracking.service'
