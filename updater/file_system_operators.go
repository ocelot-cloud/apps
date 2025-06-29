package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type DockerCompose struct {
	Services map[string]struct {
		Image string   `yaml:"image"`
		Ports []string `yaml:"ports,omitempty"`
	} `yaml:"services"`
}

type AppYaml struct {
	Port string `yaml:"port"`
	Path string `yaml:"path"`
}

type FileSystemOperatorImpl struct{}

func (f FileSystemOperatorImpl) GetListOfApps(appsDir string) ([]string, error) {
	ents, err := os.ReadDir(appsDir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

func (f FileSystemOperatorImpl) GetPortAndPathOfApp(appDir string) (string, string, error) {
	port, path := "80", "/"
	data, err := os.ReadFile(filepath.Join(appDir, "app.yml"))
	if err != nil {
		if os.IsNotExist(err) {
			return port, path, nil
		}
		return "", "", err
	}
	var c AppYaml
	if err := yaml.Unmarshal(data, &c); err != nil {
		return "", "", err
	}
	if c.Port != "" {
		port = c.Port
	}
	if c.Path != "" {
		path = c.Path
	}
	return port, path, nil
}

func (f FileSystemOperatorImpl) RunDockerCompose(appDir, port string) (string, error) {
	appName := filepath.Base(appDir)
	raw, err := f.GetDockerComposeFileContent(appDir)
	if err != nil {
		return "", err
	}
	var c map[string]any
	if err := yaml.Unmarshal(raw, &c); err != nil {
		return "", err
	}
	services, ok := c["services"].(map[string]any)
	if ok {
		for _, v := range services {
			svc, ok := v.(map[string]any)
			if !ok {
				continue
			}
			suffix := fmt.Sprintf("_%s_%s", appName, appName)
			if cn, ok := svc["container_name"].(string); ok && strings.HasSuffix(cn, suffix) {
				ports, _ := svc["ports"].([]any)
				svc["ports"] = append(ports, fmt.Sprintf("%s:%s", port, port))
				break
			}
		}
	}
	out, err := yaml.Marshal(&c)
	if err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp("", "compose-*.yml")
	if err != nil {
		return "", err
	}
	defer tmp.Close()
	if _, err = tmp.Write(out); err != nil {
		return "", err
	}
	cmd := exec.Command("docker", "compose", "-f", tmp.Name(), "up", "-d")
	if err := cmd.Run(); err != nil {
		logger.Error("failed to run docker compose up: %v", err)
		return "", err
	}
	return tmp.Name(), nil
}

func (f FileSystemOperatorImpl) ShutdownStackAndDeleteComposeFile(composeFilePath string) error {
	cmd := exec.Command("docker", "compose", "-f", composeFilePath, "down")
	if err := cmd.Run(); err != nil {
		logger.Error("failed to run docker compose down: %v", err)
		return fmt.Errorf("failed to run docker compose down: %v", err)
	}
	return os.RemoveAll(composeFilePath)
}

func (f FileSystemOperatorImpl) GetDockerComposeFileContent(appDir string) ([]byte, error) {
	return os.ReadFile(filepath.Join(appDir, "docker-compose.yml"))
}

func (f FileSystemOperatorImpl) WriteDockerComposeFileContent(appDir string, content []byte) error {
	return os.WriteFile(filepath.Join(appDir, "docker-compose.yml"), content, 0o600)
}

func (f FileSystemOperatorImpl) WriteServiceUpdatesIntoComposeFile(appDir string, serviceUpdates []ServiceUpdate) error {
	/* TODO implementation idea

	updatedDockerComposeContent, err := updateDockerComposeContent(originalDockerComposeContent, serviceUpdates)
	if err != nil {
		// TODO To be covered
		u.resetDockerComposeYamlToInitialContent(appDir, originalDockerComposeContent)
		return err
	}

	err = u.fileSystemOperator.WriteDockerComposeFileContent(appDir, updatedDockerComposeContent)
	if err != nil {
		u.resetDockerComposeYamlToInitialContent(appDir, originalDockerComposeContent)
		return err
	}
	return nil
	*/
	return nil
}

type SingleAppUpdateFileSystemOperatorImpl struct{}

func (s *SingleAppUpdateFileSystemOperatorImpl) GetAppServices(appDir string) ([]Service, error) {
	data, err := os.ReadFile(filepath.Join(appDir, "docker-compose.yml"))
	if err != nil {
		return nil, err
	}
	var c struct {
		Services map[string]struct {
			Image string `yaml:"image"`
		} `yaml:"services"`
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	res := make([]Service, 0, len(c.Services))
	for name, v := range c.Services {
		p := strings.SplitN(v.Image, ":", 2)
		tag := ""
		if len(p) == 2 {
			tag = p[1]
		}
		res = append(res, Service{Name: name, Image: p[0], Tag: tag})
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	return res, nil
}

func (s *SingleAppUpdateFileSystemOperatorImpl) WriteNewTagToTheDockerCompose(appDir, serviceName, newTag string) error {
	data, err := os.ReadFile(filepath.Join(appDir, "docker-compose.yml"))
	if err != nil {
		return err
	}
	var c struct {
		Version  any `yaml:"version,omitempty"`
		Services map[string]struct {
			Image string `yaml:"image"`
		} `yaml:"services"`
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return err
	}
	svc, ok := c.Services[serviceName]
	if !ok {
		return fmt.Errorf("service %s not found", serviceName)
	}
	base := strings.SplitN(svc.Image, ":", 2)[0]
	svc.Image = base + ":" + newTag
	c.Services[serviceName] = svc
	newData, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(appDir, "docker-compose-injected.yml"), newData, 0o600)
}
