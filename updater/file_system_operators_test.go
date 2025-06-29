//go:build integration

package main

import (
	tr "github.com/ocelot-cloud/task-runner"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func writeCompose(dir, content string) {
	_ = os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(content), 0o644)
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

func TestGetListOfApps(t *testing.T) {
	dir := t.TempDir()
	_ = os.Mkdir(filepath.Join(dir, "gitea"), 0o755)
	_ = os.Mkdir(filepath.Join(dir, "mattermost"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o644)
	f := FileSystemOperatorImpl{}
	out, err := f.GetListOfApps(dir)
	require.NoError(t, err)
	require.Equal(t, []string{"gitea", "mattermost"}, out)
}

func TestGetPortAndPathOfApp(t *testing.T) {
	dir := t.TempDir()
	f := FileSystemOperatorImpl{}
	p, pa, err := f.GetPortAndPathOfApp(dir)
	require.NoError(t, err)
	require.Equal(t, "80", p)
	require.Equal(t, "/", pa)

	_ = os.WriteFile(filepath.Join(dir, "app.yml"), []byte("port: \"8080\"\npath: /x"), 0o644)
	p, pa, err = f.GetPortAndPathOfApp(dir)
	require.NoError(t, err)
	require.Equal(t, "8080", p)
	require.Equal(t, "/x", pa)

	_ = os.WriteFile(filepath.Join(dir, "app.yml"), []byte("port: \"9090\""), 0o644)
	p, pa, err = f.GetPortAndPathOfApp(dir)
	require.NoError(t, err)
	require.Equal(t, "9090", p)
	require.Equal(t, "/", pa)
}

func TestRunComposeAndShutdown(t *testing.T) {
	cmd := exec.Command("docker", "rm", "-f", "test_nginx_nginx")
	defer cmd.Run()
	port := "12345"
	dir := t.TempDir()
	nginxDir := filepath.Join(dir, "nginx")
	_ = os.Mkdir(nginxDir, 0o755)
	_ = os.WriteFile(filepath.Join(nginxDir, "docker-compose.yml"), []byte(`version: "3"
services:
  nginx:
    container_name: test_nginx_nginx
    image: nginx:1.29.0-alpine
`), 0o600)
	fsOperator := FileSystemOperatorImpl{}
	pathToInjectedComposeYaml, err := fsOperator.RunDockerCompose(nginxDir, port)
	require.NoError(t, err)

	injectComposeYamlContentBytes, err := os.ReadFile(pathToInjectedComposeYaml)
	require.NoError(t, err)
	var injectComposeYamlContent DockerCompose
	require.NoError(t, yaml.Unmarshal(injectComposeYamlContentBytes, &injectComposeYamlContent))
	tr.WaitUntilPortIsReady(port)
	require.Contains(t, injectComposeYamlContent.Services["nginx"].Ports, port+":"+port)

	require.NoError(t, fsOperator.ShutdownStackAndDeleteComposeFile(pathToInjectedComposeYaml))
	_, err = os.Stat(pathToInjectedComposeYaml)
	require.Error(t, err)
}
