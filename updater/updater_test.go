package main

import (
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

const (
	mockAppsDir = "/test_apps_dir"
	appDir      = mockAppsDir + "/sampleapp"
)

func TestUpdater_PerformHealthCheck(t *testing.T) {
	fileSystemOperatorMock := NewFileSystemOperatorMock(t)
	endpointCheckerMock := NewEndpointCheckerMock(t)
	updater := Updater{
		appsDir:            mockAppsDir,
		fileSystemOperator: fileSystemOperatorMock,
		endpointChecker:    endpointCheckerMock,
	}

	fileSystemOperatorMock.On("GetListOfApps", mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.On("GetPortOfApp", appDir).Return("8080", nil)
	fileSystemOperatorMock.On("InjectPortInDockerCompose", appDir).Return(nil)
	fileSystemOperatorMock.On("RunInjectedDockerCompose", appDir).Return(nil)
	endpointCheckerMock.On("TryAccessingIndexPageOnLocalhost", "8080").Return(nil)

	report, err := updater.PerformHealthCheck()
	assert.Nil(t, err)
	assert.True(t, report.AllAppsHealthy)
	assert.Equal(t, 1, len(report.AppReports))
	appReport := report.AppReports[0]
	assert.Equal(t, "sampleapp", appReport.AppName)
	assert.True(t, appReport.Healthy)

	fileSystemOperatorMock.AssertExpectations(t)
	endpointCheckerMock.AssertExpectations(t)
}
