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
	updater              *Updater
	singleAppUpdaterReal *SingleAppUpdaterReal

	fileSystemOperatorMock *FileSystemOperatorMock
	endpointCheckerMock    *EndpointCheckerMock
	dockerHubClientMock    *DockerHubClientMock
	singleAppUpdaterMock   *SingleAppUpdaterMock

	singleAppUpdateFileSystemOperatorMock *SingleAppUpdateFileSystemOperatorMock
)

func TestUpdater_PerformHealthCheck(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

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

func setupUpdater(t *testing.T) {
	fileSystemOperatorMock = NewFileSystemOperatorMock(t)
	endpointCheckerMock = NewEndpointCheckerMock(t)
	dockerHubClientMock = NewDockerHubClientMock(t)
	singleAppUpdaterMock = NewSingleAppUpdaterMock(t)
	updater = &Updater{
		appsDir:            mockAppsDir,
		fileSystemOperator: fileSystemOperatorMock,
		endpointChecker:    endpointCheckerMock,
		appUpdater:         singleAppUpdaterMock,
	}
}

func setupSingleAppUpdater(t *testing.T) {
	singleAppUpdateFileSystemOperatorMock = NewSingleAppUpdateFileSystemOperatorMock(t)
	dockerHubClientMock = NewDockerHubClientMock(t)
	singleAppUpdaterReal = &SingleAppUpdaterReal{
		fsOperator:      singleAppUpdateFileSystemOperatorMock,
		dockerHubClient: dockerHubClientMock,
	}
}

func assertUpdaterMockExpectations(t *testing.T) {
	fileSystemOperatorMock.AssertExpectations(t)
	endpointCheckerMock.AssertExpectations(t)
}

func assertSingleAppUpdaterMockExpectations(t *testing.T) {
	fileSystemOperatorMock.AssertExpectations(t)
	dockerHubClientMock.AssertExpectations(t)
}

func TestUpdater_GetAppsFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, errors.New("some error"))

	report, err := updater.PerformHealthCheck()
	assert.Nil(t, report)
	assert.NotNil(t, err)
	assert.Equal(t, "some error", err.Error())
}

func TestUpdater_GetPortOfAppFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.On("GetPortOfApp", appDir).Return("", errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, updater, "Failed to get port")
}

func performHealthCheckAndAssertFailedAppReport(t *testing.T, updater *Updater, expectedErrorMessage string) {
	report, err := updater.PerformHealthCheck()
	assertErrorInReport(t, err, report, expectedErrorMessage, "some error")
}

func assertErrorInReport(t *testing.T, actualError error, report *HealthCheckReport, expectedHighLevelError, expectedLowLevelError string) {
	assert.Nil(t, actualError)
	assert.False(t, report.AllAppsHealthy)
	assert.Equal(t, 1, len(report.AppReports))
	appReport := report.AppReports[0]
	assert.Equal(t, "sampleapp", appReport.AppName)
	assert.False(t, appReport.Healthy)
	assert.Equal(t, expectedHighLevelError+": "+expectedLowLevelError, appReport.ErrorMessage)
}

func performUpdateAndAssertFailedAppReport(t *testing.T, updater *Updater, expectedHighLevelErrorMessage, expectedLowLevelErrorMessage string) {
	report, err := updater.PerformUpdate()
	assertErrorInReport(t, err, report, expectedHighLevelErrorMessage, expectedLowLevelErrorMessage)
}

func TestUpdater_InjectPortInDockerComposeFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.On("GetPortOfApp", appDir).Return("", nil)
	fileSystemOperatorMock.On("InjectPortInDockerCompose", appDir).Return(errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, updater, "Failed to inject port in docker-compose")
}

func TestUpdater_RunInjectedDockerComposeFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.On("GetPortOfApp", appDir).Return("8080", nil)
	fileSystemOperatorMock.On("InjectPortInDockerCompose", appDir).Return(nil)
	fileSystemOperatorMock.On("RunInjectedDockerCompose", appDir).Return(errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, updater, "Failed to run docker-compose")
}

func TestUpdater_TryAccessingIndexPageOnLocalhostFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.On("GetPortOfApp", appDir).Return("8080", nil)
	fileSystemOperatorMock.On("InjectPortInDockerCompose", appDir).Return(nil)
	fileSystemOperatorMock.On("RunInjectedDockerCompose", appDir).Return(nil)
	endpointCheckerMock.On("TryAccessingIndexPageOnLocalhost", "8080").Return(errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, updater, "Failed to access index page")
}

func TestUpdater_PerformUpdateSuccessfully(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, nil)
	singleAppUpdaterMock.On("update", appDir).Return(nil)
	fileSystemOperatorMock.On("GetPortOfApp", appDir).Return("8080", nil)
	fileSystemOperatorMock.On("InjectPortInDockerCompose", appDir).Return(nil)
	fileSystemOperatorMock.On("RunInjectedDockerCompose", appDir).Return(nil)
	endpointCheckerMock.On("TryAccessingIndexPageOnLocalhost", "8080").Return(nil)

	report, err := updater.PerformUpdate()
	assertHealthyReport(t, err, report)
}

func TestUpdater_PerformUpdateSuccessfullyWithoutNewTag(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, nil)
	singleAppUpdaterMock.On("update", appDir).Return(nil)
	fileSystemOperatorMock.On("GetPortOfApp", appDir).Return("8080", nil)
	fileSystemOperatorMock.On("InjectPortInDockerCompose", appDir).Return(nil)
	fileSystemOperatorMock.On("RunInjectedDockerCompose", appDir).Return(nil)
	endpointCheckerMock.On("TryAccessingIndexPageOnLocalhost", "8080").Return(nil)

	report, err := updater.PerformUpdate()
	assertHealthyReport(t, err, report)
}

func TestUpdater_PerformUpdate_GetImagesFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, nil)
	singleAppUpdaterMock.On("update", appDir).Return(errors.New("some error"))

	performUpdateAndAssertFailedAppReport(t, updater, "Failed to update app", "some error")
}

func TestAppUpdaterSuccess(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	singleAppUpdateFileSystemOperatorMock.On("GetImagesOfApp", appDir).Return([]Service{
		{Name: "sampleapp", Image: "ocelot/sampleapp", Tag: "1.0.0"},
	}, nil)
	dockerHubClientMock.On("listImageTags", "ocelot/sampleapp").Return([]string{"1.0.0", "1.0.1"}, nil)
	singleAppUpdateFileSystemOperatorMock.On("WriteNewTagToDockerCompose", appDir, "sampleapp", "1.0.1").Return(nil)

	wasAnyServiceUpdated, err := singleAppUpdaterReal.update(appDir)
	assert.Nil(t, err)
	assert.True(t, wasAnyServiceUpdated)
}

func TestAppUpdater_GetImagesOfAppFails(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	singleAppUpdateFileSystemOperatorMock.On("GetImagesOfApp", appDir).Return(nil, errors.New("some error"))

	_, err := singleAppUpdaterReal.update(appDir)
	assert.Equal(t, "some error", err.Error())
}

func TestAppUpdater_ListImageTagsFails(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	singleAppUpdateFileSystemOperatorMock.On("GetImagesOfApp", appDir).Return([]Service{
		{Name: "sampleapp", Image: "ocelot/sampleapp", Tag: "1.0.0"},
	}, nil)
	dockerHubClientMock.On("listImageTags", "ocelot/sampleapp").Return(nil, errors.New("some error"))

	_, err := singleAppUpdaterReal.update(appDir)
	assert.Equal(t, "some error", err.Error())
}

func TestAppUpdater_WriteNewTagToDockerComposeFails(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	singleAppUpdateFileSystemOperatorMock.On("GetImagesOfApp", appDir).Return([]Service{
		{Name: "sampleapp", Image: "ocelot/sampleapp", Tag: "1.0.0"},
	}, nil)
	dockerHubClientMock.On("listImageTags", "ocelot/sampleapp").Return([]string{"1.0.0", "1.0.1"}, nil)
	singleAppUpdateFileSystemOperatorMock.On("WriteNewTagToDockerCompose", appDir, "sampleapp", "1.0.1").Return(errors.New("some error"))

	_, err := singleAppUpdaterReal.update(appDir)
	assert.Equal(t, "some error", err.Error())
}

func TestAppUpdater_SuccessButNoNewUpdateFound(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	singleAppUpdateFileSystemOperatorMock.On("GetImagesOfApp", appDir).Return([]Service{
		{Name: "sampleapp", Image: "ocelot/sampleapp", Tag: "1.0.0"},
	}, nil)
	dockerHubClientMock.On("listImageTags", "ocelot/sampleapp").Return([]string{"1.0.0"}, nil)

	wasAnyServiceUpdated, err := singleAppUpdaterReal.update(appDir)
	assert.Nil(t, err)
	assert.False(t, wasAnyServiceUpdated)
}

func TestAppUpdater_FilterLatestImageTagFails(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	singleAppUpdateFileSystemOperatorMock.On("GetImagesOfApp", appDir).Return([]Service{
		{Name: "sampleapp", Image: "ocelot/sampleapp", Tag: "invalid-tag"},
	}, nil)
	dockerHubClientMock.On("listImageTags", "ocelot/sampleapp").Return([]string{"1.0.0", "1.0.1"}, nil)

	_, err := singleAppUpdaterReal.update(appDir)
	assert.Equal(t, "integer conversion failed", err.Error())
}

// TODO main: if not all apps are healthy in report, exit with code 1
