package go9p

import (
	"errors"
	"testing"
)

func newTestReq(msgType uint8) *SrvReq {
	srv := &Srv{Dotu: true, Msize: MSIZE, Upool: OsUsers}
	conn := &Conn{
		Srv:     srv,
		Msize:   MSIZE,
		Dotu:    true,
		reqs:    make(map[uint16]*SrvReq),
		fidpool: make(map[uint32]*SrvFid),
		reqout:  make(chan *SrvReq, 1),
	}
	req := &SrvReq{
		Tc:   &Fcall{Type: msgType, Tag: 1},
		Rc:   NewFcall(256),
		Conn: conn,
	}
	conn.reqs[req.Tc.Tag] = req
	return req
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name    string
		err     interface{}
		wantNum uint32
	}{
		{
			name:    "error-type",
			err:     &Error{Err: "boom", Errornum: EINVAL},
			wantNum: EINVAL,
		},
		{
			name:    "std-error",
			err:     errors.New("bad"),
			wantNum: EIO,
		},
		{
			name:    "string",
			err:     "plain",
			wantNum: EIO,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newTestReq(Tversion)
			req.RespondError(tt.err)
			if req.Rc.Type != Rerror {
				t.Fatalf("RespondError type = %d", req.Rc.Type)
			}
			if req.Rc.Errornum != tt.wantNum {
				t.Fatalf("RespondError num = %d, want %d", req.Rc.Errornum, tt.wantNum)
			}
		})
	}
}

func TestRespondMessages(t *testing.T) {
	tests := []struct {
		name     string
		call     func(*SrvReq)
		wantType uint8
	}{
		{
			name: "version",
			call: func(req *SrvReq) {
				req.RespondRversion(8192, "9P2000")
			},
			wantType: Rversion,
		},
		{
			name: "auth",
			call: func(req *SrvReq) {
				req.RespondRauth(&Qid{Type: QTAUTH, Path: 1, Version: 2})
			},
			wantType: Rauth,
		},
		{
			name: "flush",
			call: func(req *SrvReq) {
				req.RespondRflush()
			},
			wantType: Rflush,
		},
		{
			name: "attach",
			call: func(req *SrvReq) {
				req.RespondRattach(&Qid{Type: QTDIR, Path: 3, Version: 4})
			},
			wantType: Rattach,
		},
		{
			name: "walk",
			call: func(req *SrvReq) {
				req.RespondRwalk([]Qid{{Type: QTDIR, Path: 5, Version: 6}})
			},
			wantType: Rwalk,
		},
		{
			name: "open",
			call: func(req *SrvReq) {
				req.RespondRopen(&Qid{Type: QTDIR, Path: 7, Version: 8}, 0)
			},
			wantType: Ropen,
		},
		{
			name: "create",
			call: func(req *SrvReq) {
				req.RespondRcreate(&Qid{Type: QTDIR, Path: 9, Version: 10}, 0)
			},
			wantType: Rcreate,
		},
		{
			name: "read",
			call: func(req *SrvReq) {
				req.RespondRread([]byte("data"))
			},
			wantType: Rread,
		},
		{
			name: "write",
			call: func(req *SrvReq) {
				req.RespondRwrite(4)
			},
			wantType: Rwrite,
		},
		{
			name: "clunk",
			call: func(req *SrvReq) {
				req.RespondRclunk()
			},
			wantType: Rclunk,
		},
		{
			name: "remove",
			call: func(req *SrvReq) {
				req.RespondRremove()
			},
			wantType: Rremove,
		},
		{
			name: "stat",
			call: func(req *SrvReq) {
				req.RespondRstat(&Dir{Name: "file"})
			},
			wantType: Rstat,
		},
		{
			name: "wstat",
			call: func(req *SrvReq) {
				req.RespondRwstat()
			},
			wantType: Rwstat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newTestReq(Tversion)
			tt.call(req)
			if req.Rc.Type != tt.wantType {
				t.Fatalf("%s type = %d, want %d", tt.name, req.Rc.Type, tt.wantType)
			}
		})
	}
}
