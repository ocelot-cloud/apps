package main

import (
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
	// TODO not sufficient, read the yaml and check the structure explicitly. The "port" keyword and the according value should be present; but nothing else should have been added, assert no other fields or values than the ones we expect.
	assert.Contains(t, string(data), "8080:8080")
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

	healthy, err := m.HealthCheck([]string{"app1"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"app1"}, healthy)
	runner.AssertExpectations(t)
	waiter.AssertExpectations(t)
}

// TODO sampleapp should not be in folder "production" but in a folder next to prodcution, like "test" folder or so
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
	// TODO same as mentioned earlier. Assert the data structure explicitly.
	assert.Contains(t, string(data), "8080:8080")
}

// TODO if a port from app.yml is already present, then some error or so should be thrown as this is not allowed.
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
	tmp, err := injectPort(path, "web", "8080")
	assert.NoError(t, err)
	data, _ := os.ReadFile(tmp)
	// TODO same as mentioned earlier. Assert the data structure explicitly.
	lines := strings.Count(string(data), "8080:8080")
	assert.Equal(t, 1, lines)
}

// TODO I think when doing a test on "all apps" there should be at least two.
func TestHealthCheckAllApps(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "app1")
	os.Mkdir(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "app.yml"), []byte("port: 80"), 0644)
	os.WriteFile(filepath.Join(appDir, "docker-compose.yml"), []byte("services:\n  app1:\n    image: nginx"), 0644)

	runner := new(mocks2.RunnerMock)
	waiter := new(mocks2.WaiterMock)
	m := AppManager{AppsDir: dir, Runner: runner, Waiter: waiter}

	runner.On("Run", appDir, mock.Anything).Return(nil)
	waiter.On("WaitPort", "80").Return(nil)
	waiter.On("WaitWeb", "http://localhost:80").Return(nil)

	healthy, err := m.HealthCheck(nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"app1"}, healthy)
}

// TODO I think when doing a test on "all apps" there should be at least two.
func TestUpdateAllApps(t *testing.T) {
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

	res, err := m.Update(nil)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.True(t, res[0].Success)
}

// TODO not sure if that if given at the moment, but we also need an integration test, that actually update an app, asserts that the current image is newer than the previous one before the update, and passes. I assume in order to achieve that, we must add an extra cobra command in main.go, like "test-update-integration" or so, which really conducts an app update. Also, when that update is successful, write the changed image tag to the docker compose file and assert that. FOr production we want this tool to update the image tags, and when they are healthy, a developer can commit them manually.
