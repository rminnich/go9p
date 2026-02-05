package go9p

import "testing"

func TestFidFile(t *testing.T) {
	fid := &Fid{Fid: 1}
	file := FidFile(fid, 42)
	if file.Fid != fid {
		t.Fatalf("FidFile fid mismatch")
	}
	if file.offset != 42 {
		t.Fatalf("FidFile offset = %d", file.offset)
	}
}

func TestClntLogFcall(t *testing.T) {
	clnt := &Clnt{Debuglevel: DbgLogPackets | DbgLogFcalls, Log: NewLogger(8)}
	fc := &Fcall{Type: Tversion, Pkt: []byte("pkt")}
	clnt.logFcall(fc)
}
