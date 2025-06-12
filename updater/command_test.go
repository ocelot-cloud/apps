package main

import (
	"gopkg.in/yaml.v3"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type RunnerMock struct{ mock.Mock }

func (m *RunnerMock) Run(dir, command string) error {
	args := m.Called(dir, command)
	return args.Error(0)
}

type WaiterMock struct{ mock.Mock }

func (m *WaiterMock) WaitPort(port string) error {
	args := m.Called(port)
	return args.Error(0)
}

func (m *WaiterMock) WaitWeb(url string) error {
	args := m.Called(url)
	return args.Error(0)
}

func TestReadAppPort(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "app.yml"), []byte("port: 1234"), 0644)
	p, err := ReadAppPort(dir)
	assert.NoError(t, err)
	assert.Equal(t, "1234", p)
}

func TestInjectPort(t *testing.T) {
	dir := t.TempDir()
	compose := `services:
  test:
    image: nginx
`
	path := filepath.Join(dir, "docker-compose.yml")
	os.WriteFile(path, []byte(compose), 0644)
	tmp, err := injectPort(path, "test", "8080")
	assert.NoError(t, err)
	data, err := os.ReadFile(tmp)
	assert.NoError(t, err)
	var obj map[string]any
	err = yaml.Unmarshal(data, &obj)
	assert.NoError(t, err)
	services := obj["services"].(map[string]any)
	svc := services["test"].(map[string]any)
	ports := svc["ports"].([]any)
	assert.Equal(t, 1, len(ports))
	assert.Equal(t, "8080:8080", ports[0])
}

func TestHealthCheck(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "app1")
	os.Mkdir(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "app.yml"), []byte("port: 80"), 0644)
	os.WriteFile(filepath.Join(appDir, "docker-compose.yml"), []byte("services:\n  app1:\n    image: nginx"), 0644)

	runner := new(RunnerMock)
	waiter := new(WaiterMock)
	m := AppManager{AppsDir: dir, Runner: runner, Waiter: waiter}

	runner.On("Run", appDir, mock.Anything).Return(nil)
	waiter.On("WaitPort", "80").Return(nil)
	waiter.On("WaitWeb", "http://localhost:80").Return(nil)

	results, err := m.HealthCheck([]string{"app1"})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.True(t, results[0].Success)
	assert.Equal(t, "app1", results[0].App)
	runner.AssertExpectations(t)
	waiter.AssertExpectations(t)
}

func TestUpdate(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "app1")
	os.Mkdir(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "app.yml"), []byte("port: 80"), 0644)
	os.WriteFile(filepath.Join(appDir, "docker-compose.yml"), []byte("services:\n  app1:\n    image: nginx"), 0644)

	runner := new(RunnerMock)
	waiter := new(WaiterMock)
	m := AppManager{AppsDir: dir, Runner: runner, Waiter: waiter}

	runner.On("Run", appDir, "docker compose pull").Return(nil)
	runner.On("Run", appDir, mock.Anything).Return(nil)
	waiter.On("WaitPort", "80").Return(nil)
	waiter.On("WaitWeb", "http://localhost:80").Return(nil)

	res, err := m.Update([]string{"app1"})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.True(t, res[0].Success)
	assert.Len(t, res[0].Services, 1)
}

func TestUpdateFailure(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "app1")
	os.Mkdir(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "app.yml"), []byte("port: 80"), 0644)
	os.WriteFile(filepath.Join(appDir, "docker-compose.yml"), []byte("services:\n  app1:\n    image: nginx"), 0644)

	runner := new(RunnerMock)
	waiter := new(WaiterMock)
	m := AppManager{AppsDir: dir, Runner: runner, Waiter: waiter}

	runner.On("Run", appDir, "docker compose pull").Return(nil)
	runner.On("Run", appDir, mock.Anything).Return(nil)
	waiter.On("WaitPort", "80").Return(assert.AnError)

	res, err := m.Update([]string{"app1"})
	assert.NoError(t, err)
	assert.False(t, res[0].Success)
	assert.Len(t, res[0].Services, 1)
}

func TestReadAppPortMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := ReadAppPort(dir)
	assert.Error(t, err)
}

func TestReadAppPortDefault(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "app.yml"), []byte("name: test"), 0644)
	p, err := ReadAppPort(dir)
	assert.NoError(t, err)
	assert.Equal(t, "80", p)
}

func TestListApps(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "a"), 0755)
	os.Mkdir(filepath.Join(dir, "b"), 0755)
	os.WriteFile(filepath.Join(dir, "file"), []byte("x"), 0644)

	m := AppManager{AppsDir: dir}
	apps, err := m.ListApps()
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"a", "b"}, apps)
}

func TestCmdRunnerRun(t *testing.T) {
	dir := t.TempDir()
	r := CmdRunner{}
	err := r.Run(dir, "sh -c 'echo hi > out'")
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(dir, "out"))
	assert.NoError(t, err)
}

func TestDefaultWaiter(t *testing.T) {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	port := strings.Split(ln.Addr().String(), ":")[1]
	w := DefaultWaiter{Timeout: time.Second}
	assert.NoError(t, w.WaitPort(port))
	srv := httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, r *http.Request) { wr.WriteHeader(200) }))
	w.Timeout = time.Second
	assert.NoError(t, w.WaitWeb(srv.URL))
	ln.Close()
	srv.Close()
	w.Timeout = time.Millisecond
	assert.Error(t, w.WaitPort(port))
	assert.Error(t, w.WaitWeb(srv.URL))
}

func TestInjectPortServiceMismatch(t *testing.T) {
	dir := t.TempDir()
	compose := `services:
  web:
    image: nginx
`
	path := filepath.Join(dir, "docker-compose.yml")
	os.WriteFile(path, []byte(compose), 0644)
	tmp, err := injectPort(path, "app", "8080")
	assert.NoError(t, err)
	data, _ := os.ReadFile(tmp)
	var obj map[string]any
	err = yaml.Unmarshal(data, &obj)
	assert.NoError(t, err)
	services := obj["services"].(map[string]any)
	svc := services["web"].(map[string]any)
	ports := svc["ports"].([]any)
	assert.Equal(t, 1, len(ports))
	assert.Equal(t, "8080:8080", ports[0])
}

func TestInjectPortNoDuplicate(t *testing.T) {
	dir := t.TempDir()
	compose := `services:
  web:
    image: nginx
    ports:
      - "8080:8080"
`
	path := filepath.Join(dir, "docker-compose.yml")
	os.WriteFile(path, []byte(compose), 0644)
	_, err := injectPort(path, "web", "8080")
	assert.Error(t, err)
}

func TestHealthCheckAllApps(t *testing.T) {
	dir := t.TempDir()
	appDir1 := filepath.Join(dir, "app1")
	os.Mkdir(appDir1, 0755)
	os.WriteFile(filepath.Join(appDir1, "app.yml"), []byte("port: 80"), 0644)
	os.WriteFile(filepath.Join(appDir1, "docker-compose.yml"), []byte("services:\n  app1:\n    image: nginx"), 0644)

	appDir2 := filepath.Join(dir, "app2")
	os.Mkdir(appDir2, 0755)
	os.WriteFile(filepath.Join(appDir2, "app.yml"), []byte("port: 81"), 0644)
	os.WriteFile(filepath.Join(appDir2, "docker-compose.yml"), []byte("services:\n  app2:\n    image: nginx"), 0644)

	runner := new(RunnerMock)
	waiter := new(WaiterMock)
	m := AppManager{AppsDir: dir, Runner: runner, Waiter: waiter}

	runner.On("Run", appDir1, mock.Anything).Return(nil)
	runner.On("Run", appDir2, mock.Anything).Return(nil)
	waiter.On("WaitPort", "80").Return(nil)
	waiter.On("WaitPort", "81").Return(nil)
	waiter.On("WaitWeb", "http://localhost:80").Return(nil)
	waiter.On("WaitWeb", "http://localhost:81").Return(nil)

	results, err := m.HealthCheck(nil)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.True(t, results[1].Success)
}

func TestUpdateAllApps(t *testing.T) {
	dir := t.TempDir()
	appDir1 := filepath.Join(dir, "app1")
	os.Mkdir(appDir1, 0755)
	os.WriteFile(filepath.Join(appDir1, "app.yml"), []byte("port: 80"), 0644)
	os.WriteFile(filepath.Join(appDir1, "docker-compose.yml"), []byte("services:\n  app1:\n    image: nginx"), 0644)

	appDir2 := filepath.Join(dir, "app2")
	os.Mkdir(appDir2, 0755)
	os.WriteFile(filepath.Join(appDir2, "app.yml"), []byte("port: 81"), 0644)
	os.WriteFile(filepath.Join(appDir2, "docker-compose.yml"), []byte("services:\n  app2:\n    image: nginx"), 0644)

	runner := new(RunnerMock)
	waiter := new(WaiterMock)
	m := AppManager{AppsDir: dir, Runner: runner, Waiter: waiter}

	runner.On("Run", appDir1, "docker compose pull").Return(nil)
	runner.On("Run", appDir2, "docker compose pull").Return(nil)
	runner.On("Run", appDir1, mock.Anything).Return(nil)
	runner.On("Run", appDir2, mock.Anything).Return(nil)
	waiter.On("WaitPort", "80").Return(nil)
	waiter.On("WaitPort", "81").Return(nil)
	waiter.On("WaitWeb", "http://localhost:80").Return(nil)
	waiter.On("WaitWeb", "http://localhost:81").Return(nil)

	res, err := m.Update(nil)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.True(t, res[0].Success)
	assert.True(t, res[1].Success)
	assert.Len(t, res[0].Services, 1)
	assert.Len(t, res[1].Services, 1)
}

func TestUpdatePartialFailure(t *testing.T) {
	dir := t.TempDir()
	appDir1 := filepath.Join(dir, "app1")
	os.Mkdir(appDir1, 0755)
	os.WriteFile(filepath.Join(appDir1, "app.yml"), []byte("port: 80"), 0644)
	os.WriteFile(filepath.Join(appDir1, "docker-compose.yml"), []byte("services:\n  app1:\n    image: nginx"), 0644)

	appDir2 := filepath.Join(dir, "app2")
	os.Mkdir(appDir2, 0755)
	os.WriteFile(filepath.Join(appDir2, "app.yml"), []byte("port: 81"), 0644)
	os.WriteFile(filepath.Join(appDir2, "docker-compose.yml"), []byte("services:\n  app2:\n    image: nginx"), 0644)

	runner := new(RunnerMock)
	waiter := new(WaiterMock)
	m := AppManager{AppsDir: dir, Runner: runner, Waiter: waiter}

	runner.On("Run", appDir1, "docker compose pull").Return(nil)
	runner.On("Run", appDir2, "docker compose pull").Return(nil)
	runner.On("Run", appDir1, mock.Anything).Return(nil)
	runner.On("Run", appDir2, mock.Anything).Return(nil)
	waiter.On("WaitPort", "80").Return(assert.AnError)
	waiter.On("WaitPort", "81").Return(nil)
	waiter.On("WaitWeb", "http://localhost:81").Return(nil)

	res, err := m.Update([]string{"app1", "app2"})
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.False(t, res[0].Success)
	assert.True(t, res[1].Success)
	assert.Len(t, res[0].Services, 1)
	assert.Len(t, res[1].Services, 1)
}

func TestUpdateWritesComposeFile(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "app1")
	os.Mkdir(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "app.yml"), []byte("port: 80"), 0644)
	compose := `services:
  app1:
    image: nginx:1.0-alpine`
	os.WriteFile(filepath.Join(appDir, "docker-compose.yml"), []byte(compose), 0644)

	runner := new(RunnerMock)
	waiter := new(WaiterMock)
	m := AppManager{AppsDir: dir, Runner: runner, Waiter: waiter}

	runner.On("Run", appDir, "docker compose pull").Return(nil)
	runner.On("Run", appDir, mock.Anything).Return(nil)
	waiter.On("WaitPort", "80").Return(nil)
	waiter.On("WaitWeb", "http://localhost:80").Return(nil)

	_, err := m.Update([]string{"app1"})
	assert.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(appDir, "docker-compose.yml"))
	assert.NoError(t, err)
	var obj map[string]any
	err = yaml.Unmarshal(data, &obj)
	assert.NoError(t, err)
	services := obj["services"].(map[string]any)
	svc := services["app1"].(map[string]any)
	assert.Equal(t, "nginx:2.1-alpine", svc["image"])
}
