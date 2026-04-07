package sshexec

import "testing"

func TestParseUserHost(t *testing.T) {
	cases := []struct {
		in       string
		wantUser string
		wantAddr string
	}{
		{"root@192.168.1.1", "root", "192.168.1.1:22"},
		{"u@myhost:2222", "u", "myhost:2222"},
		{"u@[::1]:22", "u", "[::1]:22"},
	}
	for _, tc := range cases {
		u, a, err := ParseUserHost(tc.in)
		if err != nil || u != tc.wantUser || a != tc.wantAddr {
			t.Fatalf("%q: got %q %q err=%v want %q %q", tc.in, u, a, err, tc.wantUser, tc.wantAddr)
		}
	}
}
