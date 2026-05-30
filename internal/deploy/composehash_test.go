package deploy

import (
	"testing"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/compose-spec/compose-go/v2/types"
)

func TestComposeConfigFilesLabelRemote(t *testing.T) {
	cfg := &config.Config{
		RemoteSSH:  "user@host",
		RemotePath: "/opt/downloadbot",
	}
	got := composeConfigFilesLabel(cfg, "/home/proj", "docker-compose.image.yml")
	want := "/opt/downloadbot/docker-compose.image.yml"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestServiceConfigHashStable(t *testing.T) {
	svc := types.ServiceConfig{
		Image:   "myapp:1",
		Restart: "unless-stopped",
	}
	h1, err := serviceConfigHash(svc)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := serviceConfigHash(svc)
	if err != nil {
		t.Fatal(err)
	}
	if h1 == "" || h1 != h2 {
		t.Fatalf("hash=%q want stable non-empty", h1)
	}
}
