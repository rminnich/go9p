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

// Two files, dotu was true.
var testunpackbytes = []byte{
	79, 0, 0, 0, 0, 0, 0, 0, 0, 228, 193, 233, 248, 44, 145, 3, 0, 0, 0, 0, 0, 164, 1, 0, 0, 0, 0, 0, 0, 47, 117, 180, 83, 102, 3, 0, 0, 0, 0, 0, 0, 6, 0, 112, 97, 115, 115, 119, 100, 4, 0, 110, 111, 110, 101, 4, 0, 110, 111, 110, 101, 4, 0, 110, 111, 110, 101, 0, 0, 232, 3, 0, 0, 232, 3, 0, 0, 255, 255, 255, 255, 78, 0, 0, 0, 0, 0, 0, 0, 0, 123, 171, 233, 248, 42, 145, 3, 0, 0, 0, 0, 0, 164, 1, 0, 0, 0, 0, 0, 0, 41, 117, 180, 83, 195, 0, 0, 0, 0, 0, 0, 0, 5, 0, 104, 111, 115, 116, 115, 4, 0, 110, 111, 110, 101, 4, 0, 110, 111, 110, 101, 4, 0, 110, 111, 110, 101, 0, 0, 232, 3, 0, 0, 232, 3, 0, 0, 255, 255, 255, 255,
}

func TestUnpackDir(t *testing.T) {
	b := testunpackbytes
	for len(b) > 0 {
		var d *Dir
		var err error
		t.Logf("len(b) %v\n", len(b))
		t.Logf("b %v\n", b)
		if d, b, err = UnpackDir(b, true); err != nil {
			t.Fatalf("Unpackdir: %v", err)
		} else {
			t.Logf("Unpacked: %d \n", d)
			t.Logf("b len now %v\n", len(b))
		}
	}
}

func TestAttachOpenReaddir(t *testing.T) {
	var err error
	flag.Parse()
	ufs := new(Ufs)
	ufs.Dotu = false
	ufs.Id = "ufs"
	ufs.Root = *root
	ufs.Debuglevel = *debug
	ufs.Start(ufs)

	t.Log("ufs starting\n")
	// determined by build tags
	//extraFuncs()
	go func() {
		if err = ufs.StartNetListener("tcp", *addr); err != nil {
			t.Fatalf("Can not start listener: %v", err)
		}
	}()
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
	dirfid := clnt.FidAlloc()
	if _, err = clnt.Walk(rootfid, dirfid, []string{"."}); err != nil {
		t.Fatalf("%v", err)
	}
	if err = clnt.Open(dirfid, 0); err != nil {
		t.Fatalf("%v", err)
	}
	var b []byte
	if b, err = clnt.Read(dirfid, 0, 64*1024); err != nil {
		t.Fatalf("%v", err)
	}
	for len(b) > 0 {
		var d *Dir
		t.Logf("len(b) %v\n", len(b))
		if d, b, err = UnpackDir(b, ufs.Dotu); err != nil {
			t.Fatalf("Unpackdir: %v", err)
		} else {
			t.Logf("Unpacked: %d \n", d)
			t.Logf("b len now %v\n", len(b))
		}
	}
}
