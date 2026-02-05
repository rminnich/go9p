package go9p

import (
	"strings"
	"testing"
	"time"
)

type stubUser struct {
	name string
	id   int
}

func (u stubUser) Name() string          { return u.name }
func (u stubUser) Id() int               { return u.id }
func (u stubUser) Groups() []Group       { return nil }
func (u stubUser) IsMember(g Group) bool { return false }

func newTestTag(t *testing.T) (*Clnt, *Tag, chan *Req) {
	t.Helper()
	clnt := &Clnt{
		Msize:   256,
		reqout:  make(chan *Req, 16),
		tagpool: NewPool(0, uint32(NOTAG)),
	}
	reqchan := make(chan *Req, 16)
	tag := clnt.TagAlloc(reqchan)
	t.Cleanup(func() {
		clnt.TagFree(tag)
	})
	return clnt, tag, reqchan
}

func waitReq(t *testing.T, ch <-chan *Req) *Req {
	t.Helper()
	select {
	case req := <-ch:
		return req
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for request")
		return nil
	}
}

func TestTagMethods(t *testing.T) {
	clnt, tag, _ := newTestTag(t)
	user := stubUser{name: "user", id: 1}

	tests := []struct {
		name     string
		call     func(*Tag) (*Fid, *Fid, error)
		wantType uint8
		check    func(t *testing.T, fid *Fid, newfid *Fid, req *Req)
	}{
		{
			name: "auth",
			call: func(tag *Tag) (*Fid, *Fid, error) {
				afid := &Fid{Fid: 1, Clnt: clnt}
				err := tag.Auth(afid, user, "srv")
				return afid, nil, err
			},
			wantType: Tauth,
			check: func(t *testing.T, fid *Fid, newfid *Fid, req *Req) {
				if fid.User == nil || fid.User.Name() != user.Name() {
					t.Fatalf("Auth user = %#v", fid.User)
				}
			},
		},
		{
			name: "attach",
			call: func(tag *Tag) (*Fid, *Fid, error) {
				fid := &Fid{Fid: 2, Clnt: clnt}
				err := tag.Attach(fid, nil, user, "srv")
				return fid, nil, err
			},
			wantType: Tattach,
			check: func(t *testing.T, fid *Fid, newfid *Fid, req *Req) {
				if fid.User == nil || fid.User.Name() != user.Name() {
					t.Fatalf("Attach user = %#v", fid.User)
				}
			},
		},
		{
			name: "walk-no-names",
			call: func(tag *Tag) (*Fid, *Fid, error) {
				fid := &Fid{
					Fid:  3,
					Clnt: clnt,
					User: user,
					Qid:  Qid{Type: QTDIR, Path: 1, Version: 2},
				}
				newfid := &Fid{Fid: 4, Clnt: clnt}
				err := tag.Walk(fid, newfid, nil)
				return fid, newfid, err
			},
			wantType: Twalk,
			check: func(t *testing.T, fid *Fid, newfid *Fid, req *Req) {
				if newfid.Qid != fid.Qid {
					t.Fatalf("Walk qid = %+v, want %+v", newfid.Qid, fid.Qid)
				}
				if newfid.User != fid.User {
					t.Fatalf("Walk user = %#v", newfid.User)
				}
			},
		},
		{
			name: "open",
			call: func(tag *Tag) (*Fid, *Fid, error) {
				fid := &Fid{Fid: 5, Clnt: clnt}
				err := tag.Open(fid, OWRITE)
				return fid, nil, err
			},
			wantType: Topen,
			check: func(t *testing.T, fid *Fid, newfid *Fid, req *Req) {
				if fid.Mode != OWRITE {
					t.Fatalf("Open mode = %d", fid.Mode)
				}
			},
		},
		{
			name: "create",
			call: func(tag *Tag) (*Fid, *Fid, error) {
				fid := &Fid{Fid: 6, Clnt: clnt}
				err := tag.Create(fid, "file", 0644, OREAD, "")
				return fid, nil, err
			},
			wantType: Tcreate,
			check: func(t *testing.T, fid *Fid, newfid *Fid, req *Req) {
				if fid.Mode != OREAD {
					t.Fatalf("Create mode = %d", fid.Mode)
				}
			},
		},
		{
			name: "read",
			call: func(tag *Tag) (*Fid, *Fid, error) {
				fid := &Fid{Fid: 7, Clnt: clnt}
				err := tag.Read(fid, 0, 16)
				return fid, nil, err
			},
			wantType: Tread,
		},
		{
			name: "write",
			call: func(tag *Tag) (*Fid, *Fid, error) {
				fid := &Fid{Fid: 8, Clnt: clnt}
				err := tag.Write(fid, []byte("data"), 0)
				return fid, nil, err
			},
			wantType: Twrite,
			check: func(t *testing.T, fid *Fid, newfid *Fid, req *Req) {
				if req.Tc.Count != 4 {
					t.Fatalf("Write count = %d", req.Tc.Count)
				}
			},
		},
		{
			name: "clunk",
			call: func(tag *Tag) (*Fid, *Fid, error) {
				fid := &Fid{Fid: 9, Clnt: clnt}
				err := tag.Clunk(fid)
				return fid, nil, err
			},
			wantType: Tclunk,
		},
		{
			name: "remove",
			call: func(tag *Tag) (*Fid, *Fid, error) {
				fid := &Fid{Fid: 10, Clnt: clnt}
				err := tag.Remove(fid)
				return fid, nil, err
			},
			wantType: Tremove,
		},
		{
			name: "stat",
			call: func(tag *Tag) (*Fid, *Fid, error) {
				fid := &Fid{Fid: 11, Clnt: clnt}
				err := tag.Stat(fid)
				return fid, nil, err
			},
			wantType: Tstat,
		},
		{
			name: "wstat",
			call: func(tag *Tag) (*Fid, *Fid, error) {
				fid := &Fid{Fid: 12, Clnt: clnt}
				dir := &Dir{Name: "file"}
				err := tag.Wstat(fid, dir)
				return fid, nil, err
			},
			wantType: Twstat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fid, newfid, err := tt.call(tag)
			if err != nil {
				t.Fatalf("%s error = %v", tt.name, err)
			}
			req := waitReq(t, clnt.reqout)
			if req.Tc.Type != tt.wantType {
				t.Fatalf("%s type = %d, want %d", tt.name,
					req.Tc.Type, tt.wantType)
			}
			if tt.check != nil {
				tt.check(t, fid, newfid, req)
			}
			tag.ReqFree(req)
		})
	}
}

func TestTagReqprocUpdates(t *testing.T) {
	clnt, tag, reqchan := newTestTag(t)
	user := stubUser{name: "user", id: 1}

	tests := []struct {
		name        string
		tcType      uint8
		rc          *Fcall
		startUser   User
		startMode   uint8
		wantQid     Qid
		wantIounit  uint32
		wantUserNil bool
		wantMode    uint8
		wantWalked  bool
	}{
		{
			name:      "attach-success",
			tcType:    Tattach,
			rc:        &Fcall{Type: Rattach, Qid: Qid{Type: QTDIR, Path: 1}},
			startUser: user,
			wantQid:   Qid{Type: QTDIR, Path: 1},
		},
		{
			name:        "walk-error",
			tcType:      Twalk,
			rc:          &Fcall{Type: Rerror, Error: "missing"},
			startUser:   user,
			wantUserNil: true,
		},
		{
			name:       "create-success",
			tcType:     Tcreate,
			rc:         &Fcall{Type: Rcreate, Qid: Qid{Type: QTFILE, Path: 2}, Iounit: 99},
			startUser:  user,
			startMode:  OWRITE,
			wantQid:    Qid{Type: QTFILE, Path: 2},
			wantMode:   OWRITE,
			wantIounit: 99,
		},
		{
			name:        "create-error",
			tcType:      Tcreate,
			rc:          &Fcall{Type: Rerror, Error: "fail"},
			startUser:   user,
			startMode:   OWRITE,
			wantUserNil: false,
			wantMode:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fid := &Fid{
				Fid:  1,
				Clnt: clnt,
				User: tt.startUser,
				Mode: tt.startMode,
			}
			req := &Req{Tc: &Fcall{Type: tt.tcType}, Rc: tt.rc, fid: fid}
			tag.respchan <- req
			resp := waitReq(t, reqchan)
			if resp != req {
				t.Fatalf("unexpected req returned")
			}
			if tt.wantUserNil && fid.User != nil {
				t.Fatalf("user not cleared")
			}
			if !tt.wantUserNil && fid.User == nil {
				t.Fatalf("user cleared unexpectedly")
			}
			if tt.wantQid != (Qid{}) && fid.Qid != tt.wantQid {
				t.Fatalf("qid = %+v, want %+v", fid.Qid, tt.wantQid)
			}
			if tt.wantIounit != 0 && fid.Iounit != tt.wantIounit {
				t.Fatalf("iounit = %d, want %d", fid.Iounit, tt.wantIounit)
			}
			if fid.Mode != tt.wantMode {
				t.Fatalf("mode = %d, want %d", fid.Mode, tt.wantMode)
			}
		})
	}
}

func TestTagReqAllocFree(t *testing.T) {
	_, tag, _ := newTestTag(t)
	req := tag.reqAlloc()
	if req.Tc == nil {
		t.Fatalf("reqAlloc tc is nil")
	}
	tag.ReqFree(req)
}

func TestTagReqprocErrorMessage(t *testing.T) {
	clnt, tag, reqchan := newTestTag(t)
	user := stubUser{name: "user", id: 1}
	fid := &Fid{Fid: 1, Clnt: clnt, User: user}
	req := &Req{
		Tc:  &Fcall{Type: Tattach},
		Rc:  &Fcall{Type: Rerror, Error: "boom"},
		fid: fid,
	}
	tag.respchan <- req
	resp := waitReq(t, reqchan)
	if resp.Rc.Type != Rerror {
		t.Fatalf("rc type = %d", resp.Rc.Type)
	}
	if !strings.Contains(resp.Rc.Error, "boom") {
		t.Fatalf("error = %q", resp.Rc.Error)
	}
}
