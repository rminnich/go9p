
// Copyright 2009 The go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go9p

import (
	"flag"
	"net"
	"testing"
)

var addr = flag.String("addr", ":5640", "network address")
var debug = flag.Int("debug", 0, "print debug messages")
var root = flag.String("root", "/", "root filesystem")

func TestAttachOpenReaddir(t*testing.T) {
	var err error
	flag.Parse()
	ufs := new(Ufs)
	ufs.Dotu = true
	ufs.Id = "ufs"
	ufs.Root = *root
	ufs.Debuglevel = *debug
	ufs.Start(ufs)

	t.Log("ufs starting\n");
	// determined by build tags
	//extraFuncs()
	go func () {
		if err = ufs.StartNetListener("tcp", *addr); err != nil {
			t.Fatalf("Can not start listener: %v", err)
		}
	} ()
	/* this may take a few tries ... */
	var conn net.Conn
	for i := 0; i < 16; i++ {
		if conn, err = net.Dial("tcp", *addr); err != nil {
			t.Logf("%v", err)
		} else {
			t.Logf("Got a conn, %v\n", conn)
			break
		}
	}
	if err != nil {
		t.Fatalf("Connect failed after many tries ...")
	}
		
	clnt := NewClnt(conn, 8192, false)
	var rootfid *Fid
	root := OsUsers.Uid2User(0)
	if rootfid, err = clnt.Attach(nil, root, "/tmp"); err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("attached, rootfid %v\n", rootfid)
	var dirfid *Fid
	if dirfid, err = clnt.FWalk("."); err != nil {
		t.Fatalf("%v", err)
	}
	if err = clnt.Open(dirfid, 0); err != nil {
		t.Fatalf("%v", err)
	}
	var b []byte
	if b, err = clnt.Read(dirfid, 0, 64*1024); err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("b %v\n", b)
}
