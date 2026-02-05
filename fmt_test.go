package go9p

import (
	"fmt"
	"testing"
)

func TestPermToString(t *testing.T) {
	tests := []struct {
		name string
		perm uint32
		want string
	}{
		{
			name: "dir-append",
			perm: DMDIR | DMAPPEND | 0755,
			want: "da755",
		},
		{
			name: "auth-excl-special",
			perm: DMAUTH | DMEXCL | DMTMP | DMDEVICE | DMSOCKET |
				DMNAMEDPIPE | DMSYMLINK | 0644,
			want: "AltDSPL644",
		},
		{
			name: "plain",
			perm: 0600,
			want: "600",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := permToString(tt.perm)
			if got != tt.want {
				t.Fatalf("permToString(%#o) = %q, want %q",
					tt.perm, got, tt.want)
			}
		})
	}
}

func TestQidString(t *testing.T) {
	tests := []struct {
		name string
		qid  Qid
		want string
	}{
		{
			name: "dir-symlink",
			qid:  Qid{Type: QTDIR | QTSYMLINK, Path: 0x12, Version: 0x34},
			want: "(12 34 'dL')",
		},
		{
			name: "append-auth",
			qid:  Qid{Type: QTAPPEND | QTAUTH, Path: 0x1, Version: 0x2},
			want: "(1 2 'aA')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.qid.String()
			if got != tt.want {
				t.Fatalf("Qid.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDirString(t *testing.T) {
	tests := []struct {
		name string
		dir  Dir
		want string
	}{
		{
			name: "with-ext",
			dir: Dir{
				Name:   "name",
				Uid:    "user",
				Gid:    "group",
				Muid:   "mod",
				Qid:    Qid{Type: QTDIR | QTAUTH, Path: 0x1, Version: 0x2},
				Mode:   DMDIR | 0755,
				Atime:  10,
				Mtime:  11,
				Length: 12,
				Type:   13,
				Dev:    14,
				Ext:    "ext",
			},
			want: "'name' 'user' 'group' 'mod' q (1 2 'dA') m d755 " +
				"at 10 mt 11 l 12 t 13 d 14 ext ext",
		},
		{
			name: "no-ext",
			dir: Dir{
				Name:   "file",
				Uid:    "owner",
				Gid:    "group",
				Muid:   "mod",
				Qid:    Qid{Type: 0, Path: 0xff, Version: 0x1},
				Mode:   0644,
				Atime:  1,
				Mtime:  2,
				Length: 3,
				Type:   4,
				Dev:    5,
				Ext:    "",
			},
			want: "'file' 'owner' 'group' 'mod' q (ff 1 '') m 644 " +
				"at 1 mt 2 l 3 t 4 d 5 ext ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dir.String()
			if got != tt.want {
				t.Fatalf("Dir.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFcallString(t *testing.T) {
	tests := []struct {
		name string
		fc   Fcall
		want string
	}{
		{
			name: "invalid",
			fc:   Fcall{Type: 0},
			want: "invalid call: 0",
		},
		{
			name: "tcreate",
			fc: Fcall{
				Type: Tcreate,
				Tag:  1,
				Fid:  2,
				Name: "file",
				Perm: DMDIR | 0755,
				Mode: OWRITE,
			},
			want: fmt.Sprintf(
				"Tcreate tag 1 fid 2 name 'file' perm %s mode 1 ",
				permToString(DMDIR|0755),
			),
		},
		{
			name: "rerror",
			fc: Fcall{
				Type:     Rerror,
				Tag:      3,
				Error:    "boom",
				Errornum: 5,
			},
			want: "Rerror tag 3 ename 'boom' ecode 5",
		},
		{
			name: "twalk",
			fc: Fcall{
				Type:   Twalk,
				Tag:    4,
				Fid:    9,
				Newfid: 10,
				Wname:  []string{"a", "b"},
			},
			want: "Twalk tag 4 fid 9 newfid 10 0:'a' 1:'b' ",
		},
		{
			name: "ropen",
			fc: Fcall{
				Type:   Ropen,
				Tag:    5,
				Qid:    Qid{Type: QTDIR, Path: 0x1, Version: 0x2},
				Iounit: 4096,
			},
			want: "Ropen tag 5 qid (1 2 'd') iounit 4096",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fc.String()
			if got != tt.want {
				t.Fatalf("Fcall.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
