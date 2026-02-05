package go9p

import (
	"net"
	"strings"
	"testing"
)

func TestConnAddressesAndLogFcall(t *testing.T) {
	c1, c2 := net.Pipe()
	t.Cleanup(func() {
		_ = c1.Close()
		_ = c2.Close()
	})
	srv := &Srv{Log: NewLogger(8)}
	conn := &Conn{Srv: srv, conn: c1, Debuglevel: DbgLogPackets | DbgLogFcalls}

	if conn.RemoteAddr() == nil {
		t.Fatalf("RemoteAddr is nil")
	}
	if conn.LocalAddr() == nil {
		t.Fatalf("LocalAddr is nil")
	}

	fc := &Fcall{Type: Tversion, Pkt: []byte("pkt")}
	conn.logFcall(fc)
}

func TestStartNetListenerError(t *testing.T) {
	srv := &Srv{}
	if err := srv.StartNetListener("invalid", "addr"); err == nil {
		t.Fatalf("expected error")
	} else if !strings.Contains(err.Error(), "unknown network") {
		t.Fatalf("error = %v", err)
	}
}
