package go9p

import "testing"

func TestOsUsersLookup(t *testing.T) {
	userFirst := OsUsers.Uid2User(100)
	userSecond := OsUsers.Uid2User(100)
	if userFirst == nil || userSecond == nil {
		t.Fatalf("Uid2User returned nil")
	}
	if userFirst.Id() != 100 || userSecond.Id() != 100 {
		t.Fatalf("Uid2User id mismatch: %d %d", userFirst.Id(), userSecond.Id())
	}
	if userFirst != userSecond {
		t.Fatalf("Uid2User should return cached user")
	}

	groupFirst := OsUsers.Gid2Group(200)
	groupSecond := OsUsers.Gid2Group(200)
	if groupFirst == nil || groupSecond == nil {
		t.Fatalf("Gid2Group returned nil")
	}
	if groupFirst.Id() != 200 || groupSecond.Id() != 200 {
		t.Fatalf("Gid2Group id mismatch: %d %d", groupFirst.Id(), groupSecond.Id())
	}
	if groupFirst != groupSecond {
		t.Fatalf("Gid2Group should return cached group")
	}
}

func TestOsUsersUnsupportedLookups(t *testing.T) {
	tests := []struct {
		name string
		call func() interface{}
	}{
		{
			name: "uname",
			call: func() interface{} {
				return OsUsers.Uname2User("name")
			},
		},
		{
			name: "gname",
			call: func() interface{} {
				return OsUsers.Gname2Group("group")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.call() != nil {
				t.Fatalf("%s expected nil", tt.name)
			}
		})
	}
}
