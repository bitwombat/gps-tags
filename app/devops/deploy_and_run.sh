#!/bin/bash

ssh proxy 'killall -9 main'
go build main.go && scp main proxy:. && ssh proxy ./main
