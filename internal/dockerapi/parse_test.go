package dockerapi

import "testing"

func TestParseDockerHost(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		{"unix:///var/run/docker.sock", "/var/run/docker.sock", true},
		{"/run/user/1000/docker.sock", "/run/user/1000/docker.sock", true},
		{"tcp://127.0.0.1:2375", "", false},
		{"", "", false},
	}
	for _, tc := range tests {
		got, ok := ParseDockerHost(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Errorf("ParseDockerHost(%q) = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestExpandListenStream(t *testing.T) {
	got := ExpandListenStream("%t/docker.sock", "/run/user/1000", 1000)
	if got != "/run/user/1000/docker.sock" {
		t.Fatalf("got %q", got)
	}
	got = ExpandListenStream("/var/run/docker.sock", "", 0)
	if got != "/var/run/docker.sock" {
		t.Fatalf("got %q", got)
	}
}

func TestParseSystemdShow(t *testing.T) {
	out := "ListenStream=/run/docker.sock\nExecStart={ path=/usr/bin/dockerd ; argv[]=dockerd -H fd:// }"
	got, ok := ParseSystemdShow(out, "/run", 0)
	if !ok || got != "/run/docker.sock" {
		t.Fatalf("ParseSystemdShow = (%q, %v)", got, ok)
	}
	out = "ExecStart={ path=/usr/bin/dockerd ; argv[]=dockerd -H unix:///var/run/docker.sock }"
	got, ok = ParseSystemdShow(out, "", 0)
	if !ok || got != "/var/run/docker.sock" {
		t.Fatalf("ExecStart parse = (%q, %v)", got, ok)
	}
}
