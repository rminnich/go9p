#!/bin/bash

set -x

if [ "$TRAVIS_GO_VERSION" = "tip" ]; then
	goveralls -service=travis-ci;
else
	go test ./...;
fi
