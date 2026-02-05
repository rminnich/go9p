package go9p

import (
	"strings"
	"testing"
)

func TestSrvAndConnString(t *testing.T) {
	srv := &Srv{Id: "srv"}
	if srv.String() != "srv" {
		t.Fatalf("Srv.String() = %q", srv.String())
	}
	conn := &Conn{Srv: srv, Id: "conn"}
	if conn.String() != "srv/conn" {
		t.Fatalf("Conn.String() = %q", conn.String())
	}
}

func TestSrvReqFlush(t *testing.T) {
	req := newTestReq(Tflush)
	req.Flush()
	if req.status&reqFlush == 0 {
		t.Fatalf("reqFlush not set")
	}
	if req.status&reqResponded == 0 {
		t.Fatalf("reqResponded not set")
	}
}

func TestSrvReqProcessErrors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, req *SrvReq)
		wantErr string
	}{
		{
			name: "unknown-type",
			setup: func(t *testing.T, req *SrvReq) {
				req.Tc.Type = 0xff
				req.Tc.Fid = NOFID
			},
			wantErr: "unknown message type",
		},
		{
			name: "unknown-fid",
			setup: func(t *testing.T, req *SrvReq) {
				req.Tc.Type = Topen
				req.Tc.Fid = 123
			},
			wantErr: "unknown fid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newTestReq(Tversion)
			tt.setup(t, req)
			req.Process()
			if req.Rc.Type != Rerror {
				t.Fatalf("%s type = %d", tt.name, req.Rc.Type)
			}
			if !strings.Contains(req.Rc.Error, tt.wantErr) {
				t.Fatalf("%s error = %q", tt.name, req.Rc.Error)
			}
		})
	}
}
