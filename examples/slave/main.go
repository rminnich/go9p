package main

import (
	"flag"
	"fmt"
	"github.com/hagna/go9p"
	"log"
	"time"
	"os"
)

var addr = flag.String("addr", ":5640", "network address")
var debug = flag.Int("debug", 0, "print debug messages")
var root = flag.String("root", "/", "root filesystem")

type Time struct {
	go9p.SrvFile
}

func (*Time) Read(fid *go9p.FFid, buf []byte, offset uint64) (int, error) {
	b := []byte(time.Now().String())
	have := len(b)
	off := int(offset)

	if off >= have {
		return 0, nil
	}

	return copy(buf, b[off:]), nil
}

func (*Time) Write(fid *go9p.FFid, data []byte, offset uint64) (int, error)  {
	log.Println("fid is", fid)
	log.Println("buf is", string(data))

	log.Println("offset is", offset)
	return len(data), nil
}

func (*Time) Open(fid *go9p.FFid, mode uint8) error {
	log.Println("Open", fid)
	return nil
}


func main() {
	flag.Parse()
	pfs := new(go9p.Procfs)
	ufs := new(go9p.Ufs)
	pfs.Ufs = ufs 
	ufs.Dotu = true
	ufs.Id = "ufs"
	ufs.Root = *root
	ufs.Debuglevel = *debug
//	ufs.statsRegister()
	ufs.Start(pfs)

	user := go9p.OsUsers.Uid2User(os.Geteuid())
	root := new(go9p.SrvFile)
	err := root.Add(nil, "/", user, nil, go9p.DMDIR|0555, nil)
	if err != nil {
		log.Fatal(err)
	}
	tm := new(Time)
	err = tm.Add(root, "time", go9p.OsUsers.Uid2User(os.Geteuid()), nil, 0777, tm)
	if err != nil {
		log.Fatal(err)
	}

	tm2 := new(Time)
	err = tm2.Add(root, "d", go9p.OsUsers.Uid2User(os.Geteuid()), nil, 0444&go9p.DMDIR, tm2)
	if err != nil {
		log.Fatal(err)
	}
	fs := go9p.NewsrvFileSrv(root)
	fs.Dotu = true
	pfs.Fsrv = fs
	fs.Start(pfs)

	fmt.Print("procfs starting\n")
	// determined by build tags
	//extraFuncs()
	err = ufs.StartNetListener("tcp", *addr)
	if err != nil {
		log.Println(err)
	}
}
