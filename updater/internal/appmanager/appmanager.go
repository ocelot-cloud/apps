package appmanager

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Runner executes shell commands in a directory.
type Runner interface {
	Run(dir, command string) error
}

// Waiter waits for ports and URLs to become ready.
type Waiter interface {
	WaitPort(port string) error
	WaitWeb(url string) error
}

// CmdRunner is a Runner implementation using os/exec.
type CmdRunner struct{}

func (CmdRunner) Run(dir, command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// DefaultWaiter waits until network endpoints are reachable.
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

// AppManager runs health checks and updates for apps located in AppsDir.
type AppManager struct {
	AppsDir string
	Runner  Runner
	Waiter  Waiter
}

// ListApps returns available apps in AppsDir.
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

// ReadAppPort reads the exposed port from app.yml.
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
	return "", errors.New("port not found")
}

// injectPort injects a port mapping into a compose file and writes a temp file.
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
		// if service key differs, use first service
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
	found := false
	for _, p := range ports {
		if ps, ok := p.(string); ok && strings.HasPrefix(ps, port+":") {
			found = true
		}
	}
	if !found {
		ports = append(ports, mapping)
		svc["ports"] = ports
	}
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

// HealthCheck runs health checks for the given apps. If appNames is empty, all apps are checked.
func (m AppManager) HealthCheck(appNames []string) ([]string, error) {
	if len(appNames) == 0 {
		var err error
		appNames, err = m.ListApps()
		if err != nil {
			return nil, err
		}
	}
	var healthy []string
	for _, name := range appNames {
		appDir := filepath.Join(m.AppsDir, name)
		port, err := ReadAppPort(appDir)
		if err != nil {
			continue
		}
		composePath := filepath.Join(appDir, "docker-compose.yml")
		temp, err := injectPort(composePath, name, port)
		if err != nil {
			return healthy, err
		}
		up := fmt.Sprintf("docker compose -f %s up -d", filepath.Base(temp))
		if err := m.Runner.Run(appDir, up); err != nil {
			os.Remove(temp)
			return healthy, err
		}
		if err := m.Waiter.WaitPort(port); err != nil {
			m.Runner.Run(appDir, fmt.Sprintf("docker compose -f %s down -v", filepath.Base(temp)))
			os.Remove(temp)
			return healthy, err
		}
		if err := m.Waiter.WaitWeb("http://localhost:" + port); err != nil {
			m.Runner.Run(appDir, fmt.Sprintf("docker compose -f %s down -v", filepath.Base(temp)))
			os.Remove(temp)
			return healthy, err
		}
		healthy = append(healthy, name)
		m.Runner.Run(appDir, fmt.Sprintf("docker compose -f %s down -v", filepath.Base(temp)))
		os.Remove(temp)
	}
	return healthy, nil
}

// Update updates docker images and verifies stability.
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
		pull := "docker compose pull"
		if err := m.Runner.Run(appDir, pull); err != nil {
			results = append(results, UpdateResult{App: name, Success: false, Error: err})
			continue
		}
		_, err := m.HealthCheck([]string{name})
		if err != nil {
			m.Runner.Run(appDir, "git checkout -- docker-compose.yml")
			results = append(results, UpdateResult{App: name, Success: false, Error: err})
			continue
		}
		results = append(results, UpdateResult{App: name, Success: true})
	}
	return results, nil
}

type UpdateResult struct {
	App     string
	Success bool
	Error   error
}
