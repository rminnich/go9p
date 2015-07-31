#!/bin/bash

set -x

if [ "" = "tip" ]; then
	goveralls -service=travis-ci;
else
	go test ./...;
fi
