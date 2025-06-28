package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileSystemOperatorImpl struct{}

func (f FileSystemOperatorImpl) GetListOfApps(appsDir string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (f FileSystemOperatorImpl) GetPortAndPathOfApp(appDir string) (string, string, error) {
	//TODO implement me
	panic("implement me")
}

// TODO isnt there a port argument missing?
// TODO !! I could simply merge GetPortAndPathOfApp, InjectPortInDockerCompose, RunInjectedDockerCompose
func (f FileSystemOperatorImpl) InjectPortInDockerCompose(appDir string) error {
	//TODO implement me
	panic("implement me")
}

func (f FileSystemOperatorImpl) RunInjectedDockerCompose(appDir string) error {
	//TODO implement me
	panic("implement me")
}

// TODO use this in business logic
func (f FileSystemOperatorImpl) ShutdownStackAndDeleteComposeFile(appDir string) error {
	//TODO implement me
	panic("implement me")
}

func (f FileSystemOperatorImpl) GetDockerComposeFileContent(appDir string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (f FileSystemOperatorImpl) WriteDockerComposeFileContent(appDir string, content []byte) error {
	//TODO implement me
	panic("implement me")
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
	return os.WriteFile(filepath.Join(appDir, "docker-compose-injected.yml"), newData, 0o644)
}
