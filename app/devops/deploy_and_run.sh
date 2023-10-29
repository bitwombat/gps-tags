#!/bin/bash

set -e
echo "Stopping service"
ssh -q proxy 'systemctl stop dog-tracking.service'

echo "Building app"
go build -ldflags="-s -w" main.go

echo "Transferring service definition"
scp -C devops/dog-tracking.service proxy:/etc/systemd/system/dog-tracking.service

echo "Transferring application binary"
scp -C main proxy:.

echo "Reloading systemctl daemon"
ssh -q proxy systemctl daemon-reload

echo "Starting service"
ssh -q proxy 'systemctl start dog-tracking.service'
