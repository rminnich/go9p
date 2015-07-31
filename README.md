[![Build Status](https://travis-ci.org/rminnich/go9p.svg?branch=master)](https://travis-ci.org/rminnich/go9p)
[![Coverage Status](https://coveralls.io/repos/rminnich/go9p/badge.svg?branch=master&service=github)](https://coveralls.io/github/rminnich/go9p)

This is go9p done in a way that I can understand.

To install:

    export GOPATH=~/go
    go get -a github.com/rminnich/go9p
    go get -a github.com/rminnich/go9p/ufs
    go install -a github.com/rminnich/go9p/ufs

Then to start serving the root fs via 9p at port 5640:

    ~/go/bin/ufs

