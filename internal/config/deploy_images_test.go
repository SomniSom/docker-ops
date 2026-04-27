package config

import "testing"

func TestParseDeployImagesList(t *testing.T) {
	m := ParseDeployImagesList("app=ghcr.io/x:1,worker=ghcr.io/y:2")
	if len(m) != 2 || m["app"] != "ghcr.io/x:1" || m["worker"] != "ghcr.io/y:2" {
		t.Fatalf("got %#v", m)
	}
	if ParseDeployImagesList("") != nil || ParseDeployImagesList("  ") != nil {
		t.Fatal("want nil for empty")
	}
}

func TestConfig_HasDeployImageOrImages(t *testing.T) {
	if (&Config{DeployImages: map[string]string{"a": "t:1"}}).HasDeployImageOrImages() != true {
		t.Fatal("want true from deploy_images")
	}
	if (&Config{DeployImages: map[string]string{"a": ""}}).HasDeployImageOrImages() != false {
		t.Fatal("want false when only empty tags")
	}
	if (&Config{DeployImage: "x:1"}).HasDeployImageOrImages() != true {
		t.Fatal("want true from deploy_image")
	}
}
