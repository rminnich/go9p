// Copyright 2009 The go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go9p

import (
	"flag"
	"io/ioutil"
	"net"
	"os"
	"path"
	"testing"
	"syscall"
)

var addr = flag.String("addr", ":5640", "network address")
var pipefsaddr = flag.String("pipefsaddr", ":5641", "pipefs network address")
var debug = flag.Int("debug", 0, "print debug messages")
var root = flag.String("root", "/", "root filesystem")

// Two files, dotu was true.
var testunpackbytes = []byte{
	79, 0, 0, 0, 0, 0, 0, 0, 0, 228, 193, 233, 248, 44, 145, 3, 0, 0, 0, 0, 0, 164, 1, 0, 0, 0, 0, 0, 0, 47, 117, 180, 83, 102, 3, 0, 0, 0, 0, 0, 0, 6, 0, 112, 97, 115, 115, 119, 100, 4, 0, 110, 111, 110, 101, 4, 0, 110, 111, 110, 101, 4, 0, 110, 111, 110, 101, 0, 0, 232, 3, 0, 0, 232, 3, 0, 0, 255, 255, 255, 255, 78, 0, 0, 0, 0, 0, 0, 0, 0, 123, 171, 233, 248, 42, 145, 3, 0, 0, 0, 0, 0, 164, 1, 0, 0, 0, 0, 0, 0, 41, 117, 180, 83, 195, 0, 0, 0, 0, 0, 0, 0, 5, 0, 104, 111, 115, 116, 115, 4, 0, 110, 111, 110, 101, 4, 0, 110, 111, 110, 101, 4, 0, 110, 111, 110, 101, 0, 0, 232, 3, 0, 0, 232, 3, 0, 0, 255, 255, 255, 255,
}

func TestUnpackDir(t *testing.T) {
	b := testunpackbytes
	for len(b) > 0 {
		var err error
		if _, b, _, err = UnpackDir(b, true); err != nil {
			t.Fatalf("Unpackdir: %v", err)
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
	for b != nil && len(b) > 0 {
		var d *Dir
		t.Logf("len(b) %v\n", len(b))
		if d, b, _, err = UnpackDir(b, ufs.Dotu); err != nil {
			t.Fatalf("Unpackdir: %v", err)
		} else {
			t.Logf("Unpacked: %d \n", d)
			t.Logf("b len now %v\n", len(b))
		}
	}
	// now test partial reads.
	// Read 128 bytes at a time. Remember the last successful offset.
	// if UnpackDir fails, read again from that offset
	t.Logf("NOW TRY PARTIAL")
	offset := uint64(0)
	for {
		var b []byte
		var d *Dir
		var amt int
		if b, err = clnt.Read(dirfid, offset, 128); err != nil {
			t.Fatalf("%v", err)
		}
		if len(b) == 0 {
			break
		}
		t.Logf("b %v\n", b)
		for b != nil && len(b) > 0 {
			t.Logf("len(b) %v\n", len(b))
			if d, b, amt, err = UnpackDir(b, ufs.Dotu); err != nil {
				// this error is expected ...
				t.Logf("unpack failed (it's ok!). retry at offset %v\n", 
					offset)
				break
			} else {
				t.Logf("d %v\n", d)
				offset += uint64(amt)
			}
		}
	}
}

var f *File
var b = make([]byte, 1048576/8)

// Not sure we want this, and the test has issues. Revive it if we ever find a use for it.
func TestPipefs(t *testing.T) {
	pipefs := new(Pipefs)
	pipefs.Dotu = false
	pipefs.Msize = 1048576
	pipefs.Id = "pipefs"
	pipefs.Root = *root
	pipefs.Debuglevel = *debug
	pipefs.Start(pipefs)

	t.Logf("pipefs starting\n");
	d, err := ioutil.TempDir("", "TestPipeFS")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := os.Remove(d); err != nil {
			t.Fatalf("%v", err)
		}
	}()
	fn := path.Join(d, "fifo")
	if err := syscall.Mkfifo(fn, 0600); err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		if err := os.Remove(fn); err != nil {
			t.Fatalf("%v", err)
		}
	}()
	// determined by build tags
	//extraFuncs()
	go func() {
		err := pipefs.StartNetListener("tcp", *pipefsaddr)
		if err != nil {
			t.Fatalf("StartNetListener failed: %v\n", err)
		}
	}()
	root := OsUsers.Uid2User(0)

	var c *Clnt
	for i := 0; i < 16; i++ {
		c, err = Mount("tcp", *pipefsaddr, "/", uint32(len(b)), root)
	}
	if err != nil {
		t.Fatalf("Connect failed: %v\n", err)
	}
	t.Logf("Connected to %v\n", *c)
	if f, err = c.FOpen(fn, ORDWR); err != nil {
		t.Fatalf("Open failed: %v\n", err)
	} else {
		for i := 0; i < 1048576/8; i++ {
			b[i] = byte(i)
		}
		t.Logf("f %v \n", f)
		if n, err := f.Write(b); err != nil {
			t.Fatalf("write failed: %v\n", err)
		} else {
			t.Logf("Wrote %v bytes\n", n)
		}
		if n, err := f.Read(b); err != nil {
			t.Fatalf("read failed: %v\n", err)
		} else {
			t.Logf("read %v bytes\n", n)
		}
		
	}
}

func BenchmarkPipeFS(bb *testing.B) {
		for i := 0; i < bb.N; i++ {
			if _, err := f.Write(b); err != nil {
				bb.Errorf("write failed: %v\n", err)
			}
			if _, err := f.Read(b); err != nil {
				bb.Errorf("read failed: %v\n", err)
			}
		}
}
