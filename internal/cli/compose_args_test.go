package cli

import "testing"

func strSliceEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSplitLeadingServiceArgs(t *testing.T) {
	tests := []struct {
		args     []string
		services []string
		rest     []string
	}{
		{nil, nil, nil},
		{[]string{}, nil, nil},
		{[]string{"parser"}, []string{"parser"}, nil},
		{[]string{"parser", "worker"}, []string{"parser", "worker"}, nil},
		{[]string{"--tail", "50"}, nil, []string{"--tail", "50"}},
		{[]string{"parser", "--tail", "50"}, []string{"parser"}, []string{"--tail", "50"}},
		{[]string{"a", "b", "-x"}, []string{"a", "b"}, []string{"-x"}},
	}
	for _, tt := range tests {
		s, r := splitLeadingServiceArgs(tt.args)
		if !strSliceEq(s, tt.services) {
			t.Errorf("splitLeadingServiceArgs(%v) services = %v, want %v", tt.args, s, tt.services)
		}
		if !strSliceEq(r, tt.rest) {
			t.Errorf("splitLeadingServiceArgs(%v) rest = %v, want %v", tt.args, r, tt.rest)
		}
	}
}
