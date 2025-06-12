package main

import (
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestUpdateCommand builds the updater binary and runs an update on sampleapp.
// It verifies that the docker-compose.yml file in apps/test is updated with a
// bumped tag. The file is restored after the test.
func TestUpdateCommand(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed")
	}
	dir, _ := os.Getwd()
	bin := filepath.Join(dir, "updater_bin")
	build := exec.Command("go", "build", "-o", bin)
	build.Dir = dir
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, string(out))
	}
	composePath := filepath.Join("..", "apps", "test", "sampleapp", "docker-compose.yml")
	data, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("read compose: %v", err)
	}
	defer os.WriteFile(composePath, data, 0644)

	var compose map[string]any
	if err := yaml.Unmarshal(data, &compose); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	svc := compose["services"].(map[string]any)["sampleapp"].(map[string]any)
	img := svc["image"].(string)
	before := imageTag(img)

	cmd := exec.Command(bin, "update", "sampleapp", "-p", "../apps/test")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("update failed: %v\n%s", err, string(out))
	}

	data, err = os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("read compose after: %v", err)
	}
	if err := yaml.Unmarshal(data, &compose); err != nil {
		t.Fatalf("unmarshal after: %v", err)
	}
	svc = compose["services"].(map[string]any)["sampleapp"].(map[string]any)
	img = svc["image"].(string)
	after := imageTag(img)
	expected := bumpTag(before)
	if after != expected {
		t.Fatalf("expected tag %s, got %s", expected, after)
	}
}
