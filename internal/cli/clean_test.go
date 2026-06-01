package cli

import "testing"

func TestBuildSystemPruneArgs(t *testing.T) {
	got := buildSystemPruneArgs(false, false)
	want := []string{"system", "prune", "-f"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}

	got = buildSystemPruneArgs(true, true)
	if got[len(got)-2] != "-a" || got[len(got)-1] != "--volumes" {
		t.Fatalf("got %v", got)
	}
}
