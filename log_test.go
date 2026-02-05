package go9p

import (
	"fmt"
	"testing"
	"time"
)

func waitForLogs(t *testing.T, logger *Logger, owner interface{}, itype int, want int) []*Log {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		items := logger.Filter(owner, itype)
		if len(items) == want {
			return items
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for %d logs, got %d", want, len(items))
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestLoggerFilter(t *testing.T) {
	logger := NewLogger(4)
	if logger == nil {
		t.Fatalf("NewLogger returned nil")
	}

	logger.Log("first", "owner", 1)
	logger.Log("second", "owner", 2)
	logger.Log("third", "other", 1)

	tests := []struct {
		name  string
		owner interface{}
		itype int
		want  int
	}{
		{
			name:  "all",
			owner: nil,
			itype: 0,
			want:  3,
		},
		{
			name:  "owner",
			owner: "owner",
			itype: 0,
			want:  2,
		},
		{
			name:  "type",
			owner: nil,
			itype: 1,
			want:  2,
		},
		{
			name:  "owner-type",
			owner: "owner",
			itype: 1,
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := waitForLogs(t, logger, tt.owner, tt.itype, tt.want)
			if len(items) != tt.want {
				t.Fatalf("Filter count = %d, want %d", len(items), tt.want)
			}
		})
	}
}

func TestLoggerResize(t *testing.T) {
	logger := NewLogger(2)
	if logger == nil {
		t.Fatalf("NewLogger returned nil")
	}

	logger.Log("first", "owner", 1)
	logger.Log("second", "owner", 1)
	waitForLogs(t, logger, nil, 0, 2)

	logger.Resize(1)
	logger.Log("third", "owner", 1)

	// After resize to 1 and logging "third", the single slot holds "third".
	// Poll until the goroutine processes the log (count stays 1 but data changes).
	deadline := time.Now().Add(time.Second)
	for {
		items := logger.Filter(nil, 0)
		if len(items) == 1 && fmt.Sprintf("%v", items[0].Data) == "third" {
			return
		}
		if time.Now().After(deadline) {
			var got string
			if len(items) > 0 {
				got = fmt.Sprintf("%v", items[0].Data)
			}
			t.Fatalf("timeout: got %q, want %q", got, "third")
		}
		time.Sleep(5 * time.Millisecond)
	}
}
