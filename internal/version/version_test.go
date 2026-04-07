package version

import "testing"

func TestVersionFieldsNeverUnknown(t *testing.T) {
	if Version == "unknown" || Version == "" {
		t.Fatalf("Version = %q", Version)
	}
	if Commit == "unknown" || Commit == "" {
		t.Fatalf("Commit = %q", Commit)
	}
}
