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
	healthCheckerMock    *HealthCheckerMock
	singleAppUpdaterReal *SingleAppUpdaterReal
	dockerHubClientMock  *DockerHubClientMock
	singleAppUpdaterMock *SingleAppUpdaterMock

	singleAppUpdateFileSystemOperatorMock *SingleAppUpdateFileSystemOperatorMock
)

func setupUpdater(t *testing.T) {
	fileSystemOperatorMock = NewFileSystemOperatorMock(t)
	singleAppUpdaterMock = NewSingleAppUpdaterMock(t)
	dockerHubClientMock = NewDockerHubClientMock(t)
	healthCheckerMock = NewHealthCheckerMock(t)

	updater = &Updater{
		appsDir:            mockAppsDir,
		fileSystemOperator: fileSystemOperatorMock,
		appUpdater:         singleAppUpdaterMock,
		dockerHubClient:    dockerHubClientMock,
		healthChecker:      healthCheckerMock,
	}
}

func assertUpdaterMockExpectations(t *testing.T) {
	fileSystemOperatorMock.AssertExpectations(t)
	dockerHubClientMock.AssertExpectations(t)
	healthCheckerMock.AssertExpectations(t)
}

func TestUpdater_PerformUpdateSuccessfully(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)
	appUpdate := &AppUpdate{
		WasUpdateFound: true,
		ServiceUpdates: []ServiceUpdate{
			{
				ServiceName: "service1",
				OldTag:      "1.0.0",
				NewTag:      "1.0.1",
			},
		},
	}

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetDockerComposeFileContent(appDir).Return([]byte("sample content"), nil)
	singleAppUpdaterMock.EXPECT().update(appDir).Return(appUpdate, nil)

	healthCheckerMock.EXPECT().ConductHealthcheckForSingleApp("sampleapp").Return(getHealthyReport())

	report, err := updater.PerformUpdate()
	assert.Nil(t, err)
	assert.True(t, report.WasSuccessful)
	assert.Equal(t, 1, len(report.AppUpdateReport))
	updateReport := report.AppUpdateReport[0]
	assert.True(t, updateReport.WasSuccessful)
	assert.True(t, updateReport.WasUpdateAvailable)
	assert.Equal(t, "", updateReport.UpdateErrorMessage)
	appUpdates := updateReport.AppUpdates
	assert.True(t, appUpdates.WasUpdateFound)
	assert.Equal(t, "", appUpdates.ErrorMessage)
	assert.Equal(t, 1, len(appUpdates.ServiceUpdates))
	serviceUpdate := appUpdates.ServiceUpdates[0]
	assert.Equal(t, "service1", serviceUpdate.ServiceName)
	assert.Equal(t, "1.0.0", serviceUpdate.OldTag)
	assert.Equal(t, "1.0.1", serviceUpdate.NewTag)
}

func getHealthyReport() AppHealthReport {
	return AppHealthReport{
		AppName:      "sampleapp",
		Healthy:      true,
		ErrorMessage: "",
	}
}

func TestUpdater_PerformUpdateSuccessfullyWithoutNewTag(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)
	appUpdate := &AppUpdate{
		WasUpdateFound: false,
		ServiceUpdates: []ServiceUpdate{},
	}

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetDockerComposeFileContent(appDir).Return([]byte("sample content"), nil)
	singleAppUpdaterMock.EXPECT().update(appDir).Return(appUpdate, nil)

	report, err := updater.PerformUpdate()
	assert.Nil(t, err)
	assert.True(t, report.WasSuccessful)
	actualReport := report.AppUpdateReport[0]
	expectedReport := getEmptyReport()
	expectedReport.WasSuccessful = true
	assert.Equal(t, expectedReport, actualReport)
}

func TestUpdater_PerformUpdate_SingleAppUpdateFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetDockerComposeFileContent(appDir).Return([]byte("sample content"), nil)
	singleAppUpdaterMock.EXPECT().update(appDir).Return(nil, errors.New("some error"))
	fileSystemOperatorMock.EXPECT().WriteDockerComposeFileContent(appDir, []byte("sample content")).Return(nil)

	updateReport, err := updater.PerformUpdate()
	assert.Nil(t, err)
	assert.False(t, updateReport.WasSuccessful)
	assert.Equal(t, 1, len(updateReport.AppUpdateReport))
	actualReport := updateReport.AppUpdateReport[0]
	expectedReport := getEmptyReport()
	expectedReport.UpdateErrorMessage = "Failed to update app: some error"
	assert.Equal(t, expectedReport, actualReport)
}

func TestUpdater_PerformUpdate_WriteDockerComposeFileContentFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetDockerComposeFileContent(appDir).Return([]byte("sample content"), nil)
	singleAppUpdaterMock.EXPECT().update(appDir).Return(nil, errors.New("some error"))
	fileSystemOperatorMock.EXPECT().WriteDockerComposeFileContent(appDir, []byte("sample content")).Return(errors.New("some other error"))

	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, "Failed to write docker-compose file back to original state: some other error", r)
		} else {
			t.Fatal("expected panic but none occurred")
		}
	}()
	updater.PerformUpdate()
}

func TestUpdater_PerformUpdate_GetDockerComposeFileContentFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetDockerComposeFileContent(appDir).Return(nil, errors.New("some error"))

	updateReport, err := updater.PerformUpdate()
	assert.Nil(t, err)
	assert.False(t, updateReport.WasSuccessful)
	assert.Equal(t, 1, len(updateReport.AppUpdateReport))
	actualReport := updateReport.AppUpdateReport[0]
	expectedReport := getEmptyReport()
	expectedReport.UpdateErrorMessage = "Failed to get docker-compose file content: some error"
	assert.Equal(t, expectedReport, actualReport)
}

// TODO main: if not all apps are healthy in report, exit with code 1
