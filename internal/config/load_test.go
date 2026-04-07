package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_NoFiles_UsesDefaults(t *testing.T) {
	dir := t.TempDir()
	res, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	c := res.Config
	if c.ComposeProjectName != filepath.Base(dir) {
		t.Fatalf("ComposeProjectName: got %q want base %q", c.ComposeProjectName, filepath.Base(dir))
	}
	if c.ComposeFile != "docker-compose.yml" {
		t.Fatalf("ComposeFile: %q", c.ComposeFile)
	}
	if c.ComposeService != "app" {
		t.Fatalf("ComposeService: %q", c.ComposeService)
	}
	if c.RemoteConfigured() {
		t.Fatal("remote should not be configured")
	}
}

func TestLoad_YAMLThenDQEnvPriority(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "docker-ops.yaml"), []byte(`
compose_project_name: from-yaml
remote_ssh: yaml-user@host
remote_path: /yaml/path
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dq.env"), []byte(`
REMOTE_SSH=env-user@host
# empty should not override
COMPOSE_PROJECT_NAME=
`), 0o600); err != nil {
		t.Fatal(err)
	}

	res, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	c := res.Config
	if c.ComposeProjectName != "from-yaml" {
		t.Fatalf("empty dq.env must not clear YAML: got %q", c.ComposeProjectName)
	}
	if c.RemoteSSH != "env-user@host" {
		t.Fatalf("dq.env should override yaml remote_ssh: got %q", c.RemoteSSH)
	}
	if c.RemotePath != "/yaml/path" {
		t.Fatalf("RemotePath: %q", c.RemotePath)
	}
}

func TestLoad_ProcessEnvHighest(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "docker-ops.yaml"), []byte(`remote_ssh: a@b`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dq.env"), []byte(`REMOTE_SSH=c@d`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REMOTE_SSH", "proc@host")

	res, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if res.Config.RemoteSSH != "proc@host" {
		t.Fatalf("process env should win: got %q", res.Config.RemoteSSH)
	}
}

func TestRemoteConfigured_UseRemoteFalse(t *testing.T) {
	c := &Config{RemoteSSH: "u@h", RemotePath: "/p"}
	f := false
	c.UseRemote = &f
	if c.RemoteConfigured() {
		t.Fatal("UseRemote false disables remote")
	}
}

func TestParseDQEnv_InvalidLine(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "dq.env")
	if err := os.WriteFile(p, []byte("not_a_valid_line\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ParseDQEnv(p)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoteYAMLTemplate_Anonymize(t *testing.T) {
	cfg := &Config{
		ComposeProjectName: "myapp",
		RemoteSSH:          "secret@host",
		RemotePath:         "/secret",
		ComposeFile:        "dc.yml",
		ComposeService:     "svc",
	}
	out := RemoteYAMLTemplate(cfg, true)
	if strings.Contains(out, "secret@host") || strings.Contains(out, "/secret") {
		t.Fatalf("anonymize leaked remote: %s", out)
	}
	if !strings.Contains(out, "user@host") || !strings.Contains(out, "/opt/myapp") {
		t.Fatalf("expected placeholders: %s", out)
	}
}
