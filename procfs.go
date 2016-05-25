package go9p

import (
	"log"
	"os"
	pathpkg "path"
	"strings"
)

type Procfs struct {
	*Ufs
	*Fsrv
}

func (pfs *Procfs) Attach(req *SrvReq) {
	switch req.Tc.Aname {
	case "/d":
		pfs.Ufs.Attach(req)
	case "", "/":
		pfs.Fsrv.Attach(req)
	}
}

type stfn func(*wctx) stfn

type wctx struct {
	req *SrvReq
	fid interface{}
	nfid interface{}
	qids []Qid
	path string
	count int
	fp *SrvFile
	mnt string
	ufsRoot string
	fsRoot *SrvFile
}

func walker(req *SrvReq, ctx *wctx) {
	st := st_init(ctx)
	for st != nil {
		st = st(ctx)
	}
}

func st_init(ctx *wctx) stfn {
	ctx.qids = make([]Qid, len(ctx.req.Tc.Wname))
	switch ctx.req.Fid.Aux.(type) {
	case *ufsFid:
		req := ctx.req
		tc := req.Tc
		fid := req.Fid.Aux.(*ufsFid)
		ctx.fid = fid
		
		err := fid.stat()
		if err != nil {
			log.Printf("failed to stat fid %+v error %+v", fid, err)
			req.RespondError(err)
			return nil
		}
		ctx.path = fid.path
		if req.Newfid.Aux == nil {
			req.Newfid.Aux = new(ufsFid)
		}

		ctx.nfid = req.Newfid.Aux.(*ufsFid)
		ctx.count = len(tc.Wname)
		return st_ufswalk

	case *FFid:
		req := ctx.req
		fid := req.Fid.Aux.(*FFid)
		if req.Newfid.Aux == nil {
			nfid := new(FFid)
			nfid.Fid = req.Newfid
			req.Newfid.Aux = nfid
		}

		ctx.nfid = req.Newfid.Aux.(*FFid)
		ctx.count = len(req.Tc.Wname)
		ctx.fp = fid.F
		return st_fswalk
	}
	return nil
}

func st_ufswalk(ctx *wctx) stfn {
	path := ctx.path
	req := ctx.req
	tc := req.Tc
	i := len(tc.Wname) - ctx.count
	if len(tc.Wname) == 0 {
		return st_ufswalk_done
	}
	p := path + "/" + tc.Wname[i]
	if "/" + tc.Wname[i] == ctx.mnt && i == 0  {
		log.Println("found the mountpoint prefix on item 0 so I'm removing it")
		p = ctx.ufsRoot
	}
	if tc.Wname[i] == ".." {
		chk := pathpkg.Join("realroot", p)
		if chk == "realroot" {
			fid := new(FFid)
			fid.F = ctx.fsRoot
			req.Fid.Aux = fid
			ctx.nfid = new(FFid)
			req.Newfid.Aux = ctx.nfid
			return st_fswalk
		}
	}
	st, err := os.Lstat(p)
	if err != nil {
		if i == 0 {
			log.Println("Enoent", err)
			req.RespondError(Enoent)
			return nil
		}
		log.Printf("this is the error about failed os.Lstat on %+v: %v", p, err)
		return st_ufswalk_done
	}
	ctx.qids[i] = *dir2Qid(st)
	ctx.path = p
	ctx.count -= 1
	if ctx.count == 0 {
		return st_ufswalk_done
	}
	return st_ufswalk
}

func st_ufswalk_done(ctx *wctx) stfn {
	req := ctx.req
	tc := req.Tc
	if ctx.count != 0 {
		log.Println("oops did not read all of Wname", tc.Wname)
	}
	ctx.nfid.(*ufsFid).path = ctx.path
	ctx.req.RespondRwalk(ctx.qids[0:len(tc.Wname) - ctx.count])
	return nil
}


func st_fswalk(ctx *wctx) stfn {
	req := ctx.req
	tc := req.Tc
	i :=  len(tc.Wname) - ctx.count
	f := ctx.fp
	wqids := ctx.qids
	if len(tc.Wname) == 0 {
		return st_fswalk_done
	}
	if tc.Wname[i] == ".." {
		// handle dotdot
		f = f.Parent
		wqids[i] = f.Qid
		ctx.count -= 1
		if ctx.count == 0 {
			return st_fswalk_done
		}
		return st_fswalk
	}
	if "/" + tc.Wname[i] == ctx.mnt {
		req.Fid.Aux = new(ufsFid)
		req.Fid.Aux.(*ufsFid).path = "/"
		ctx.nfid = new(ufsFid)
		req.Newfid.Aux = ctx.nfid 
		return st_ufswalk
	}
	if (wqids[i].Type & QTDIR) > 0 {
		if !f.CheckPerm(req.Fid.User, DMEXEC) {
			return st_fswalk_done
		}
	}

	p := f.Find(tc.Wname[i])
	if p == nil {
		return st_fswalk_done
	}
	f = p
	ctx.fp = f
	wqids[i] = f.Qid
	ctx.count -= 1
	if ctx.count == 0 {
		return st_fswalk_done
	}
	return st_fswalk
}

func st_fswalk_done(ctx *wctx) stfn {
	req := ctx.req
	tc := req.Tc
	if len(tc.Wname) > 0 && ctx.count != 0 {
		req.RespondError(Enoent)
		return nil
	}
	ctx.nfid.(*FFid).F = ctx.fp
	req.RespondRwalk(ctx.qids[0:len(tc.Wname) - ctx.count])
	return nil
}


// hasPathPrefix returns true if x == y or x == y + "/" + more
func hasPathPrefix(x, y string) bool {
	return x == y || strings.HasPrefix(x, y) && (strings.HasSuffix(y, "/") || strings.HasPrefix(x[len(y):], "/"))
}

func translate(old, new, path string) string {
	path = pathpkg.Clean("/" + path)
	if !hasPathPrefix(path, old) {
		log.Println("translate " + path + " but old=" + old)
	}
	return pathpkg.Join(new, path[len(old):])
}


func (pfs *Procfs) Walk(req *SrvReq) {
	ctx := new(wctx)
	ctx.req = req
	ctx.ufsRoot = pfs.Ufs.Root
	ctx.fsRoot = pfs.Fsrv.Root
	ctx.mnt = "/d"
	walker(req, ctx)

}

func (pfs *Procfs) Open(req *SrvReq) {
	switch req.Fid.Aux.(type) {
	case *ufsFid:
		pfs.Ufs.Open(req)
	case *FFid:
		pfs.Fsrv.Open(req)
	}
}

func (pfs *Procfs) Create(req *SrvReq) {
	switch req.Fid.Aux.(type) {
	case *ufsFid:
		pfs.Ufs.Create(req)
	case *FFid:
		pfs.Fsrv.Create(req)
	}
}

func (pfs *Procfs) Read(req *SrvReq) {
	switch req.Fid.Aux.(type) {
	case *ufsFid:
		pfs.Ufs.Read(req)
	case *FFid:
		pfs.Fsrv.Read(req)
	}
}

func (pfs *Procfs) Write(req *SrvReq) {
	switch req.Fid.Aux.(type) {
	case *ufsFid:
		pfs.Ufs.Write(req)
	case *FFid:
		pfs.Fsrv.Write(req)
	}
}

func (pfs *Procfs) Clunk(req *SrvReq) {
	switch req.Fid.Aux.(type) {
	case *ufsFid:
		pfs.Ufs.Clunk(req)
	case *FFid:
		pfs.Fsrv.Clunk(req)
	}
}

func (pfs *Procfs) Remove(req *SrvReq) {
	switch req.Fid.Aux.(type) {
	case *ufsFid:
		pfs.Ufs.Remove(req)
	case *FFid:
		pfs.Fsrv.Remove(req)
	}
}

func (pfs *Procfs) Stat(req *SrvReq) {
	switch req.Fid.Aux.(type) {
	case *ufsFid:
		pfs.Ufs.Stat(req)
	case *FFid:
		pfs.Fsrv.Stat(req)
	}
}

func (pfs *Procfs) Wstat(req *SrvReq) {
	switch req.Fid.Aux.(type) {
	case *ufsFid:
		pfs.Ufs.Wstat(req)
	case *FFid:
		pfs.Fsrv.Wstat(req)
	}
}
