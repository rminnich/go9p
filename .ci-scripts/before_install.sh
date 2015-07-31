#!/bin/bash

set -x

if [ "$TRAVIS_GO_VERSION" = "tip" ]; then
	go get github.com/axw/gocov/gocov
	go get github.com/mattn/goveralls
	if ! go get code.google.com/p/go.tools/cmd/cover; then 
		go get golang.org/x/tools/cmd/cover; 
	fi
fi

if [ "$UPSTREAM_OWNER" != "" ]; then
	REPO_NAME=`basename $TRAVIS_REPO_SLUG`
	mkdir ../../$UPSTREAM_OWNER && ln -s `pwd` ../../$UPSTREAM_OWNER/$REPO_NAME
fi

