package main

import (
	"github.com/hagna/go9p"
	"flag"
	"io"
	"log"
	"os"
)

var debuglevel = flag.Int("d", 0, "debuglevel")
var addr = flag.String("addr", "127.0.0.1:5640", "network address")

func main() {
	var n int
	var user go9p.User
	var err error
	var c *go9p.Clnt
	var file *go9p.File
	var buf []byte

	flag.Parse()
	user = go9p.OsUsers.Uid2User(os.Geteuid())
	go9p.DefaultDebuglevel = *debuglevel
	c, err = go9p.Mount("tcp", *addr, "/", 8192, user)
	if err != nil {
		goto error
	}

	if flag.NArg() != 1 {
		log.Println("invalid arguments")
		return
	}

	file, err = c.FOpen(flag.Arg(0), go9p.OREAD)
	if err != nil {
		goto error
	}

	buf = make([]byte, 8192)
	for {
		n, err = file.Read(buf)
		if n == 0 {
			break
		}

		os.Stdout.Write(buf[0:n])
	}

	file.Close()

	if err != nil && err != io.EOF {
		goto error
	}

	return

error:
	log.Println("Error", err)
}
