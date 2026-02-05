package go9p

import (
	"strings"
	"testing"
)

type testSrvOps struct {
	authInitCalled    bool
	authReadCalled    bool
	authWriteCalled   bool
	authDestroyCalled bool
	createCalled      bool
	clunkCalled       bool
	removeCalled      bool
	statCalled        bool
	wstatCalled       bool
}

func (ops *testSrvOps) AuthInit(afid *SrvFid, aname string) (*Qid, error) {
	ops.authInitCalled = true
	return &Qid{Type: QTAUTH, Path: 1}, nil
}

func (ops *testSrvOps) AuthDestroy(afid *SrvFid) {
	ops.authDestroyCalled = true
}

func (ops *testSrvOps) AuthCheck(fid *SrvFid, afid *SrvFid, aname string) error {
	return nil
}

func (ops *testSrvOps) AuthRead(afid *SrvFid, offset uint64, data []byte) (int, error) {
	ops.authReadCalled = true
	return 0, nil
}

func (ops *testSrvOps) AuthWrite(afid *SrvFid, offset uint64, data []byte) (int, error) {
	ops.authWriteCalled = true
	return 0, nil
}

func (ops *testSrvOps) Attach(req *SrvReq) {
	req.RespondRattach(&Qid{Type: QTDIR, Path: 1})
}

func (ops *testSrvOps) Walk(req *SrvReq) {
	req.RespondRwalk(nil)
}

func (ops *testSrvOps) Open(req *SrvReq) {
	req.RespondRopen(&Qid{Type: QTFILE, Path: 2}, 0)
}

func (ops *testSrvOps) Create(req *SrvReq) {
	ops.createCalled = true
	req.RespondRcreate(&Qid{Type: QTFILE, Path: 3}, 0)
}

func (ops *testSrvOps) Read(req *SrvReq) {
	req.RespondRread([]byte("data"))
}

func (ops *testSrvOps) Write(req *SrvReq) {
	req.RespondRwrite(req.Tc.Count)
}

func (ops *testSrvOps) Clunk(req *SrvReq) {
	ops.clunkCalled = true
	req.RespondRclunk()
}

func (ops *testSrvOps) Remove(req *SrvReq) {
	ops.removeCalled = true
	req.RespondRremove()
}

func (ops *testSrvOps) Stat(req *SrvReq) {
	ops.statCalled = true
	req.RespondRstat(&Dir{Name: "file"})
}

func (ops *testSrvOps) Wstat(req *SrvReq) {
	ops.wstatCalled = true
	req.RespondRwstat()
}

func newSrvReq(msgType uint8, ops *testSrvOps) *SrvReq {
	srv := &Srv{Dotu: true, Msize: MSIZE, Upool: OsUsers, Maxpend: 1}
	srv.Start(ops)
	conn := &Conn{
		Srv:     srv,
		Msize:   MSIZE,
		Dotu:    true,
		fidpool: make(map[uint32]*SrvFid),
		reqs:    make(map[uint16]*SrvReq),
		reqout:  make(chan *SrvReq, 1),
	}
	req := &SrvReq{
		Tc:   &Fcall{Type: msgType, Tag: 1},
		Rc:   NewFcall(4096),
		Conn: conn,
	}
	conn.reqs[req.Tc.Tag] = req
	return req
}

func TestSrvAuth(t *testing.T) {
	ops := &testSrvOps{}
	req := newSrvReq(Tauth, ops)
	req.Tc.Afid = 1
	req.Tc.Unamenum = 2

	req.Conn.Srv.auth(req)
	if req.Rc.Type != Rauth {
		t.Fatalf("auth type = %d", req.Rc.Type)
	}
	if !ops.authInitCalled {
		t.Fatalf("AuthInit not called")
	}
	afid := req.Conn.fidpool[req.Tc.Afid]
	if afid == nil {
		t.Fatalf("afid missing")
	}
	if afid.Type&QTAUTH == 0 {
		t.Fatalf("afid type = %d", afid.Type)
	}
	if afid.refcount != 1 {
		t.Fatalf("afid refcount = %d", afid.refcount)
	}
}

func TestSrvFlushNoRequest(t *testing.T) {
	ops := &testSrvOps{}
	req := newSrvReq(Tflush, ops)
	req.Tc.Oldtag = 99

	req.Conn.Srv.flush(req)
	if req.Rc.Type != Rflush {
		t.Fatalf("flush type = %d", req.Rc.Type)
	}
}

func TestSrvCreate(t *testing.T) {
	ops := &testSrvOps{}
	req := newSrvReq(Tcreate, ops)
	user := OsUsers.Uid2User(1)
	req.Fid = &SrvFid{Fconn: req.Conn, User: user, Type: QTDIR}
	req.Tc.Name = "file"
	req.Tc.Perm = 0644
	req.Tc.Mode = OREAD

	req.Conn.Srv.create(req)
	if req.Rc.Type != Rcreate {
		t.Fatalf("create type = %d", req.Rc.Type)
	}
	if !ops.createCalled {
		t.Fatalf("Create not called")
	}
}

func TestSrvClunkRemoveStatWstat(t *testing.T) {
	tests := []struct {
		name     string
		msgType  uint8
		call     func(*Srv, *SrvReq)
		wantType uint8
		called   func(*testSrvOps) bool
	}{
		{
			name:     "clunk",
			msgType:  Tclunk,
			call:     func(srv *Srv, req *SrvReq) { srv.clunk(req) },
			wantType: Rclunk,
			called:   func(ops *testSrvOps) bool { return ops.clunkCalled },
		},
		{
			name:     "remove",
			msgType:  Tremove,
			call:     func(srv *Srv, req *SrvReq) { srv.remove(req) },
			wantType: Rremove,
			called:   func(ops *testSrvOps) bool { return ops.removeCalled },
		},
		{
			name:     "stat",
			msgType:  Tstat,
			call:     func(srv *Srv, req *SrvReq) { srv.stat(req) },
			wantType: Rstat,
			called:   func(ops *testSrvOps) bool { return ops.statCalled },
		},
		{
			name:     "wstat",
			msgType:  Twstat,
			call:     func(srv *Srv, req *SrvReq) { srv.wstat(req) },
			wantType: Rwstat,
			called:   func(ops *testSrvOps) bool { return ops.wstatCalled },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := &testSrvOps{}
			req := newSrvReq(tt.msgType, ops)
			user := OsUsers.Uid2User(1)
			req.Fid = &SrvFid{Fconn: req.Conn, User: user, Type: QTFILE}

			req.Conn.Srv.ops = ops
			tt.call(req.Conn.Srv, req)
			if req.Rc.Type != tt.wantType {
				t.Fatalf("%s type = %d", tt.name, req.Rc.Type)
			}
			if !tt.called(ops) {
				t.Fatalf("%s op not called", tt.name)
			}
		})
	}
}

func TestSrvReadPaths(t *testing.T) {
	user := OsUsers.Uid2User(1)
	tests := []struct {
		name       string
		setup      func(t *testing.T, req *SrvReq, fid *SrvFid, ops *testSrvOps)
		wantType   uint8
		wantErr    string
		wantAuth   bool
		wantOffset uint64
	}{
		{
			name: "auth-read",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid, ops *testSrvOps) {
				req.Fid = fid
				fid.Type = QTAUTH
				req.Tc.Count = 4
			},
			wantType: Rread,
			wantAuth: true,
		},
		{
			name: "too-large",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid, ops *testSrvOps) {
				req.Fid = fid
				fid.Type = QTFILE
				req.Tc.Count = req.Conn.Msize
			},
			wantType: Rerror,
			wantErr:  "i/o count too large",
		},
		{
			name: "dir-read",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid, ops *testSrvOps) {
				req.Fid = fid
				fid.Type = QTDIR
				req.Tc.Count = 4
				req.Tc.Offset = 0
			},
			wantType:   Rread,
			wantOffset: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := &testSrvOps{}
			req := newSrvReq(Tread, ops)
			fid := &SrvFid{Fconn: req.Conn, User: user}
			tt.setup(t, req, fid, ops)

			req.Conn.Srv.read(req)
			if req.Rc.Type != tt.wantType {
				t.Fatalf("%s type = %d", tt.name, req.Rc.Type)
			}
			if tt.wantErr != "" && !strings.Contains(req.Rc.Error, tt.wantErr) {
				t.Fatalf("%s error = %q", tt.name, req.Rc.Error)
			}
			if tt.wantAuth && !ops.authReadCalled {
				t.Fatalf("%s auth read not called", tt.name)
			}
			if tt.wantOffset != 0 && fid.Diroffset != tt.wantOffset {
				t.Fatalf("%s offset = %d", tt.name, fid.Diroffset)
			}
		})
	}
}

func TestSrvWritePaths(t *testing.T) {
	user := OsUsers.Uid2User(1)
	tests := []struct {
		name     string
		setup    func(t *testing.T, req *SrvReq, fid *SrvFid, ops *testSrvOps)
		wantType uint8
		wantErr  string
		wantAuth bool
	}{
		{
			name: "auth-write",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid, ops *testSrvOps) {
				req.Fid = fid
				fid.Type = QTAUTH
				req.Tc.Data = []byte("data")
				req.Tc.Count = uint32(len(req.Tc.Data))
			},
			wantType: Rwrite,
			wantAuth: true,
		},
		{
			name: "bad-use",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid, ops *testSrvOps) {
				req.Fid = fid
				fid.Type = QTFILE
				fid.opened = false
				req.Tc.Count = 1
			},
			wantType: Rerror,
			wantErr:  "bad use of fid",
		},
		{
			name: "too-large",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid, ops *testSrvOps) {
				req.Fid = fid
				fid.Type = QTFILE
				fid.opened = true
				fid.Omode = OWRITE
				req.Tc.Count = req.Conn.Msize
			},
			wantType: Rerror,
			wantErr:  "i/o count too large",
		},
		{
			name: "normal",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid, ops *testSrvOps) {
				req.Fid = fid
				fid.Type = QTFILE
				fid.opened = true
				fid.Omode = OWRITE
				req.Tc.Data = []byte("data")
				req.Tc.Count = uint32(len(req.Tc.Data))
			},
			wantType: Rwrite,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := &testSrvOps{}
			req := newSrvReq(Twrite, ops)
			fid := &SrvFid{Fconn: req.Conn, User: user}
			tt.setup(t, req, fid, ops)

			req.Conn.Srv.write(req)
			if req.Rc.Type != tt.wantType {
				t.Fatalf("%s type = %d", tt.name, req.Rc.Type)
			}
			if tt.wantErr != "" && !strings.Contains(req.Rc.Error, tt.wantErr) {
				t.Fatalf("%s error = %q", tt.name, req.Rc.Error)
			}
			if tt.wantAuth && !ops.authWriteCalled {
				t.Fatalf("%s auth write not called", tt.name)
			}
		})
	}
}

func TestSrvAttachErrors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, req *SrvReq)
		wantErr string
	}{
		{
			name: "no-fid",
			setup: func(t *testing.T, req *SrvReq) {
				req.Tc.Fid = NOFID
			},
			wantErr: "unknown fid",
		},
		{
			name: "no-user",
			setup: func(t *testing.T, req *SrvReq) {
				req.Tc.Fid = 1
				req.Conn.Dotu = false
				req.Tc.Unamenum = NOUID
			},
			wantErr: "unknown user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := &testSrvOps{}
			req := newSrvReq(Tattach, ops)
			tt.setup(t, req)

			req.Conn.Srv.attach(req)
			if req.Rc.Type != Rerror {
				t.Fatalf("%s type = %d", tt.name, req.Rc.Type)
			}
			if !strings.Contains(req.Rc.Error, tt.wantErr) {
				t.Fatalf("%s error = %q", tt.name, req.Rc.Error)
			}
		})
	}
}

func TestSrvWalkErrors(t *testing.T) {
	user := OsUsers.Uid2User(1)
	tests := []struct {
		name    string
		setup   func(t *testing.T, req *SrvReq, fid *SrvFid)
		wantErr string
	}{
		{
			name: "not-dir",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid) {
				fid.Type = QTFILE
				req.Tc.Wname = []string{"child"}
				req.Tc.Fid = 1
				req.Tc.Newfid = 1
			},
			wantErr: "not a directory",
		},
		{
			name: "opened",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid) {
				fid.Type = QTDIR
				fid.opened = true
				req.Tc.Wname = nil
				req.Tc.Fid = 1
				req.Tc.Newfid = 1
			},
			wantErr: "bad use of fid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := &testSrvOps{}
			req := newSrvReq(Twalk, ops)
			fid := &SrvFid{Fconn: req.Conn, User: user}
			req.Fid = fid
			tt.setup(t, req, fid)

			req.Conn.Srv.walk(req)
			if req.Rc.Type != Rerror {
				t.Fatalf("%s type = %d", tt.name, req.Rc.Type)
			}
			if !strings.Contains(req.Rc.Error, tt.wantErr) {
				t.Fatalf("%s error = %q", tt.name, req.Rc.Error)
			}
		})
	}
}

func TestSrvOpenErrors(t *testing.T) {
	user := OsUsers.Uid2User(1)
	tests := []struct {
		name    string
		setup   func(t *testing.T, req *SrvReq, fid *SrvFid)
		wantErr string
	}{
		{
			name: "already-open",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid) {
				fid.opened = true
				req.Tc.Mode = OREAD
			},
			wantErr: "fid already opened",
		},
		{
			name: "dir-write",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid) {
				fid.Type = QTDIR
				req.Tc.Mode = OWRITE
			},
			wantErr: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := &testSrvOps{}
			req := newSrvReq(Topen, ops)
			fid := &SrvFid{Fconn: req.Conn, User: user}
			req.Fid = fid
			tt.setup(t, req, fid)

			req.Conn.Srv.open(req)
			if req.Rc.Type != Rerror {
				t.Fatalf("%s type = %d", tt.name, req.Rc.Type)
			}
			if !strings.Contains(req.Rc.Error, tt.wantErr) {
				t.Fatalf("%s error = %q", tt.name, req.Rc.Error)
			}
		})
	}
}

func TestSrvCreateErrors(t *testing.T) {
	user := OsUsers.Uid2User(1)
	tests := []struct {
		name    string
		setup   func(t *testing.T, req *SrvReq, fid *SrvFid)
		wantErr string
	}{
		{
			name: "opened",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid) {
				fid.Type = QTDIR
				fid.opened = true
				req.Tc.Perm = 0644
			},
			wantErr: "fid already opened",
		},
		{
			name: "not-dir",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid) {
				fid.Type = QTFILE
				req.Tc.Perm = 0644
			},
			wantErr: "not a directory",
		},
		{
			name: "dir-mode",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid) {
				fid.Type = QTDIR
				req.Tc.Perm = DMDIR
				req.Tc.Mode = OWRITE
			},
			wantErr: "permission denied",
		},
		{
			name: "special-no-dotu",
			setup: func(t *testing.T, req *SrvReq, fid *SrvFid) {
				fid.Type = QTDIR
				req.Conn.Dotu = false
				req.Tc.Perm = DMSYMLINK
				req.Tc.Mode = OREAD
			},
			wantErr: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := &testSrvOps{}
			req := newSrvReq(Tcreate, ops)
			fid := &SrvFid{Fconn: req.Conn, User: user}
			req.Fid = fid
			tt.setup(t, req, fid)

			req.Conn.Srv.create(req)
			if req.Rc.Type != Rerror {
				t.Fatalf("%s type = %d", tt.name, req.Rc.Type)
			}
			if !strings.Contains(req.Rc.Error, tt.wantErr) {
				t.Fatalf("%s error = %q", tt.name, req.Rc.Error)
			}
		})
	}
}

func TestSrvClunkAuth(t *testing.T) {
	ops := &testSrvOps{}
	req := newSrvReq(Tclunk, ops)
	fid := &SrvFid{Fconn: req.Conn, User: OsUsers.Uid2User(1), Type: QTAUTH}
	req.Fid = fid

	req.Conn.Srv.clunk(req)
	if req.Rc.Type != Rclunk {
		t.Fatalf("clunk type = %d", req.Rc.Type)
	}
	if !ops.authDestroyCalled {
		t.Fatalf("AuthDestroy not called")
	}
}
