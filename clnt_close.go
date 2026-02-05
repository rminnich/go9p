// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go9p

// Clunks a fid. Returns nil if successful.
func (clnt *Clnt) Clunk(fid *Fid) error {
	if fid.walked {
		tc := clnt.NewFcall()
		if err := PackTclunk(tc, fid.Fid); err != nil {
			return err
		}

		if _, err := clnt.Rpc(tc); err != nil {
			fid.walked = false
			fid.Fid = NOFID
			return err
		}
	}

	fid.walked = false
	fid.Fid = NOFID
	return nil
}

// Closes a file. Returns nil if successful.
func (file *File) Close() error {
	// Should we cancel all pending requests for the File
	return file.Fid.Clnt.Clunk(file.Fid)
}
