package main

import (
	"errors"
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

var (
	sampleAppName = "sampleapp"
	sampleAppDir  = appsDir + "/" + sampleAppName
)

var (
	updater              *Updater
	healthCheckerMock    *HealthCheckerMock
	singleAppUpdaterReal *SingleAppUpdaterImpl
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

	fileSystemOperatorMock.EXPECT().GetListOfApps(appsDir).Return([]string{sampleAppName}, nil)
	fileSystemOperatorMock.EXPECT().GetDockerComposeFileContent(sampleAppDir).Return([]byte("sample content"), nil)
	singleAppUpdaterMock.EXPECT().update(sampleAppDir).Return(appUpdate, nil)

	healthCheckerMock.EXPECT().ConductHealthcheckForSingleApp(sampleAppName).Return(getHealthyReport())

	report, err := updater.PerformUpdate()
	assert.Nil(t, err)
	assert.True(t, report.WasSuccessful)
	assert.Equal(t, 1, len(report.AppUpdateReport))
	updateReport := report.AppUpdateReport[0]
	assert.Equal(t, sampleAppName, updateReport.AppName)
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
		AppName:      sampleAppName,
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

	fileSystemOperatorMock.EXPECT().GetListOfApps(appsDir).Return([]string{sampleAppName}, nil)
	fileSystemOperatorMock.EXPECT().GetDockerComposeFileContent(sampleAppDir).Return([]byte("sample content"), nil)
	singleAppUpdaterMock.EXPECT().update(sampleAppDir).Return(appUpdate, nil)

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

	fileSystemOperatorMock.EXPECT().GetListOfApps(appsDir).Return([]string{sampleAppName}, nil)
	fileSystemOperatorMock.EXPECT().GetDockerComposeFileContent(sampleAppDir).Return([]byte("sample content"), nil)
	singleAppUpdaterMock.EXPECT().update(sampleAppDir).Return(nil, errors.New("some error"))
	fileSystemOperatorMock.EXPECT().WriteDockerComposeFileContent(sampleAppDir, []byte("sample content")).Return(nil)

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

	fileSystemOperatorMock.EXPECT().GetListOfApps(appsDir).Return([]string{sampleAppName}, nil)
	fileSystemOperatorMock.EXPECT().GetDockerComposeFileContent(sampleAppDir).Return([]byte("sample content"), nil)
	singleAppUpdaterMock.EXPECT().update(sampleAppDir).Return(nil, errors.New("some error"))
	fileSystemOperatorMock.EXPECT().WriteDockerComposeFileContent(sampleAppDir, []byte("sample content")).Return(errors.New("some other error"))

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

	fileSystemOperatorMock.EXPECT().GetListOfApps(appsDir).Return([]string{sampleAppName}, nil)
	fileSystemOperatorMock.EXPECT().GetDockerComposeFileContent(sampleAppDir).Return(nil, errors.New("some error"))

	updateReport, err := updater.PerformUpdate()
	assert.Nil(t, err)
	assert.False(t, updateReport.WasSuccessful)
	assert.Equal(t, 1, len(updateReport.AppUpdateReport))
	actualReport := updateReport.AppUpdateReport[0]
	expectedReport := getEmptyReport()
	expectedReport.UpdateErrorMessage = "Failed to get docker-compose file content: some error"
	assert.Equal(t, expectedReport, actualReport)
}

func TestUpdater_PerformUpdate_GetListOfAppsFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(appsDir).Return(nil, errors.New("some error"))

	_, err := updater.PerformUpdate()
	assert.Equal(t, "some error", err.Error())
}
