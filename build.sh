#!/bin/sh
CGO_ENABLED=0 go build -o oodles -v -ldflags="-buildid= -X github.com/starshine-sys/oodles/common.Version=`git rev-parse --short HEAD`"
strip oodles
