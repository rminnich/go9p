package go9p

import "testing"

type testGroup struct {
	name string
	id   int
}

func (g testGroup) Name() string { return g.name }
func (g testGroup) Id() int      { return g.id }
func (g testGroup) Members() []User {
	return nil
}

type testUser struct {
	name   string
	id     int
	groups []Group
}

func (u testUser) Name() string    { return u.name }
func (u testUser) Id() int         { return u.id }
func (u testUser) Groups() []Group { return u.groups }
func (u testUser) IsMember(g Group) bool {
	return false
}

func TestSrvFileAddAndFind(t *testing.T) {
	owner := testUser{name: "owner", id: 1}
	group := testGroup{name: "group", id: 2}
	root := &srvFile{}
	if err := root.Add(nil, "root", owner, group, DMDIR|0755, nil); err != nil {
		t.Fatalf("Add root error = %v", err)
	}

	child := &srvFile{}
	if err := child.Add(root, "child", owner, group, 0644, nil); err != nil {
		t.Fatalf("Add child error = %v", err)
	}
	if got := root.Find("child"); got != child {
		t.Fatalf("Find(child) = %+v, want %+v", got, child)
	}

	dup := &srvFile{}
	if err := dup.Add(root, "child", owner, group, 0644, nil); err != Eexist {
		t.Fatalf("Add duplicate error = %v, want %v", err, Eexist)
	}

	noneUser := &srvFile{}
	if err := noneUser.Add(root, "anon", nil, nil, 0644, nil); err != nil {
		t.Fatalf("Add anon error = %v", err)
	}
	if noneUser.Uid != "none" || noneUser.Gid != "none" {
		t.Fatalf("Add anon user gid = %q %q", noneUser.Uid, noneUser.Gid)
	}
}

func TestSrvFileRemoveAndRename(t *testing.T) {
	owner := testUser{name: "owner", id: 1}
	group := testGroup{name: "group", id: 2}
	root := &srvFile{}
	if err := root.Add(nil, "root", owner, group, DMDIR|0755, nil); err != nil {
		t.Fatalf("Add root error = %v", err)
	}

	child := &srvFile{}
	if err := child.Add(root, "child", owner, group, 0644, nil); err != nil {
		t.Fatalf("Add child error = %v", err)
	}
	other := &srvFile{}
	if err := other.Add(root, "other", owner, group, 0644, nil); err != nil {
		t.Fatalf("Add other error = %v", err)
	}

	if err := child.Rename("other"); err != Eexist {
		t.Fatalf("Rename error = %v, want %v", err, Eexist)
	}
	if err := child.Rename("renamed"); err != nil {
		t.Fatalf("Rename error = %v", err)
	}
	if child.Name != "renamed" {
		t.Fatalf("Rename name = %q", child.Name)
	}

	child.Remove()
	if got := root.Find("renamed"); got != nil {
		t.Fatalf("Find after remove = %+v", got)
	}
}

func TestSrvFileCheckPerm(t *testing.T) {
	ownerGroup := testGroup{name: "staff", id: 2}
	owner := testUser{name: "owner", id: 1, groups: []Group{ownerGroup}}
	member := testUser{name: "member", id: 3, groups: []Group{ownerGroup}}
	other := testUser{name: "other", id: 4}

	file := &srvFile{
		Dir: Dir{
			Mode:   0644,
			Uid:    owner.name,
			Uidnum: uint32(owner.id),
			Gid:    ownerGroup.name,
			Gidnum: uint32(ownerGroup.id),
		},
	}

	tests := []struct {
		name string
		user User
		perm uint32
		want bool
	}{
		{
			name: "owner-write",
			user: owner,
			perm: DMWRITE,
			want: true,
		},
		{
			name: "group-read",
			user: member,
			perm: DMREAD,
			want: true,
		},
		{
			name: "other-read",
			user: other,
			perm: DMREAD,
			want: true,
		},
		{
			name: "other-write",
			user: other,
			perm: DMWRITE,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := file.CheckPerm(tt.user, tt.perm)
			if got != tt.want {
				t.Fatalf("CheckPerm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMode2Perm(t *testing.T) {
	tests := []struct {
		name string
		mode uint8
		want uint32
	}{
		{
			name: "read",
			mode: OREAD,
			want: DMREAD,
		},
		{
			name: "write",
			mode: OWRITE,
			want: DMWRITE,
		},
		{
			name: "read-write",
			mode: ORDWR,
			want: DMREAD | DMWRITE,
		},
		{
			name: "truncate",
			mode: OREAD | OTRUNC,
			want: DMREAD | DMWRITE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mode2Perm(tt.mode)
			if got != tt.want {
				t.Fatalf("mode2Perm(%d) = %d, want %d", tt.mode, got, tt.want)
			}
		})
	}
}
