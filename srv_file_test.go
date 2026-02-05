package go9p

import (
	"errors"
	"strings"
	"testing"
)

type testGroup struct {
	name string
	id   int
}

func (g testGroup) Name() string { return g.name }
func (g testGroup) Id() int      { return g.id }
func (g testGroup) Members() []User {
	return nil
}

type testUser struct {
	name   string
	id     int
	groups []Group
}

func (u testUser) Name() string    { return u.name }
func (u testUser) Id() int         { return u.id }
func (u testUser) Groups() []Group { return u.groups }
func (u testUser) IsMember(g Group) bool {
	return false
}

type testFileOps struct {
	openCalled    bool
	createCalled  bool
	readCalled    bool
	writeCalled   bool
	removeCalled  bool
	statCalled    bool
	wstatCalled   bool
	clunkCalled   bool
	destroyCalled bool

	openErr   error
	createErr error
	readErr   error
	writeErr  error
	removeErr error
	statErr   error
	wstatErr  error
	clunkErr  error

	readData     []byte
	writeData    []byte
	bytesWritten int
	createFile   *srvFile
}

func (ops *testFileOps) Open(fid *FFid, mode uint8) error {
	ops.openCalled = true
	return ops.openErr
}

func (ops *testFileOps) Create(fid *FFid, name string, perm uint32) (*srvFile, error) {
	ops.createCalled = true
	return ops.createFile, ops.createErr
}

func (ops *testFileOps) Read(fid *FFid, buf []byte, offset uint64) (int, error) {
	ops.readCalled = true
	if ops.readErr != nil {
		return 0, ops.readErr
	}
	return copy(buf, ops.readData), nil
}

func (ops *testFileOps) Write(fid *FFid, data []byte, offset uint64) (int, error) {
	ops.writeCalled = true
	ops.writeData = append([]byte(nil), data...)
	if ops.writeErr != nil {
		return 0, ops.writeErr
	}
	if ops.bytesWritten != 0 {
		return ops.bytesWritten, nil
	}
	return len(data), nil
}

func (ops *testFileOps) Remove(fid *FFid) error {
	ops.removeCalled = true
	return ops.removeErr
}

func (ops *testFileOps) Stat(fid *FFid) error {
	ops.statCalled = true
	return ops.statErr
}

func (ops *testFileOps) Wstat(fid *FFid, dir *Dir) error {
	ops.wstatCalled = true
	return ops.wstatErr
}

func (ops *testFileOps) Clunk(fid *FFid) error {
	ops.clunkCalled = true
	return ops.clunkErr
}

func (ops *testFileOps) FidDestroy(fid *FFid) {
	ops.destroyCalled = true
}

func newFsrvReq(msgType uint8) *SrvReq {
	req := newTestReq(msgType)
	req.Rc = NewFcall(4096)
	return req
}

func TestSrvFileAddAndFind(t *testing.T) {
	owner := testUser{name: "owner", id: 1}
	group := testGroup{name: "group", id: 2}
	root := &srvFile{}
	if err := root.Add(nil, "root", owner, group, DMDIR|0755, nil); err != nil {
		t.Fatalf("Add root error = %v", err)
	}

	child := &srvFile{}
	if err := child.Add(root, "child", owner, group, 0644, nil); err != nil {
		t.Fatalf("Add child error = %v", err)
	}
	if got := root.Find("child"); got != child {
		t.Fatalf("Find(child) = %+v, want %+v", got, child)
	}

	dup := &srvFile{}
	if err := dup.Add(root, "child", owner, group, 0644, nil); err != Eexist {
		t.Fatalf("Add duplicate error = %v, want %v", err, Eexist)
	}

	noneUser := &srvFile{}
	if err := noneUser.Add(root, "anon", nil, nil, 0644, nil); err != nil {
		t.Fatalf("Add anon error = %v", err)
	}
	if noneUser.Uid != "none" || noneUser.Gid != "none" {
		t.Fatalf("Add anon user gid = %q %q", noneUser.Uid, noneUser.Gid)
	}
}

func TestSrvFileRemoveAndRename(t *testing.T) {
	owner := testUser{name: "owner", id: 1}
	group := testGroup{name: "group", id: 2}
	root := &srvFile{}
	if err := root.Add(nil, "root", owner, group, DMDIR|0755, nil); err != nil {
		t.Fatalf("Add root error = %v", err)
	}

	child := &srvFile{}
	if err := child.Add(root, "child", owner, group, 0644, nil); err != nil {
		t.Fatalf("Add child error = %v", err)
	}
	other := &srvFile{}
	if err := other.Add(root, "other", owner, group, 0644, nil); err != nil {
		t.Fatalf("Add other error = %v", err)
	}

	if err := child.Rename("other"); err != Eexist {
		t.Fatalf("Rename error = %v, want %v", err, Eexist)
	}
	if err := child.Rename("renamed"); err != nil {
		t.Fatalf("Rename error = %v", err)
	}
	if child.Name != "renamed" {
		t.Fatalf("Rename name = %q", child.Name)
	}

	child.Remove()
	if got := root.Find("renamed"); got != nil {
		t.Fatalf("Find after remove = %+v", got)
	}
}

func TestSrvFileCheckPerm(t *testing.T) {
	ownerGroup := testGroup{name: "staff", id: 2}
	owner := testUser{name: "owner", id: 1, groups: []Group{ownerGroup}}
	member := testUser{name: "member", id: 3, groups: []Group{ownerGroup}}
	other := testUser{name: "other", id: 4}

	file := &srvFile{
		Dir: Dir{
			Mode:   0644,
			Uid:    owner.name,
			Uidnum: uint32(owner.id),
			Gid:    ownerGroup.name,
			Gidnum: uint32(ownerGroup.id),
		},
	}

	tests := []struct {
		name string
		user User
		perm uint32
		want bool
	}{
		{
			name: "owner-write",
			user: owner,
			perm: DMWRITE,
			want: true,
		},
		{
			name: "group-read",
			user: member,
			perm: DMREAD,
			want: true,
		},
		{
			name: "other-read",
			user: other,
			perm: DMREAD,
			want: true,
		},
		{
			name: "other-write",
			user: other,
			perm: DMWRITE,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := file.CheckPerm(tt.user, tt.perm)
			if got != tt.want {
				t.Fatalf("CheckPerm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMode2Perm(t *testing.T) {
	tests := []struct {
		name string
		mode uint8
		want uint32
	}{
		{
			name: "read",
			mode: OREAD,
			want: DMREAD,
		},
		{
			name: "write",
			mode: OWRITE,
			want: DMWRITE,
		},
		{
			name: "read-write",
			mode: ORDWR,
			want: DMREAD | DMWRITE,
		},
		{
			name: "truncate",
			mode: OREAD | OTRUNC,
			want: DMREAD | DMWRITE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mode2Perm(tt.mode)
			if got != tt.want {
				t.Fatalf("mode2Perm(%d) = %d, want %d", tt.mode, got, tt.want)
			}
		})
	}
}

func TestFsrvAttachAndWalk(t *testing.T) {
	owner := testUser{name: "owner", id: 1}
	group := testGroup{name: "group", id: 2}
	root := &srvFile{}
	if err := root.Add(nil, "root", owner, group, DMDIR|0755, nil); err != nil {
		t.Fatalf("Add root error = %v", err)
	}
	child := &srvFile{}
	if err := child.Add(root, "child", owner, group, DMDIR|0755, nil); err != nil {
		t.Fatalf("Add child error = %v", err)
	}

	srv := NewsrvFileSrv(root)
	attachReq := newFsrvReq(Tattach)
	fid := &SrvFid{Fconn: attachReq.Conn, User: owner}
	attachReq.Fid = fid
	srv.Attach(attachReq)
	if attachReq.Rc.Type != Rattach {
		t.Fatalf("Attach type = %d", attachReq.Rc.Type)
	}
	if fid.Aux == nil {
		t.Fatalf("Attach fid aux nil")
	}

	tests := []struct {
		name      string
		wname     []string
		wantErr   string
		wantCount int
	}{
		{
			name:      "walk-child",
			wname:     []string{"child"},
			wantCount: 1,
		},
		{
			name:    "walk-missing",
			wname:   []string{"missing"},
			wantErr: "file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			walkReq := newFsrvReq(Twalk)
			walkFid := &SrvFid{Fconn: walkReq.Conn, User: owner}
			walkReq.Fid = walkFid
			walkReq.Fid.Aux = &FFid{F: root, Fid: walkFid}
			newfid := &SrvFid{Fconn: walkReq.Conn}
			walkReq.Newfid = newfid
			walkReq.Tc.Wname = tt.wname

			srv.Walk(walkReq)
			if tt.wantErr != "" {
				if walkReq.Rc.Type != Rerror {
					t.Fatalf("Walk type = %d", walkReq.Rc.Type)
				}
				if !strings.Contains(walkReq.Rc.Error, tt.wantErr) {
					t.Fatalf("Walk error = %q", walkReq.Rc.Error)
				}
				return
			}
			if walkReq.Rc.Type != Rwalk {
				t.Fatalf("Walk type = %d", walkReq.Rc.Type)
			}
			if len(walkReq.Rc.Wqid) != tt.wantCount {
				t.Fatalf("Walk count = %d", len(walkReq.Rc.Wqid))
			}
			newAux := newfid.Aux.(*FFid)
			if newAux.F != child {
				t.Fatalf("Walk newfid = %+v", newAux.F)
			}
		})
	}
}

func TestFsrvOpen(t *testing.T) {
	owner := testUser{name: "owner", id: 1}

	tests := []struct {
		name    string
		user    User
		wantErr string
	}{
		{
			name: "allowed",
			user: owner,
		},
		{
			name:    "denied",
			user:    nil,
			wantErr: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := &testFileOps{}
			file := &srvFile{Dir: Dir{Mode: 0644, Uid: owner.name, Uidnum: uint32(owner.id)}}
			file.ops = ops

			req := newFsrvReq(Topen)
			req.Fid = &SrvFid{Fconn: req.Conn, User: tt.user}
			req.Fid.Aux = &FFid{F: file, Fid: req.Fid}
			req.Tc.Mode = OREAD

			(&Fsrv{}).Open(req)
			if tt.wantErr != "" {
				if req.Rc.Type != Rerror {
					t.Fatalf("Open type = %d", req.Rc.Type)
				}
				if !strings.Contains(req.Rc.Error, tt.wantErr) {
					t.Fatalf("Open error = %q", req.Rc.Error)
				}
				if ops.openCalled {
					t.Fatalf("Open called unexpectedly")
				}
				return
			}
			if req.Rc.Type != Ropen {
				t.Fatalf("Open type = %d", req.Rc.Type)
			}
			if !ops.openCalled {
				t.Fatalf("Open not called")
			}
		})
	}
}

func TestFsrvCreate(t *testing.T) {
	owner := testUser{name: "owner", id: 1}
	group := testGroup{name: "group", id: 2}

	tests := []struct {
		name    string
		user    User
		wantErr string
	}{
		{
			name: "allowed",
			user: owner,
		},
		{
			name:    "denied",
			user:    nil,
			wantErr: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := &srvFile{}
			if err := dir.Add(nil, "root", owner, group, DMDIR|0755, nil); err != nil {
				t.Fatalf("Add dir error = %v", err)
			}
			ops := &testFileOps{}
			var child *srvFile
			if tt.wantErr == "" {
				child = &srvFile{}
				if err := child.Add(dir, "new", owner, group, 0644, nil); err != nil {
					t.Fatalf("Add child error = %v", err)
				}
				ops.createFile = child
			}
			dir.ops = ops

			req := newFsrvReq(Tcreate)
			fid := &SrvFid{Fconn: req.Conn, User: tt.user}
			req.Fid = fid
			req.Fid.Aux = &FFid{F: dir, Fid: fid}
			req.Tc.Name = "new"
			req.Tc.Perm = 0644
			req.Tc.Mode = OREAD

			(&Fsrv{}).Create(req)
			if tt.wantErr != "" {
				if req.Rc.Type != Rerror {
					t.Fatalf("Create type = %d", req.Rc.Type)
				}
				if !strings.Contains(req.Rc.Error, tt.wantErr) {
					t.Fatalf("Create error = %q", req.Rc.Error)
				}
				if ops.createCalled {
					t.Fatalf("Create called unexpectedly")
				}
				return
			}
			if req.Rc.Type != Rcreate {
				t.Fatalf("Create type = %d", req.Rc.Type)
			}
			if !ops.createCalled {
				t.Fatalf("Create not called")
			}
			created := fid.Aux.(*FFid)
			if created.F != child {
				t.Fatalf("Create fid = %+v", created.F)
			}
		})
	}
}

func TestFsrvReadWrite(t *testing.T) {
	owner := testUser{name: "owner", id: 1}
	ops := &testFileOps{readData: []byte("data")}
	file := &srvFile{Dir: Dir{Mode: 0644, Uid: owner.name, Uidnum: uint32(owner.id)}}
	file.ops = ops

	readReq := newFsrvReq(Tread)
	readReq.Fid = &SrvFid{Fconn: readReq.Conn, User: owner}
	readReq.Fid.Aux = &FFid{F: file, Fid: readReq.Fid}
	readReq.Tc.Count = 4
	readReq.Tc.Offset = 0
	(&Fsrv{}).Read(readReq)
	if readReq.Rc.Type != Rread {
		t.Fatalf("Read type = %d", readReq.Rc.Type)
	}
	if string(readReq.Rc.Data[:readReq.Rc.Count]) != "data" {
		t.Fatalf("Read data = %q", string(readReq.Rc.Data[:readReq.Rc.Count]))
	}
	if !ops.readCalled {
		t.Fatalf("Read not called")
	}

	writeReq := newFsrvReq(Twrite)
	writeReq.Fid = &SrvFid{Fconn: writeReq.Conn, User: owner}
	writeReq.Fid.Aux = &FFid{F: file, Fid: writeReq.Fid}
	writeReq.Tc.Data = []byte("ping")
	writeReq.Tc.Count = uint32(len(writeReq.Tc.Data))
	writeReq.Tc.Offset = 0
	(&Fsrv{}).Write(writeReq)
	if writeReq.Rc.Type != Rwrite {
		t.Fatalf("Write type = %d", writeReq.Rc.Type)
	}
	if string(ops.writeData) != "ping" {
		t.Fatalf("Write data = %q", string(ops.writeData))
	}
	if !ops.writeCalled {
		t.Fatalf("Write not called")
	}
}

func TestFsrvRemove(t *testing.T) {
	owner := testUser{name: "owner", id: 1}
	group := testGroup{name: "group", id: 2}

	tests := []struct {
		name      string
		withChild bool
		wantErr   string
	}{
		{
			name:      "not-empty",
			withChild: true,
			wantErr:   "directory not empty",
		},
		{
			name: "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &srvFile{}
			if err := root.Add(nil, "root", owner, group, DMDIR|0755, nil); err != nil {
				t.Fatalf("Add root error = %v", err)
			}
			target := &srvFile{}
			ops := &testFileOps{}
			if err := target.Add(root, "target", owner, group, 0644, ops); err != nil {
				t.Fatalf("Add target error = %v", err)
			}
			if tt.withChild {
				child := &srvFile{}
				if err := child.Add(target, "child", owner, group, 0644, nil); err != nil {
					t.Fatalf("Add child error = %v", err)
				}
			}

			req := newFsrvReq(Tremove)
			req.Fid = &SrvFid{Fconn: req.Conn, User: owner}
			req.Fid.Aux = &FFid{F: target, Fid: req.Fid}
			(&Fsrv{}).Remove(req)

			if tt.wantErr != "" {
				if req.Rc.Type != Rerror {
					t.Fatalf("Remove type = %d", req.Rc.Type)
				}
				if !strings.Contains(req.Rc.Error, tt.wantErr) {
					t.Fatalf("Remove error = %q", req.Rc.Error)
				}
				return
			}
			if req.Rc.Type != Rremove {
				t.Fatalf("Remove type = %d", req.Rc.Type)
			}
			if root.Find("target") != nil {
				t.Fatalf("Remove did not detach")
			}
			if !ops.removeCalled {
				t.Fatalf("Remove op not called")
			}
		})
	}
}

func TestFsrvStatAndWstat(t *testing.T) {
	owner := testUser{name: "owner", id: 1}
	file := &srvFile{Dir: Dir{Mode: 0644, Uid: owner.name, Uidnum: uint32(owner.id)}}

	statTests := []struct {
		name    string
		statErr error
		wantErr string
	}{
		{
			name: "ok",
		},
		{
			name:    "error",
			statErr: errors.New("stat boom"),
			wantErr: "stat boom",
		},
	}

	for _, tt := range statTests {
		t.Run("stat-"+tt.name, func(t *testing.T) {
			ops := &testFileOps{statErr: tt.statErr}
			file.ops = ops
			req := newFsrvReq(Tstat)
			req.Fid = &SrvFid{Fconn: req.Conn, User: owner}
			req.Fid.Aux = &FFid{F: file, Fid: req.Fid}
			(&Fsrv{}).Stat(req)
			if tt.wantErr != "" {
				if req.Rc.Type != Rerror {
					t.Fatalf("Stat type = %d", req.Rc.Type)
				}
				if !strings.Contains(req.Rc.Error, tt.wantErr) {
					t.Fatalf("Stat error = %q", req.Rc.Error)
				}
				return
			}
			if req.Rc.Type != Rstat {
				t.Fatalf("Stat type = %d", req.Rc.Type)
			}
			if !ops.statCalled {
				t.Fatalf("Stat op not called")
			}
		})
	}

	wstatTests := []struct {
		name     string
		withOps  bool
		wstatErr error
		wantErr  string
	}{
		{
			name:    "ok",
			withOps: true,
		},
		{
			name:     "error",
			withOps:  true,
			wstatErr: errors.New("wstat boom"),
			wantErr:  "wstat boom",
		},
		{
			name:    "missing-ops",
			withOps: false,
			wantErr: "permission denied",
		},
	}

	for _, tt := range wstatTests {
		t.Run("wstat-"+tt.name, func(t *testing.T) {
			var ops *testFileOps
			if tt.withOps {
				ops = &testFileOps{wstatErr: tt.wstatErr}
				file.ops = ops
			} else {
				file.ops = nil
			}
			req := newFsrvReq(Twstat)
			req.Fid = &SrvFid{Fconn: req.Conn, User: owner}
			req.Fid.Aux = &FFid{F: file, Fid: req.Fid}
			req.Tc.Dir = Dir{Name: "new"}
			(&Fsrv{}).Wstat(req)
			if tt.wantErr != "" {
				if req.Rc.Type != Rerror {
					t.Fatalf("Wstat type = %d", req.Rc.Type)
				}
				if !strings.Contains(req.Rc.Error, tt.wantErr) {
					t.Fatalf("Wstat error = %q", req.Rc.Error)
				}
				return
			}
			if req.Rc.Type != Rwstat {
				t.Fatalf("Wstat type = %d", req.Rc.Type)
			}
			if ops != nil && !ops.wstatCalled {
				t.Fatalf("Wstat op not called")
			}
		})
	}
}

func TestFsrvClunkAndFidDestroy(t *testing.T) {
	owner := testUser{name: "owner", id: 1}
	ops := &testFileOps{}
	file := &srvFile{Dir: Dir{Mode: 0644, Uid: owner.name, Uidnum: uint32(owner.id)}}
	file.ops = ops

	clunkReq := newFsrvReq(Tclunk)
	clunkReq.Fid = &SrvFid{Fconn: clunkReq.Conn, User: owner}
	clunkReq.Fid.Aux = &FFid{F: file, Fid: clunkReq.Fid}
	(&Fsrv{}).Clunk(clunkReq)
	if clunkReq.Rc.Type != Rclunk {
		t.Fatalf("Clunk type = %d", clunkReq.Rc.Type)
	}
	if !ops.clunkCalled {
		t.Fatalf("Clunk op not called")
	}

	srvfid := &SrvFid{Fconn: clunkReq.Conn}
	srvfid.Aux = &FFid{F: file, Fid: srvfid}
	(&Fsrv{}).FidDestroy(srvfid)
	if !ops.destroyCalled {
		t.Fatalf("FidDestroy not called")
	}
}
