package main

import (
	"errors"
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

func (f FileSystemOperatorImpl) GetAppServices(appDir string) ([]Service, error) {
	path := filepath.Join(appDir, "docker-compose.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Error("failed to read docker-compose.yml: %v", err)
		return nil, err
	}
	var doc map[string]interface{}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		logger.Error("failed to unmarshal docker-compose.yml: %v", err)
		return nil, err
	}
	svcMap, ok := doc["services"].(map[string]interface{})
	if !ok {
		logger.Error("services key missing in docker-compose.yml")
		return nil, errors.New("services key missing")
	}
	var services []Service
	for name, raw := range svcMap {
		svc, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		imageStr, ok := svc["image"].(string)
		if !ok {
			continue
		}
		parts := strings.SplitN(imageStr, ":", 2)
		image := parts[0]
		tag := ""
		if len(parts) == 2 {
			tag = parts[1]
		}
		services = append(services, Service{Name: name, Image: image, Tag: tag})
	}
	return services, nil
}

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
			logger.Info("app.yml not found in %s, using default port and path", appDir)
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
		logger.Error("failed to unmarshal docker-compose.yml: %v", err)
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
		logger.Error("failed to marshal updated docker-compose.yml: %v", err)
		return "", err
	}
	tmp, err := os.CreateTemp("", "compose-*.yml")
	if err != nil {
		logger.Error("failed to create temporary docker-compose file: %v", err)
		return "", err
	}
	defer tmp.Close()
	if _, err = tmp.Write(out); err != nil {
		logger.Error("failed to write to temporary docker-compose file: %v", err)
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
	data, err := os.ReadFile(filepath.Join(appDir, "docker-compose.yml"))
	if err != nil {
		logger.Error("failed to read docker-compose.yml: %v", err)
		return nil, fmt.Errorf("failed to read docker-compose.yml")
	}
	return data, nil
}

func (f FileSystemOperatorImpl) WriteDockerComposeFileContent(appDir string, content []byte) error {
	err := os.WriteFile(filepath.Join(appDir, "docker-compose.yml"), content, 0o600)
	if err != nil {
		logger.Error("failed to write docker-compose.yml: %v", err)
		return fmt.Errorf("failed to write docker-compose.yml")
	}
	return nil
}

func (f FileSystemOperatorImpl) WriteServiceUpdatesIntoComposeFile(appDir string, serviceUpdates []ServiceUpdate) error {
	path := filepath.Join(appDir, "docker-compose.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Error("failed to read docker-compose.yml: %v", err)
		return err
	}
	updated, err := UpdateComposeTags(data, serviceUpdates)
	if err != nil {
		logger.Error("failed to update docker-compose.yml with service updates: %v", err)
		return err
	}

	err = os.WriteFile(path, updated, 0600)
	if err != nil {
		logger.Error("failed to write updated docker-compose.yml: %v", err)
		return fmt.Errorf("failed to write updated docker-compose.yml")
	}
	return nil
}
