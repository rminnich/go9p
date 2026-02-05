package go9p

import (
	"reflect"
	"strings"
	"testing"
)

func TestPackTauth(t *testing.T) {
	tests := []struct {
		name      string
		dotu      bool
		wantUname uint32
	}{
		{
			name:      "dotu",
			dotu:      true,
			wantUname: 100,
		},
		{
			name:      "plain",
			dotu:      false,
			wantUname: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewFcall(256)
			err := PackTauth(fc, 42, "user", "root", 100, tt.dotu)
			if err != nil {
				t.Fatalf("PackTauth() error = %v", err)
			}
			if fc.Type != Tauth {
				t.Fatalf("PackTauth() type = %d, want %d", fc.Type, Tauth)
			}
			if fc.Fid != 42 || fc.Uname != "user" || fc.Aname != "root" {
				t.Fatalf("PackTauth() fields = %+v", fc)
			}
			if fc.Unamenum != tt.wantUname {
				t.Fatalf("PackTauth() Unamenum = %d, want %d",
					fc.Unamenum, tt.wantUname)
			}
		})
	}
}

func TestPackTflush(t *testing.T) {
	fc := NewFcall(64)
	err := PackTflush(fc, 7)
	if err != nil {
		t.Fatalf("PackTflush() error = %v", err)
	}
	if fc.Type != Tflush || fc.Oldtag != 7 {
		t.Fatalf("PackTflush() fields = %+v", fc)
	}
}

func TestPackTcreate(t *testing.T) {
	tests := []struct {
		name    string
		dotu    bool
		wantExt string
	}{
		{
			name:    "dotu",
			dotu:    true,
			wantExt: "ext",
		},
		{
			name:    "plain",
			dotu:    false,
			wantExt: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewFcall(256)
			err := PackTcreate(fc, 10, "name", 0644, OREAD, "ext", tt.dotu)
			if err != nil {
				t.Fatalf("PackTcreate() error = %v", err)
			}
			if fc.Type != Tcreate || fc.Fid != 10 {
				t.Fatalf("PackTcreate() fields = %+v", fc)
			}
			if fc.Ext != tt.wantExt {
				t.Fatalf("PackTcreate() Ext = %q, want %q", fc.Ext, tt.wantExt)
			}
		})
	}
}

func TestPackTwstat(t *testing.T) {
	fc := NewFcall(256)
	dirValue := Dir{
		Name:   "file",
		Uid:    "u",
		Gid:    "g",
		Muid:   "m",
		Qid:    Qid{Type: QTDIR, Path: 0x1, Version: 0x2},
		Mode:   DMDIR | 0755,
		Atime:  1,
		Mtime:  2,
		Length: 3,
		Type:   4,
		Dev:    5,
		Ext:    "ext",
	}
	err := PackTwstat(fc, 55, &dirValue, true)
	if err != nil {
		t.Fatalf("PackTwstat() error = %v", err)
	}
	if fc.Type != Twstat || fc.Fid != 55 {
		t.Fatalf("PackTwstat() fields = %+v", fc)
	}
	if !reflect.DeepEqual(fc.Dir, dirValue) {
		t.Fatalf("PackTwstat() dir = %+v, want %+v", fc.Dir, dirValue)
	}
}

func TestPackTsimpleOps(t *testing.T) {
	tests := []struct {
		name     string
		call     func(*Fcall) error
		wantType uint8
	}{
		{
			name: "clunk",
			call: func(fc *Fcall) error {
				return PackTclunk(fc, 9)
			},
			wantType: Tclunk,
		},
		{
			name: "remove",
			call: func(fc *Fcall) error {
				return PackTremove(fc, 9)
			},
			wantType: Tremove,
		},
		{
			name: "stat",
			call: func(fc *Fcall) error {
				return PackTstat(fc, 9)
			},
			wantType: Tstat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewFcall(64)
			err := tt.call(fc)
			if err != nil {
				t.Fatalf("%s error = %v", tt.name, err)
			}
			if fc.Type != tt.wantType || fc.Fid != 9 {
				t.Fatalf("%s fields = %+v", tt.name, fc)
			}
		})
	}
}

func TestPackTerrors(t *testing.T) {
	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "tauth-small-buffer",
			call: func() error {
				fc := &Fcall{Buf: make([]byte, 2)}
				return PackTauth(fc, 1, "u", "a", 0, false)
			},
		},
		{
			name: "tcreate-small-buffer",
			call: func() error {
				fc := &Fcall{Buf: make([]byte, 2)}
				return PackTcreate(fc, 1, "n", 0, OREAD, "", false)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if err == nil {
				t.Fatalf("%s expected error", tt.name)
			}
			if !strings.Contains(err.Error(), "buffer too small") {
				t.Fatalf("%s error = %q", tt.name, err.Error())
			}
		})
	}
}
