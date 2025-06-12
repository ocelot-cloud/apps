package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

//go:generate mockery --name Runner --output ./mocks --filename runner_mock.go
type Runner interface {
	Run(dir, command string) error
}

//go:generate mockery --name Waiter --output ./mocks --filename waiter_mock.go
type Waiter interface {
	WaitPort(port string) error
	WaitWeb(url string) error
}

type CmdRunner struct{}

func (CmdRunner) Run(dir, command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type DefaultWaiter struct {
	Timeout time.Duration
}

func (w DefaultWaiter) WaitPort(port string) error {
	deadline := time.Now().Add(w.Timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("port %s not ready", port)
}

func (w DefaultWaiter) WaitWeb(url string) error {
	deadline := time.Now().Add(w.Timeout)
	client := http.Client{Timeout: time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode < 500 {
			resp.Body.Close()
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("url %s not ready", url)
}

type AppManager struct {
	AppsDir string
	Runner  Runner
	Waiter  Waiter
}

func (m AppManager) ListApps() ([]string, error) {
	entries, err := os.ReadDir(m.AppsDir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func ReadAppPort(appDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(appDir, "app.yml"))
	if err != nil {
		return "", err
	}
	var obj map[string]any
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return "", err
	}
	if p, ok := obj["port"].(int); ok {
		return fmt.Sprintf("%d", p), nil
	}
	if p, ok := obj["port"].(string); ok {
		return p, nil
	}
	if _, ok := obj["port"]; !ok {
		return "80", nil
	}
	return "", errors.New("port not found")
}

func injectPort(composePath, service, port string) (string, error) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return "", err
	}
	var compose map[string]any
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return "", err
	}
	services, ok := compose["services"].(map[string]any)
	if !ok {
		return "", errors.New("compose has no services")
	}
	svc, ok := services[service].(map[string]any)
	if !ok {
		for _, v := range services {
			if s, ok := v.(map[string]any); ok {
				svc = s
				break
			}
		}
	}
	if svc == nil {
		return "", errors.New("service not found")
	}
	ports, _ := svc["ports"].([]any)
	mapping := fmt.Sprintf("%s:%s", port, port)
	for _, p := range ports {
		if ps, ok := p.(string); ok && strings.HasPrefix(ps, port+":") {
			return "", fmt.Errorf("port %s already mapped", port)
		}
	}
	ports = append(ports, mapping)
	svc["ports"] = ports
	services[service] = svc
	compose["services"] = services

	out, err := yaml.Marshal(compose)
	if err != nil {
		return "", err
	}
	temp := composePath + ".temp"
	if err := os.WriteFile(temp, out, fs.FileMode(0644)); err != nil {
		return "", err
	}
	return temp, nil
}

func imageTag(image string) string {
	parts := strings.Split(image, ":")
	if len(parts) < 2 {
		return ""
	}
	tag := parts[1]
	if i := strings.Index(tag, "-"); i != -1 {
		tag = tag[:i]
	}
	return tag
}

func bumpTag(tag string) string {
	segments := strings.Split(tag, ".")
	for i, s := range segments {
		if n, err := strconv.Atoi(s); err == nil {
			n++
			segments[i] = fmt.Sprintf("%d", n)
		}
	}
	return strings.Join(segments, ".")
}

type HealthCheckResult struct {
	App     string
	Success bool
	Error   error
}

func (m AppManager) HealthCheck(appNames []string) ([]HealthCheckResult, error) {
	if len(appNames) == 0 {
		var err error
		appNames, err = m.ListApps()
		if err != nil {
			return nil, err
		}
	}
	var results []HealthCheckResult
	for _, name := range appNames {
		appDir := filepath.Join(m.AppsDir, name)
		port, err := ReadAppPort(appDir)
		if err != nil {
			results = append(results, HealthCheckResult{App: name, Success: false, Error: err})
			continue
		}
		composePath := filepath.Join(appDir, "docker-compose.yml")
		temp, err := injectPort(composePath, name, port)
		if err != nil {
			results = append(results, HealthCheckResult{App: name, Success: false, Error: err})
			continue
		}
		up := fmt.Sprintf("docker compose -f %s up -d", filepath.Base(temp))
		if err := m.Runner.Run(appDir, up); err != nil {
			os.Remove(temp)
			results = append(results, HealthCheckResult{App: name, Success: false, Error: err})
			continue
		}
		if err := m.Waiter.WaitPort(port); err != nil {
			m.Runner.Run(appDir, fmt.Sprintf("docker compose -f %s down -v", filepath.Base(temp)))
			os.Remove(temp)
			results = append(results, HealthCheckResult{App: name, Success: false, Error: err})
			continue
		}
		if err := m.Waiter.WaitWeb("http://localhost:" + port); err != nil {
			m.Runner.Run(appDir, fmt.Sprintf("docker compose -f %s down -v", filepath.Base(temp)))
			os.Remove(temp)
			results = append(results, HealthCheckResult{App: name, Success: false, Error: err})
			continue
		}
		results = append(results, HealthCheckResult{App: name, Success: true})
		m.Runner.Run(appDir, fmt.Sprintf("docker compose -f %s down -v", filepath.Base(temp)))
		os.Remove(temp)
	}
	return results, nil
}

func (m AppManager) Update(appNames []string) ([]UpdateResult, error) {
	if len(appNames) == 0 {
		var err error
		appNames, err = m.ListApps()
		if err != nil {
			return nil, err
		}
	}
	var results []UpdateResult
	for _, name := range appNames {
		appDir := filepath.Join(m.AppsDir, name)
		composePath := filepath.Join(appDir, "docker-compose.yml")
		data, err := os.ReadFile(composePath)
		if err != nil {
			results = append(results, UpdateResult{App: name, Success: false, Error: err})
			continue
		}
		var compose map[string]any
		if err := yaml.Unmarshal(data, &compose); err != nil {
			results = append(results, UpdateResult{App: name, Success: false, Error: err})
			continue
		}
		servicesYaml, _ := compose["services"].(map[string]any)
		var svcUpdates []ServiceUpdate
		for svcName, v := range servicesYaml {
			m := v.(map[string]any)
			img, _ := m["image"].(string)
			before := imageTag(img)
			after := bumpTag(before)
			svcUpdates = append(svcUpdates, ServiceUpdate{Service: svcName, Before: before, After: after})
		}
		pull := "docker compose pull"
		if err := m.Runner.Run(appDir, pull); err != nil {
			results = append(results, UpdateResult{App: name, Success: false, Error: err, Services: svcUpdates})
			continue
		}
		res, err := m.HealthCheck([]string{name})
		if err != nil {
			m.Runner.Run(appDir, "git checkout -- docker-compose.yml")
			results = append(results, UpdateResult{App: name, Success: false, Error: err, Services: svcUpdates})
			continue
		}
		if len(res) == 0 || !res[0].Success {
			m.Runner.Run(appDir, "git checkout -- docker-compose.yml")
			var herr error
			if len(res) > 0 {
				herr = res[0].Error
			}
			results = append(results, UpdateResult{App: name, Success: false, Error: herr, Services: svcUpdates})
			continue
		}
		results = append(results, UpdateResult{App: name, Success: true, Services: svcUpdates})
	}
	return results, nil
}

type ServiceUpdate struct {
	Service string
	Before  string
	After   string
}

type UpdateResult struct {
	App      string
	Success  bool
	Error    error
	Services []ServiceUpdate
}
