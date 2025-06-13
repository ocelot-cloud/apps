package main

import (
	"errors"
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

const (
	mockAppsDir = "/test_apps_dir"
	appDir      = mockAppsDir + "/sampleapp"
)

var (
	fileSystemOperatorMock *FileSystemOperatorMock
	endpointCheckerMock    *EndpointCheckerMock
	dockerHubClientMock    *DockerHubClientMock
	updater                *Updater
)

func TestUpdater_PerformHealthCheck(t *testing.T) {
	setup(t)
	defer assertMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.On("GetPortOfApp", appDir).Return("8080", nil)
	fileSystemOperatorMock.On("InjectPortInDockerCompose", appDir).Return(nil)
	fileSystemOperatorMock.On("RunInjectedDockerCompose", appDir).Return(nil)
	endpointCheckerMock.On("TryAccessingIndexPageOnLocalhost", "8080").Return(nil)

	report, err := updater.PerformHealthCheck()
	assertHealthyReport(t, err, report)
}

func assertHealthyReport(t *testing.T, err error, report *HealthCheckReport) {
	assert.Nil(t, err)
	assert.True(t, report.AllAppsHealthy)
	assert.Equal(t, 1, len(report.AppReports))
	appReport := report.AppReports[0]
	assert.Equal(t, "sampleapp", appReport.AppName)
	assert.True(t, appReport.Healthy)
}

func setup(t *testing.T) {
	fileSystemOperatorMock = NewFileSystemOperatorMock(t)
	endpointCheckerMock = NewEndpointCheckerMock(t)
	dockerHubClientMock = NewDockerHubClientMock(t)
	updater = &Updater{
		appsDir:            mockAppsDir,
		fileSystemOperator: fileSystemOperatorMock,
		endpointChecker:    endpointCheckerMock,
		dockerHubClient:    dockerHubClientMock,
	}
}

func assertMockExpectations(t *testing.T) {
	fileSystemOperatorMock.AssertExpectations(t)
	endpointCheckerMock.AssertExpectations(t)
	dockerHubClientMock.AssertExpectations(t)
}

func TestUpdater_GetAppsFails(t *testing.T) {
	setup(t)
	defer assertMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, errors.New("some error"))

	report, err := updater.PerformHealthCheck()
	assert.Nil(t, report)
	assert.NotNil(t, err)
	assert.Equal(t, "some error", err.Error())
}

func TestUpdater_GetPortOfAppFails(t *testing.T) {
	setup(t)
	defer assertMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.On("GetPortOfApp", appDir).Return("", errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, updater, "Failed to get port")
}

func performHealthCheckAndAssertFailedAppReport(t *testing.T, updater *Updater, expectedErrorMessage string) {
	report, err := updater.PerformHealthCheck()
	assertErrorInReport(t, err, report, expectedErrorMessage)
}

func assertErrorInReport(t *testing.T, err error, report *HealthCheckReport, expectedErrorMessage string) {
	assert.Nil(t, err)
	assert.False(t, report.AllAppsHealthy)
	assert.Equal(t, 1, len(report.AppReports))
	appReport := report.AppReports[0]
	assert.Equal(t, "sampleapp", appReport.AppName)
	assert.False(t, appReport.Healthy)
	assert.Equal(t, expectedErrorMessage+": some error", appReport.ErrorMessage)
}

func performUpdateAndAssertFailedAppReport(t *testing.T, updater *Updater, expectedErrorMessage string) {
	report, err := updater.PerformUpdate()
	assertErrorInReport(t, err, report, expectedErrorMessage)
}

func TestUpdater_InjectPortInDockerComposeFails(t *testing.T) {
	setup(t)
	defer assertMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.On("GetPortOfApp", appDir).Return("", nil)
	fileSystemOperatorMock.On("InjectPortInDockerCompose", appDir).Return(errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, updater, "Failed to inject port in docker-compose")
}

func TestUpdater_RunInjectedDockerComposeFails(t *testing.T) {
	setup(t)
	defer assertMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.On("GetPortOfApp", appDir).Return("8080", nil)
	fileSystemOperatorMock.On("InjectPortInDockerCompose", appDir).Return(nil)
	fileSystemOperatorMock.On("RunInjectedDockerCompose", appDir).Return(errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, updater, "Failed to run docker-compose")
}

func TestUpdater_TryAccessingIndexPageOnLocalhostFails(t *testing.T) {
	setup(t)
	defer assertMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.On("GetPortOfApp", appDir).Return("8080", nil)
	fileSystemOperatorMock.On("InjectPortInDockerCompose", appDir).Return(nil)
	fileSystemOperatorMock.On("RunInjectedDockerCompose", appDir).Return(nil)
	endpointCheckerMock.On("TryAccessingIndexPageOnLocalhost", "8080").Return(errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, updater, "Failed to access index page")
}

func TestUpdater_PerformUpdate(t *testing.T) {
	setup(t)
	defer assertMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.On("GetImagesOfApp", appDir).Return([]Service{
		{Name: "sampleapp", Image: "ocelot/sampleapp", Tag: "1.0.0"},
	}, nil)
	dockerHubClientMock.On("listImageTags", "ocelot/sampleapp").Return([]string{"1.0.0", "1.0.1"}, nil)
	fileSystemOperatorMock.On("WriteNewTagToDockerCompose", appDir, "sampleapp", "1.0.1").Return(nil)
	fileSystemOperatorMock.On("GetPortOfApp", appDir).Return("8080", nil)
	fileSystemOperatorMock.On("InjectPortInDockerCompose", appDir).Return(nil)
	fileSystemOperatorMock.On("RunInjectedDockerCompose", appDir).Return(nil)
	endpointCheckerMock.On("TryAccessingIndexPageOnLocalhost", "8080").Return(nil)

	report, err := updater.PerformUpdate()
	assertHealthyReport(t, err, report)
}

// TODO main: if not all apps are healthy in report, exit with code 1
