package go9p

import "testing"

func TestPoolPut(t *testing.T) {
	tests := []struct {
		name      string
		id        uint32
		wantPanic bool
	}{
		{
			name:      "in-range",
			id:        2,
			wantPanic: false,
		},
		{
			name:      "out-of-range",
			id:        3,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewPool(1, 2)
			if tt.wantPanic {
				defer func() {
					if recover() == nil {
						t.Fatalf("expected panic for id %d", tt.id)
					}
				}()
				pool.Put(tt.id)
				return
			}
			pool.Put(tt.id)
			first := pool.Get()
			second := pool.Get()
			if first != 1 || second != tt.id {
				t.Fatalf("Pool values = %d, %d", first, second)
			}
		})
	}
}
