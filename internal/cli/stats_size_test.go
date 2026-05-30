package cli

import (
	"strings"
	"testing"
)

func TestParseComposeStatsJSON(t *testing.T) {
	data := []byte(`{"ID":"abc123","Name":"app","CPUPerc":"1%","MemUsage":"10MiB / 1GiB","MemPerc":"1%","NetIO":"1kB / 2kB","BlockIO":"0B / 0B","PIDs":"3"}
`)
	rows, err := parseComposeStatsJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Name != "app" {
		t.Fatalf("rows=%+v", rows)
	}
}

func TestLookupDiskSize(t *testing.T) {
	sizes := map[string]string{
		"abc123def456": "0B (virtual 100MB)",
	}
	got := lookupDiskSize(sizes, "abc123")
	if got != "0B (virtual 100MB)" {
		t.Fatalf("got %q", got)
	}
	if lookupDiskSize(sizes, "missing") != "-" {
		t.Fatal("expected dash for missing")
	}
}

func TestParseDockerPsSizeRows(t *testing.T) {
	data := []byte(`{"ID":"abc","Names":"/my-app","Size":"1B (virtual 2MB)"}
`)
	rows := parseDockerPsSizeRows(data)
	if len(rows) != 1 || rows[0].Names != "my-app" || !strings.Contains(rows[0].Size, "virtual") {
		t.Fatalf("rows=%+v", rows)
	}
}

func TestStatsUsesCustomFormat(t *testing.T) {
	if !statsUsesCustomFormat([]string{"--format", "json"}) {
		t.Fatal("expected custom format")
	}
	if statsUsesCustomFormat([]string{"--no-stream"}) {
		t.Fatal("expected default format")
	}
}
