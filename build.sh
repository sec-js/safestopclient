#!/bin/bash
GOOS=linux GOARCH=amd64 go build -o application -v -ldflags="-s -w"