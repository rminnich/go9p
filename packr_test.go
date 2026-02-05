package go9p

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestPackRauth(t *testing.T) {
	fc := NewFcall(128)
	qidValue := Qid{Type: QTDIR, Path: 0x1, Version: 0x2}
	err := PackRauth(fc, &qidValue)
	if err != nil {
		t.Fatalf("PackRauth() error = %v", err)
	}
	if fc.Type != Rauth {
		t.Fatalf("PackRauth() type = %d, want %d", fc.Type, Rauth)
	}
	if !reflect.DeepEqual(fc.Qid, qidValue) {
		t.Fatalf("PackRauth() qid = %+v, want %+v", fc.Qid, qidValue)
	}
}

func TestPackRflush(t *testing.T) {
	fc := NewFcall(32)
	err := PackRflush(fc)
	if err != nil {
		t.Fatalf("PackRflush() error = %v", err)
	}
	if fc.Type != Rflush {
		t.Fatalf("PackRflush() type = %d, want %d", fc.Type, Rflush)
	}
}

func TestPackRcreate(t *testing.T) {
	fc := NewFcall(128)
	qidValue := Qid{Type: QTAUTH, Path: 0x9, Version: 0x1}
	err := PackRcreate(fc, &qidValue, 2048)
	if err != nil {
		t.Fatalf("PackRcreate() error = %v", err)
	}
	if fc.Type != Rcreate || fc.Iounit != 2048 {
		t.Fatalf("PackRcreate() fields = %+v", fc)
	}
	if !reflect.DeepEqual(fc.Qid, qidValue) {
		t.Fatalf("PackRcreate() qid = %+v, want %+v", fc.Qid, qidValue)
	}
}

func TestPackRread(t *testing.T) {
	fc := NewFcall(128)
	data := []byte("data")
	err := PackRread(fc, data)
	if err != nil {
		t.Fatalf("PackRread() error = %v", err)
	}
	if fc.Type != Rread || fc.Count != uint32(len(data)) {
		t.Fatalf("PackRread() fields = %+v", fc)
	}
	if !bytes.Equal(fc.Data, data) {
		t.Fatalf("PackRread() data = %q, want %q", fc.Data, data)
	}
}

func TestPackRstat(t *testing.T) {
	fc := NewFcall(256)
	dirValue := Dir{
		Name:   "file",
		Uid:    "u",
		Gid:    "g",
		Muid:   "m",
		Qid:    Qid{Type: QTDIR, Path: 0x3, Version: 0x4},
		Mode:   DMDIR | 0755,
		Atime:  2,
		Mtime:  3,
		Length: 4,
		Type:   5,
		Dev:    6,
		Ext:    "ext",
	}
	err := PackRstat(fc, &dirValue, true)
	if err != nil {
		t.Fatalf("PackRstat() error = %v", err)
	}
	if fc.Type != Rstat {
		t.Fatalf("PackRstat() type = %d, want %d", fc.Type, Rstat)
	}
	if !reflect.DeepEqual(fc.Dir, dirValue) {
		t.Fatalf("PackRstat() dir = %+v, want %+v", fc.Dir, dirValue)
	}
}

func TestPackRsimpleOps(t *testing.T) {
	tests := []struct {
		name     string
		call     func(*Fcall) error
		wantType uint8
	}{
		{
			name:     "clunk",
			call:     PackRclunk,
			wantType: Rclunk,
		},
		{
			name:     "remove",
			call:     PackRremove,
			wantType: Rremove,
		},
		{
			name:     "wstat",
			call:     PackRwstat,
			wantType: Rwstat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewFcall(32)
			err := tt.call(fc)
			if err != nil {
				t.Fatalf("%s error = %v", tt.name, err)
			}
			if fc.Type != tt.wantType {
				t.Fatalf("%s type = %d, want %d", tt.name, fc.Type,
					tt.wantType)
			}
		})
	}
}

func TestPackRerrorAkaros(t *testing.T) {
	previous := *Akaros
	*Akaros = true
	t.Cleanup(func() {
		*Akaros = previous
	})

	fc := NewFcall(128)
	err := PackRerror(fc, "boom", 1, true)
	if err != nil {
		t.Fatalf("PackRerror() error = %v", err)
	}
	if fc.Type != Rerror || fc.Errornum != 1 {
		t.Fatalf("PackRerror() fields = %+v", fc)
	}
	if !strings.HasPrefix(fc.Error, "0001 ") {
		t.Fatalf("PackRerror() error = %q", fc.Error)
	}
}
