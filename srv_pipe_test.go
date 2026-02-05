//go:build unix && !tinygo

package go9p

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newPipeReq(msgType uint8) *SrvReq {
	req := newTestReq(msgType)
	req.Rc = NewFcall(4096)
	fid := &SrvFid{Fconn: req.Conn}
	req.Fid = fid
	return req
}

func TestPipefsConnLifecycle(t *testing.T) {
	pipe := &Pipefs{}
	conn := &Conn{Srv: &Srv{Debuglevel: 1}}
	pipe.ConnOpened(conn)
	pipe.ConnClosed(conn)
}

func TestPipefsFidDestroy(t *testing.T) {
	pipe := &Pipefs{}
	file, err := os.CreateTemp(t.TempDir(), "pipe")
	if err != nil {
		t.Fatalf("CreateTemp error = %v", err)
	}
	fid := &SrvFid{Aux: &pipeFid{file: file}}
	pipe.FidDestroy(fid)
	if _, err := file.WriteString("x"); err == nil {
		t.Fatalf("expected write error on closed file")
	}
}

func TestPipefsCreateStatRemove(t *testing.T) {
	root := t.TempDir()
	pipe := &Pipefs{Root: root}

	createReq := newPipeReq(Tcreate)
	createReq.Fid.Aux = &pipeFid{path: root}
	createReq.Tc.Name = "newfile"
	createReq.Tc.Perm = 0644
	createReq.Tc.Mode = OWRITE
	pipe.Create(createReq)
	if createReq.Rc.Type != Rcreate {
		t.Fatalf("Create type = %d", createReq.Rc.Type)
	}

	createdPath := filepath.Join(root, "newfile")
	if _, err := os.Stat(createdPath); err != nil {
		t.Fatalf("created file stat error = %v", err)
	}

	statReq := newPipeReq(Tstat)
	statReq.Fid.Aux = &pipeFid{path: createdPath}
	pipe.Stat(statReq)
	if statReq.Rc.Type != Rstat {
		t.Fatalf("Stat type = %d", statReq.Rc.Type)
	}

	removeReq := newPipeReq(Tremove)
	removeReq.Fid.Aux = &pipeFid{path: createdPath}
	pipe.Remove(removeReq)
	if removeReq.Rc.Type != Rremove {
		t.Fatalf("Remove type = %d", removeReq.Rc.Type)
	}
	if _, err := os.Stat(createdPath); !os.IsNotExist(err) {
		t.Fatalf("Remove stat error = %v", err)
	}
}

func TestPipefsWstatAndClunk(t *testing.T) {
	root := t.TempDir()
	pipe := &Pipefs{Root: root}

	wstatReq := newPipeReq(Twstat)
	wstatReq.Fid.Aux = &pipeFid{path: root}
	pipe.Wstat(wstatReq)
	if wstatReq.Rc.Type != Rerror {
		t.Fatalf("Wstat type = %d", wstatReq.Rc.Type)
	}
	if !strings.Contains(wstatReq.Rc.Error, "permission denied") {
		t.Fatalf("Wstat error = %q", wstatReq.Rc.Error)
	}

	clunkReq := newPipeReq(Tclunk)
	clunkReq.Fid.Aux = &pipeFid{path: root}
	pipe.Clunk(clunkReq)
	if clunkReq.Rc.Type != Rclunk {
		t.Fatalf("Clunk type = %d", clunkReq.Rc.Type)
	}
}

func TestPipefsFlush(t *testing.T) {
	pipe := &Pipefs{}
	pipe.Flush(newPipeReq(Tflush))
}

func TestPipefsReadDirAndFile(t *testing.T) {
	pipe := &Pipefs{}

	tests := []struct {
		name    string
		setup   func(t *testing.T, req *SrvReq) *pipeFid
		wantErr string
	}{
		{
			name: "dir",
			setup: func(t *testing.T, req *SrvReq) *pipeFid {
				root := t.TempDir()
				filePath := filepath.Join(root, "file")
				if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
					t.Fatalf("WriteFile error = %v", err)
				}
				file, err := os.Open(root)
				if err != nil {
					t.Fatalf("Open dir error = %v", err)
				}
				fid := &pipeFid{path: root, file: file}
				req.Tc.Count = 2048
				req.Tc.Offset = 0
				return fid
			},
		},
		{
			name: "file",
			setup: func(t *testing.T, req *SrvReq) *pipeFid {
				root := t.TempDir()
				filePath := filepath.Join(root, "file")
				if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
					t.Fatalf("WriteFile error = %v", err)
				}
				fid := &pipeFid{path: filePath, data: []uint8("hello")}
				req.Tc.Count = 5
				req.Tc.Offset = 0
				return fid
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newPipeReq(Tread)
			fid := tt.setup(t, req)
			req.Fid.Aux = fid

			pipe.Read(req)
			if req.Rc.Type != Rread {
				t.Fatalf("%s type = %d", tt.name, req.Rc.Type)
			}
			switch tt.name {
			case "dir":
				if req.Rc.Count == 0 {
					t.Fatalf("dir read count = 0")
				}
			case "file":
				if string(req.Rc.Data[:req.Rc.Count]) != "hello" {
					t.Fatalf("file read = %q", string(req.Rc.Data[:req.Rc.Count]))
				}
				if len(fid.data) != 0 {
					t.Fatalf("file data not drained")
				}
			}
		})
	}
}

func TestPipefsCreateVariants(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, req *SrvReq, fid *pipeFid)
		wantErr  string
		wantType uint8
	}{
		{
			name: "symlink",
			setup: func(t *testing.T, req *SrvReq, fid *pipeFid) {
				root := t.TempDir()
				fid.path = root
				target := filepath.Join(root, "target")
				if err := os.WriteFile(target, []byte("data"), 0644); err != nil {
					t.Fatalf("WriteFile error = %v", err)
				}
				req.Tc.Name = "link"
				req.Tc.Perm = DMSYMLINK
				req.Tc.Ext = target
				req.Tc.Mode = OREAD
			},
			wantType: Rcreate,
		},
		{
			name: "hardlink",
			setup: func(t *testing.T, req *SrvReq, fid *pipeFid) {
				root := t.TempDir()
				fid.path = root
				target := filepath.Join(root, "target")
				if err := os.WriteFile(target, []byte("data"), 0644); err != nil {
					t.Fatalf("WriteFile error = %v", err)
				}
				linkFid := &SrvFid{Fconn: req.Conn, Aux: &pipeFid{path: target}}
				req.Conn.fidpool[99] = linkFid
				req.Tc.Name = "hardlink"
				req.Tc.Perm = DMLINK
				req.Tc.Ext = "99"
				req.Tc.Mode = OREAD
			},
			wantType: Rcreate,
		},
		{
			name: "device-error",
			setup: func(t *testing.T, req *SrvReq, fid *pipeFid) {
				root := t.TempDir()
				fid.path = root
				req.Tc.Name = "device"
				req.Tc.Perm = DMDEVICE
			},
			wantType: Rerror,
			wantErr:  "not implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newPipeReq(Tcreate)
			fid := &pipeFid{}
			req.Fid.Aux = fid
			tt.setup(t, req, fid)

			(&Pipefs{}).Create(req)
			if req.Rc.Type != tt.wantType {
				t.Fatalf("%s type = %d", tt.name, req.Rc.Type)
			}
			if tt.wantErr != "" && !strings.Contains(req.Rc.Error, tt.wantErr) {
				t.Fatalf("%s error = %q", tt.name, req.Rc.Error)
			}
		})
	}
}

func TestPipefsErrorPaths(t *testing.T) {
	pipe := &Pipefs{}
	tests := []struct {
		name    string
		setup   func(t *testing.T) *SrvReq
		wantErr string
	}{
		{
			name: "attach-afid",
			setup: func(t *testing.T) *SrvReq {
				req := newPipeReq(Tattach)
				req.Afid = &SrvFid{Fconn: req.Conn}
				pipe.Attach(req)
				return req
			},
			wantErr: "no authentication required",
		},
		{
			name: "open-missing",
			setup: func(t *testing.T) *SrvReq {
				req := newPipeReq(Topen)
				missing := filepath.Join(t.TempDir(), "missing")
				req.Fid.Aux = &pipeFid{path: missing}
				pipe.Open(req)
				return req
			},
			wantErr: "no such file",
		},
		{
			name: "remove-missing",
			setup: func(t *testing.T) *SrvReq {
				req := newPipeReq(Tremove)
				missing := filepath.Join(t.TempDir(), "missing")
				req.Fid.Aux = &pipeFid{path: missing}
				pipe.Remove(req)
				return req
			},
			wantErr: "no such file",
		},
		{
			name: "stat-missing",
			setup: func(t *testing.T) *SrvReq {
				req := newPipeReq(Tstat)
				missing := filepath.Join(t.TempDir(), "missing")
				req.Fid.Aux = &pipeFid{path: missing}
				pipe.Stat(req)
				return req
			},
			wantErr: "no such file",
		},
		{
			name: "walk-missing",
			setup: func(t *testing.T) *SrvReq {
				req := newPipeReq(Twalk)
				missing := filepath.Join(t.TempDir(), "missing")
				req.Fid.Aux = &pipeFid{path: missing}
				req.Newfid = &SrvFid{Fconn: req.Conn}
				req.Tc.Wname = []string{"child"}
				pipe.Walk(req)
				return req
			},
			wantErr: "no such file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setup(t)
			if req.Rc.Type != Rerror {
				t.Fatalf("%s type = %d", tt.name, req.Rc.Type)
			}
			if !strings.Contains(req.Rc.Error, tt.wantErr) {
				t.Fatalf("%s error = %q", tt.name, req.Rc.Error)
			}
		})
	}
}
