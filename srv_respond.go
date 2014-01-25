// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go9p

import "fmt"

// srvRequest operations. This interface should be implemented by all file servers.
// The operations correspond directly to most of the 9P2000 message types.
type srvReqOps interface {
	Attach(*srvReq)
	Walk(*srvReq)
	Open(*srvReq)
	Create(*srvReq)
	Read(*srvReq)
	Write(*srvReq)
	Clunk(*srvReq)
	Remove(*srvReq)
	Stat(*srvReq)
	Wstat(*srvReq)
}

// Respond to the request with Rerror message
func (req *srvReq) RespondError(err interface{}) {
	switch e := err.(type) {
	case *Error:
		PackRerror(req.Rc, e.Error(), uint32(e.Errornum), req.Conn.Dotu)
	case error:
		PackRerror(req.Rc, e.Error(), uint32(EIO), req.Conn.Dotu)
	default:
		PackRerror(req.Rc, fmt.Sprintf("%v", e), uint32(EIO), req.Conn.Dotu)
	}

	req.Respond()
}

// Respond to the request with Rversion message
func (req *srvReq) RespondRversion(msize uint32, version string) {
	err := PackRversion(req.Rc, msize, version)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rauth message
func (req *srvReq) RespondRauth(aqid *Qid) {
	err := PackRauth(req.Rc, aqid)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rflush message
func (req *srvReq) RespondRflush() {
	err := PackRflush(req.Rc)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rattach message
func (req *srvReq) RespondRattach(aqid *Qid) {
	err := PackRattach(req.Rc, aqid)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rwalk message
func (req *srvReq) RespondRwalk(wqids []Qid) {
	err := PackRwalk(req.Rc, wqids)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Ropen message
func (req *srvReq) RespondRopen(qid *Qid, iounit uint32) {
	err := PackRopen(req.Rc, qid, iounit)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rcreate message
func (req *srvReq) RespondRcreate(qid *Qid, iounit uint32) {
	err := PackRcreate(req.Rc, qid, iounit)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rread message
func (req *srvReq) RespondRread(data []byte) {
	err := PackRread(req.Rc, data)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rwrite message
func (req *srvReq) RespondRwrite(count uint32) {
	err := PackRwrite(req.Rc, count)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rclunk message
func (req *srvReq) RespondRclunk() {
	err := PackRclunk(req.Rc)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rremove message
func (req *srvReq) RespondRremove() {
	err := PackRremove(req.Rc)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rstat message
func (req *srvReq) RespondRstat(st *Dir) {
	err := PackRstat(req.Rc, st, req.Conn.Dotu)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}

// Respond to the request with Rwstat message
func (req *srvReq) RespondRwstat() {
	err := PackRwstat(req.Rc)
	if err != nil {
		req.RespondError(err)
	} else {
		req.Respond()
	}
}
