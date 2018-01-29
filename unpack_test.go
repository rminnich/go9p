// Copyright 2018 The Go9p Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go9p

import (
	"errors"
	"testing"
)

func TestUnpack(t *testing.T) {
	fcalls := []Fcall{
		{Type: Rwstat},
	}

	for _, fc := range fcalls {
		fc.Buf = make([]byte, MSIZE)
		err := errors.New("unknown Fcall Type")
		switch fc.Type {
		case Rwstat:
			err = PackRwstat(&fc)
		}
		if err != nil {
			t.Fatalf("Pack failed for %s: %v\n", &fc, err)
		}
		for _, dotu := range []bool{true, false} {
			fc1, err, _ := Unpack(fc.Pkt, dotu)
			if err != nil {
				t.Fatalf("Unpack failed: %v\n", err)
			}
			if fc.Type != fc1.Type || fc.Tag != fc1.Tag {
				t.Errorf("Fcall is %s; expected %s\n", &fc, fc1)
			}
		}
	}
}
