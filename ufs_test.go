package go9p

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func newUfsConn(root string) (*Ufs, *Conn) {
	ufs := &Ufs{Root: root}
	ufs.Upool = OsUsers
	conn := &Conn{
		Srv:     &ufs.Srv,
		Msize:   MSIZE,
		Dotu:    false,
		reqs:    make(map[uint16]*SrvReq),
		fidpool: make(map[uint32]*SrvFid),
		reqout:  make(chan *SrvReq, 16),
	}
	return ufs, conn
}

func newUfsReq(conn *Conn, msgType uint8) *SrvReq {
	req := &SrvReq{
		Tc:   &Fcall{Type: msgType, Tag: 1},
		Rc:   NewFcall(4096),
		Conn: conn,
		Fid:  &SrvFid{Fconn: conn},
	}
	conn.reqs[req.Tc.Tag] = req
	return req
}

func TestToError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantNum uint32
	}{
		{
			name:    "errno",
			err:     syscall.ENOENT,
			wantNum: uint32(syscall.ENOENT),
		},
		{
			name:    "generic",
			err:     errors.New("boom"),
			wantNum: EIO,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toError(tt.err)
			if got.Errornum != tt.wantNum {
				t.Fatalf("toError num = %d, want %d", got.Errornum, tt.wantNum)
			}
			if !strings.Contains(got.Error(), tt.err.Error()) {
				t.Fatalf("toError message = %q", got.Error())
			}
		})
	}
}

func TestOmode2Uflags(t *testing.T) {
	tests := []struct {
		name string
		mode uint8
		want int
	}{
		{
			name: "read",
			mode: OREAD,
			want: os.O_RDONLY,
		},
		{
			name: "write",
			mode: OWRITE,
			want: os.O_WRONLY,
		},
		{
			name: "read-write",
			mode: ORDWR,
			want: os.O_RDWR,
		},
		{
			name: "exec",
			mode: OEXEC,
			want: os.O_RDONLY,
		},
		{
			name: "truncate",
			mode: OREAD | OTRUNC,
			want: os.O_RDONLY | os.O_TRUNC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := omode2uflags(tt.mode)
			if got != tt.want {
				t.Fatalf("omode2uflags(%d) = %d, want %d", tt.mode, got, tt.want)
			}
		})
	}
}

func TestDir2QidTypeAndMode(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "file")
	if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	symlinkPath := filepath.Join(root, "link")
	if err := os.Symlink(filePath, symlinkPath); err != nil {
		t.Fatalf("Symlink error = %v", err)
	}

	fileInfo, err := os.Lstat(filePath)
	if err != nil {
		t.Fatalf("Lstat error = %v", err)
	}
	linkInfo, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("Lstat link error = %v", err)
	}

	if got := dir2QidType(fileInfo); got != 0 {
		t.Fatalf("dir2QidType file = %d", got)
	}
	if got := dir2QidType(linkInfo); got&QTSYMLINK == 0 {
		t.Fatalf("dir2QidType link = %d", got)
	}
	if got := dir2Npmode(fileInfo, false); got&DMSYMLINK != 0 {
		t.Fatalf("dir2Npmode file = %d", got)
	}
	if got := dir2Npmode(linkInfo, true); got&DMSYMLINK == 0 {
		t.Fatalf("dir2Npmode link = %d", got)
	}
}

func TestDir2Dir(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "file")
	if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	linkPath := filepath.Join(root, "link")
	if err := os.Symlink(filePath, linkPath); err != nil {
		t.Fatalf("Symlink error = %v", err)
	}

	fileInfo, err := os.Lstat(filePath)
	if err != nil {
		t.Fatalf("Lstat error = %v", err)
	}
	linkInfo, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("Lstat link error = %v", err)
	}

	plain, err := dir2Dir(filePath, fileInfo, false, OsUsers)
	if err != nil {
		t.Fatalf("dir2Dir plain error = %v", err)
	}
	if plain.Name != "file" {
		t.Fatalf("dir2Dir plain name = %q", plain.Name)
	}

	dotu, err := dir2Dir(linkPath, linkInfo, true, OsUsers)
	if err != nil {
		t.Fatalf("dir2Dir dotu error = %v", err)
	}
	if dotu.Ext == "" {
		t.Fatalf("dir2Dir dotu ext empty")
	}
	if dotu.Mode&DMSYMLINK == 0 {
		t.Fatalf("dir2Dir dotu mode = %d", dotu.Mode)
	}
}

func TestLookup(t *testing.T) {
	current, err := user.Current()
	if err != nil {
		t.Fatalf("Current user error = %v", err)
	}

	tests := []struct {
		name   string
		value  string
		isGid  bool
		wantOk bool
	}{
		{
			name:   "empty",
			value:  "",
			isGid:  false,
			wantOk: true,
		},
		{
			name:   "current-user",
			value:  current.Username,
			isGid:  false,
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, lookupErr := lookup(tt.value, tt.isGid)
			if (lookupErr == nil) != tt.wantOk {
				t.Fatalf("lookup error = %v", lookupErr)
			}
		})
	}
}

func TestUfsAttachAndStat(t *testing.T) {
	root := t.TempDir()
	ufs, conn := newUfsConn(root)

	attachReq := newUfsReq(conn, Tattach)
	attachReq.Tc.Aname = ""
	// Save Fid.Aux before Attach, since PostProcess clears req.Fid.
	fid := attachReq.Fid
	ufs.Attach(attachReq)
	if attachReq.Rc.Type != Rattach {
		t.Fatalf("Attach type = %d", attachReq.Rc.Type)
	}
	if fid.Aux == nil {
		t.Fatalf("Attach fid aux not set")
	}

	statReq := newUfsReq(conn, Tstat)
	statReq.Fid.Aux = &ufsFid{path: root}
	ufs.Stat(statReq)
	if statReq.Rc.Type != Rstat {
		t.Fatalf("Stat type = %d", statReq.Rc.Type)
	}
}

func TestUfsReadWriteRemove(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "file")
	if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	file, err := os.OpenFile(filePath, os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("OpenFile error = %v", err)
	}
	t.Cleanup(func() {
		_ = file.Close()
	})

	ufs, conn := newUfsConn(root)
	readReq := newUfsReq(conn, Tread)
	readReq.Tc.Count = 5
	readReq.Tc.Offset = 0
	readReq.Fid.Aux = &ufsFid{path: filePath, file: file}
	ufs.Read(readReq)
	if readReq.Rc.Type != Rread {
		t.Fatalf("Read type = %d", readReq.Rc.Type)
	}
	if got := string(readReq.Rc.Data[:readReq.Rc.Count]); got != "hello" {
		t.Fatalf("Read data = %q", got)
	}

	writeReq := newUfsReq(conn, Twrite)
	writeReq.Tc.Data = []byte("bye")
	writeReq.Tc.Offset = 0
	writeReq.Tc.Count = uint32(len(writeReq.Tc.Data))
	writeReq.Fid.Aux = &ufsFid{path: filePath, file: file}
	ufs.Write(writeReq)
	if writeReq.Rc.Type != Rwrite {
		t.Fatalf("Write type = %d", writeReq.Rc.Type)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	if !strings.HasPrefix(string(content), "bye") {
		t.Fatalf("Write content = %q", string(content))
	}

	removeReq := newUfsReq(conn, Tremove)
	removeReq.Fid.Aux = &ufsFid{path: filePath}
	ufs.Remove(removeReq)
	if removeReq.Rc.Type != Rremove {
		t.Fatalf("Remove type = %d", removeReq.Rc.Type)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("Remove err = %v", err)
	}
}

func TestUfsWstatRename(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "file")
	if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	ufs, conn := newUfsConn(root)
	wstatReq := newUfsReq(conn, Twstat)
	wstatReq.Fid.Aux = &ufsFid{path: filePath}
	wstatReq.Tc.Dir = Dir{
		Mode:   0xFFFFFFFF,
		Length: 0xFFFFFFFFFFFFFFFF,
		Mtime:  ^uint32(0),
		Atime:  ^uint32(0),
		Name:   "renamed",
	}
	ufs.Wstat(wstatReq)
	if wstatReq.Rc.Type != Rwstat {
		t.Fatalf("Wstat type = %d", wstatReq.Rc.Type)
	}
	if _, err := os.Stat(filepath.Join(root, "renamed")); err != nil {
		t.Fatalf("Rename stat error = %v", err)
	}
}
