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
	mocks2 "updater/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

	// TODO mocks should not be hardcoded, but be generated dynamically using "mockery" CLI command version v3.3.5; maybe we a .mockery.yaml file?
	runner := new(mocks2.RunnerMock)
	waiter := new(mocks2.WaiterMock)
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

// TODO sampleapp should not be in folder "production" but in a folder next to production, like "test" folder or so
func TestUpdate(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "app1")
	os.Mkdir(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "app.yml"), []byte("port: 80"), 0644)
	os.WriteFile(filepath.Join(appDir, "docker-compose.yml"), []byte("services:\n  app1:\n    image: nginx"), 0644)

	runner := new(mocks2.RunnerMock)
	waiter := new(mocks2.WaiterMock)
	m := AppManager{AppsDir: dir, Runner: runner, Waiter: waiter}

	runner.On("Run", appDir, "docker compose pull").Return(nil)
	runner.On("Run", appDir, mock.Anything).Return(nil)
	waiter.On("WaitPort", "80").Return(nil)
	waiter.On("WaitWeb", "http://localhost:80").Return(nil)

	res, err := m.Update([]string{"app1"})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.True(t, res[0].Success)
}

func TestUpdateFailure(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "app1")
	os.Mkdir(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "app.yml"), []byte("port: 80"), 0644)
	os.WriteFile(filepath.Join(appDir, "docker-compose.yml"), []byte("services:\n  app1:\n    image: nginx"), 0644)

	runner := new(mocks2.RunnerMock)
	waiter := new(mocks2.WaiterMock)
	m := AppManager{AppsDir: dir, Runner: runner, Waiter: waiter}

	runner.On("Run", appDir, "docker compose pull").Return(nil)
	runner.On("Run", appDir, mock.Anything).Return(nil)
	waiter.On("WaitPort", "80").Return(assert.AnError)

	res, err := m.Update([]string{"app1"})
	assert.NoError(t, err)
	assert.False(t, res[0].Success)
}

// TODO add for testing "sampleapp2"; when updating two apps and first app update fails, continue the updating process, by now checking the next app. But in the final summary, the failed app should be reported as such. Same goes for healthchecks.

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

	runner := new(mocks2.RunnerMock)
	waiter := new(mocks2.WaiterMock)
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

	runner := new(mocks2.RunnerMock)
	waiter := new(mocks2.WaiterMock)
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

	runner := new(mocks2.RunnerMock)
	waiter := new(mocks2.WaiterMock)
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
}

// TODO not sure if that if given at the moment, but we also need an integration test, that actually update an app, asserts that the current image is newer than the previous one before the update, and passes. I assume in order to achieve that, we must add an extra cobra command in main.go, like "test-update-integration" or so, which really conducts an app update. Also, when that update is successful, write the changed image tag to the docker compose file and assert that. FOr production we want this tool to update the image tags, and when they are healthy, a developer can commit them manually. -> summary: in the end it is quite easy: we execute "go build && ./updater update sampleapp", and afterwards assert that the docker image tag of its docker compose yaml was updated to a newer version than before the update.
