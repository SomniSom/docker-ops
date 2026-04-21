package composeimage

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const sample = `services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - redis
  redis:
    image: redis:7-alpine
  db:
    image: postgres:16
`

func TestGenerateForArtifacts_happy(t *testing.T) {
	out, err := GenerateForArtifacts([]byte(sample), Options{
		TargetService: "app",
		ImageExpr:     "${DEPLOY_IMAGE}",
	})
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]interface{}
	if err := yaml.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	svcs := doc["services"].(map[string]interface{})
	app := svcs["app"].(map[string]interface{})
	if _, ok := app["build"]; ok {
		t.Fatal("expected build removed from app")
	}
	if app["image"] != "${DEPLOY_IMAGE}" {
		t.Fatalf("image = %v", app["image"])
	}
	redis := svcs["redis"].(map[string]interface{})
	if redis["image"] != "redis:7-alpine" {
		t.Fatalf("redis unchanged: %v", redis)
	}
}

func TestGenerateForArtifacts_otherBuildRejected(t *testing.T) {
	yml := `services:
  app:
    build: .
  worker:
    build: ./worker
`
	_, err := GenerateForArtifacts([]byte(yml), Options{
		TargetService: "app",
	})
	if err == nil || !strings.Contains(err.Error(), "worker") {
		t.Fatalf("expected error about worker, got %v", err)
	}
}

func TestGenerateForArtifacts_serviceImages(t *testing.T) {
	yml := `services:
  app:
    build: .
  worker:
    build: ./worker
`
	out, err := GenerateForArtifacts([]byte(yml), Options{
		AllBuilt: true,
		ServiceImages: map[string]string{
			"app":    "registry/app:v1",
			"worker": "registry/worker:v1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]interface{}
	if err := yaml.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	svcs := doc["services"].(map[string]interface{})
	app := svcs["app"].(map[string]interface{})
	if app["image"] != "registry/app:v1" {
		t.Fatalf("app image %v", app["image"])
	}
	w := svcs["worker"].(map[string]interface{})
	if w["image"] != "registry/worker:v1" {
		t.Fatalf("worker image %v", w["image"])
	}
}

func TestGenerateForArtifacts_allBuilt(t *testing.T) {
	yml := `services:
  app:
    build: .
  worker:
    build: ./worker
  redis:
    image: redis:7
`
	out, err := GenerateForArtifacts([]byte(yml), Options{AllBuilt: true, ImageExpr: "${DEPLOY_IMAGE}"})
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]interface{}
	if err := yaml.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	svcs := doc["services"].(map[string]interface{})
	for _, name := range []string{"app", "worker"} {
		s := svcs[name].(map[string]interface{})
		if _, ok := s["build"]; ok {
			t.Fatalf("%s still has build", name)
		}
		if s["image"] != "${DEPLOY_IMAGE}" {
			t.Fatalf("%s image %v", name, s["image"])
		}
	}
}

func TestGenerateForArtifacts_noBuildOnTarget(t *testing.T) {
	yml := `services:
  app:
    image: nginx:latest
`
	_, err := GenerateForArtifacts([]byte(yml), Options{
		TargetService: "app",
	})
	if err == nil || !strings.Contains(err.Error(), "no build") {
		t.Fatalf("got %v", err)
	}
}
