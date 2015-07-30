// Copyright 2009 The go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/rminnich/go9p"
	"log"
)

var addr = flag.String("addr", ":5640", "network address")
var debug = flag.Int("debug", 0, "print debug messages")
var root = flag.String("root", "/", "root filesystem")

func main() {
	flag.Parse()
	ufs := new(go9p.Ufs)
	ufs.Dotu = true
	ufs.Id = "ufs"
	ufs.Root = *root
	ufs.Debuglevel = *debug
	ufs.Start(ufs)

	fmt.Print("ufs starting\n")
	// determined by build tags
	//extraFuncs()
	err := ufs.StartNetListener("tcp", *addr)
	if err != nil {
		log.Println(err)
	}
}
