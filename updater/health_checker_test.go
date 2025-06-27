package main

import (
	"errors"
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

var (
	healthChecker          HealthChecker
	fileSystemOperatorMock *FileSystemOperatorMock
	endpointCheckerMock    *EndpointCheckerMock
)

func setupHealthChecker(t *testing.T) {
	fileSystemOperatorMock = NewFileSystemOperatorMock(t)
	endpointCheckerMock = NewEndpointCheckerMock(t)
	healthChecker = &HealthCheckerImpl{
		appsDir:            mockAppsDir,
		fileSystemOperator: fileSystemOperatorMock,
		endpointChecker:    endpointCheckerMock,
	}
}

func assertHealthCheckerMockExpectations(t *testing.T) {
	fileSystemOperatorMock.AssertExpectations(t)
	endpointCheckerMock.AssertExpectations(t)
}

func TestUpdater_PerformHealthCheck(t *testing.T) {
	setupHealthChecker(t)
	defer assertHealthCheckerMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetPortOfApp(appDir).Return("8080", nil)
	fileSystemOperatorMock.EXPECT().InjectPortInDockerCompose(appDir).Return(nil)
	fileSystemOperatorMock.EXPECT().RunInjectedDockerCompose(appDir).Return(nil)
	endpointCheckerMock.EXPECT().TryAccessingIndexPageOnLocalhost("8080").Return(nil)

	report, err := healthChecker.PerformHealthChecks()
	assertHealthyReport(t, err, report)
}

func assertHealthyReport(t *testing.T, err error, report *HealthCheckReport) {
	assert.Nil(t, err)
	assert.True(t, report.AllAppsHealthy)
	assert.Equal(t, 1, len(report.AppHealthReports))
	appReport := report.AppHealthReports[0]
	assert.Equal(t, "sampleapp", appReport.AppName)
	assert.True(t, appReport.Healthy)
}

func TestUpdater_GetAppsFails(t *testing.T) {
	setupHealthChecker(t)
	defer assertHealthCheckerMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return(nil, errors.New("some error"))

	report, err := healthChecker.PerformHealthChecks()
	assert.Nil(t, report)
	assert.NotNil(t, err)
	assert.Equal(t, "some error", err.Error())
}

func TestUpdater_GetPortOfAppFails(t *testing.T) {
	setupHealthChecker(t)
	defer assertHealthCheckerMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetPortOfApp(appDir).Return("", errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, healthChecker, "Failed to get port")
}

func performHealthCheckAndAssertFailedAppReport(t *testing.T, healthChecker HealthChecker, expectedErrorMessage string) {
	report, err := healthChecker.PerformHealthChecks()
	assertErrorInReport(t, err, report, expectedErrorMessage, "some error")
}

func assertErrorInReport(t *testing.T, actualError error, report *HealthCheckReport, expectedHighLevelError, expectedLowLevelError string) {
	assert.Nil(t, actualError)
	assert.False(t, report.AllAppsHealthy)
	assert.Equal(t, 1, len(report.AppHealthReports))
	appReport := report.AppHealthReports[0]
	assert.Equal(t, "sampleapp", appReport.AppName)
	assert.False(t, appReport.Healthy)
	assert.Equal(t, expectedHighLevelError+": "+expectedLowLevelError, appReport.ErrorMessage)
}

func TestUpdater_InjectPortInDockerComposeFails(t *testing.T) {
	setupHealthChecker(t)
	defer assertHealthCheckerMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetPortOfApp(appDir).Return("", nil)
	fileSystemOperatorMock.EXPECT().InjectPortInDockerCompose(appDir).Return(errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, healthChecker, "Failed to inject port in docker-compose")
}

func TestUpdater_RunInjectedDockerComposeFails(t *testing.T) {
	setupHealthChecker(t)
	defer assertHealthCheckerMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetPortOfApp(appDir).Return("8080", nil)
	fileSystemOperatorMock.EXPECT().InjectPortInDockerCompose(appDir).Return(nil)
	fileSystemOperatorMock.EXPECT().RunInjectedDockerCompose(appDir).Return(errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, healthChecker, "Failed to run docker-compose")
}

func TestUpdater_TryAccessingIndexPageOnLocalhostFails(t *testing.T) {
	setupHealthChecker(t)
	defer assertHealthCheckerMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetPortOfApp(appDir).Return("8080", nil)
	fileSystemOperatorMock.EXPECT().InjectPortInDockerCompose(appDir).Return(nil)
	fileSystemOperatorMock.EXPECT().RunInjectedDockerCompose(appDir).Return(nil)
	endpointCheckerMock.EXPECT().TryAccessingIndexPageOnLocalhost("8080").Return(errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, healthChecker, "Failed to access index page")
}
