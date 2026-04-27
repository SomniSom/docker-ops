package version

import (
	"testing"
)

func TestShowModuleVersionNote(t *testing.T) {
	savedV, savedC := Version, Commit
	defer func() { Version, Commit = savedV, savedC }()
	Commit = "abc"

	Version = "v1.1.0"
	if ShowModuleVersionNote() {
		t.Fatal("plain release tag should not show module note")
	}
	Version = "v1.0.1-0.20260407201612-08053888e894"
	if !ShowModuleVersionNote() {
		t.Fatal("pseudo-version from go install should show module note")
	}
	Version = "dev"
	if ShowModuleVersionNote() {
		t.Fatal("dev should not show module note")
	}
}

func TestVersionFieldsNeverUnknown(t *testing.T) {
	if Version == "unknown" || Version == "" {
		t.Fatalf("Version = %q", Version)
	}
	if Commit == "unknown" || Commit == "" {
		t.Fatalf("Commit = %q", Commit)
	}
}
