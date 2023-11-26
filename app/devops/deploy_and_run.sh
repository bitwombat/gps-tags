#!/bin/bash

set -e

RSYNC='rsync -r --compress --itemize-changes '

echo "Building app"
go build -ldflags="-s -w" main.go

echo "Transferring assets"
$RSYNC ../{zones,public_html} proxy:

echo "Transferring service definition"
$RSYNC devops/dog-tracking.service proxy:/etc/systemd/system/dog-tracking.service

echo "Reloading systemctl daemon"
ssh -q proxy systemctl daemon-reload

echo "Stopping service"
ssh -q proxy 'systemctl stop dog-tracking.service'

echo "Transferring application binary"
scp -C main proxy:.

echo "Starting service"
ssh -q proxy 'systemctl start dog-tracking.service'
