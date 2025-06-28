package main

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"testing"
)

func writeCompose(dir, content string) {
	_ = os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(content), 0o644)
}

type DockerCompose struct {
	Services map[string]struct {
		Image string `yaml:"image"`
	} `yaml:"services"`
}

func TestGetImagesOfApp(t *testing.T) {
	tempDir := t.TempDir()
	writeCompose(tempDir, `version: "3"
services:
  nginx:
    image: nginx:1.0
  redis:
    image: redis:6.0
`)
	fsOperator := &SingleAppUpdateFileSystemOperatorImpl{}
	appServices, err := fsOperator.GetAppServices(tempDir)
	require.NoError(t, err)
	require.Len(t, appServices, 2)
	require.Equal(t, Service{Name: "nginx", Image: "nginx", Tag: "1.0"}, appServices[0])
	require.Equal(t, Service{Name: "redis", Image: "redis", Tag: "6.0"}, appServices[1])
}

func TestWriteNewTag(t *testing.T) {
	tempDir := t.TempDir()
	writeCompose(tempDir, `version: "3"
services:
  sample-app:
    image: nginx:1.0
`)
	fsOperator := &SingleAppUpdateFileSystemOperatorImpl{}
	require.NoError(t, fsOperator.WriteNewTagToTheDockerCompose(tempDir, "sample-app", "1.1"))

	newComposeContentBytes, err := os.ReadFile(filepath.Join(tempDir, "docker-compose-injected.yml"))
	require.NoError(t, err)

	var newDockerComposeContent DockerCompose
	require.NoError(t, yaml.Unmarshal(newComposeContentBytes, &newDockerComposeContent))
	require.Equal(t, "nginx:1.1", newDockerComposeContent.Services["sample-app"].Image)
}

func TestWriteNewTag_ServiceNotFound(t *testing.T) {
	tempDir := t.TempDir()
	writeCompose(tempDir, `version: "3"
services:
  other:
    image: nginx:1.0
`)
	fsOperator := &SingleAppUpdateFileSystemOperatorImpl{}
	err := fsOperator.WriteNewTagToTheDockerCompose(tempDir, "sample-app", "1.1")
	require.Error(t, err)
}
