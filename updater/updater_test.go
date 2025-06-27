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

	// TODO !! what is that? do I need it?
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
	dockerHubClientMock.AssertExpectations(t)
	healthCheckerMock.AssertExpectations(t)
}

func assertSingleAppUpdaterMockExpectations(t *testing.T) {
	singleAppUpdateFileSystemOperatorMock.AssertExpectations(t)
	dockerHubClientMock.AssertExpectations(t)
}

/*
	func performUpdateAndAssertFailedAppReport(t *testing.T, updater *Updater, expectedHighLevelErrorMessage, expectedLowLevelErrorMessage string) {
		report, err := updater.PerformUpdate()
		assertErrorInReport(t, err, report, expectedHighLevelErrorMessage, expectedLowLevelErrorMessage)

}
TODO add this above?

	assert.Equal(t, 1, len(report.AppReports))
	appUpdate := report.AppReports
	appReport := appUpdate[0]
	assert.Equal(t, appReport, AppHealthReport{})
*/

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
	appHealthReport := AppHealthReport{
		AppName:      "sampleapp",
		Healthy:      true,
		ErrorMessage: "",
	}
	healthCheckerMock.EXPECT().ConductHealthcheckForSingleApp("sampleapp").Return(appHealthReport)

	report, err := updater.PerformUpdate()
	assert.Nil(t, err)
	assert.True(t, report.WasSuccessful)
	assert.Equal(t, 1, len(report.AppUpdateReport))
	updateReport := report.AppUpdateReport[0]
	assert.True(t, updateReport.WasSuccessful)
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

/* TODO !!
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
	assertHealthyReport(t, err, report)
	singleAppReport := report.AppHealthReports[0]
	assert.Equal(t, *appUpdate, *singleAppReport.AppUpdate)
}
*/

/* TODO !!
func TestUpdater_PerformUpdate_GetImagesFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetDockerComposeFileContent(appDir).Return([]byte("sample content"), nil)
	singleAppUpdaterMock.EXPECT().update(appDir).Return(nil, errors.New("some error"))
	fileSystemOperatorMock.EXPECT().WriteDockerComposeFileContent(appDir, []byte("sample content")).Return(nil)

	performUpdateAndAssertFailedAppReport(t, updater, "Failed to update app", "some error")
}
*/

func TestAppUpdaterSuccess(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	singleAppUpdateFileSystemOperatorMock.EXPECT().GetImagesOfApp(appDir).Return([]Service{
		{Name: "sampleapp", Image: "ocelot/sampleapp", Tag: "1.0.0"},
	}, nil)
	dockerHubClientMock.EXPECT().listImageTags("ocelot/sampleapp").Return([]string{"1.0.0", "1.0.1"}, nil)

	appUpdate, err := singleAppUpdaterReal.update(appDir)
	assert.Nil(t, err)
	assert.True(t, appUpdate.WasUpdateFound)
	assert.Equal(t, 1, len(appUpdate.ServiceUpdates))
	update := appUpdate.ServiceUpdates[0]
	assert.Equal(t, "sampleapp", update.ServiceName)
	assert.Equal(t, "1.0.0", update.OldTag)
	assert.Equal(t, "1.0.1", update.NewTag)
}

func TestAppUpdater_GetImagesOfAppFails(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	singleAppUpdateFileSystemOperatorMock.EXPECT().GetImagesOfApp(appDir).Return(nil, errors.New("some error"))

	_, err := singleAppUpdaterReal.update(appDir)
	assert.Equal(t, "some error", err.Error())
}

func TestAppUpdater_ListImageTagsFails(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	singleAppUpdateFileSystemOperatorMock.EXPECT().GetImagesOfApp(appDir).Return([]Service{
		{Name: "sampleapp", Image: "ocelot/sampleapp", Tag: "1.0.0"},
	}, nil)
	dockerHubClientMock.EXPECT().listImageTags("ocelot/sampleapp").Return(nil, errors.New("some error"))

	_, err := singleAppUpdaterReal.update(appDir)
	assert.Equal(t, "some error", err.Error())
}

func TestAppUpdater_SuccessButNoNewUpdateFound(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	singleAppUpdateFileSystemOperatorMock.EXPECT().GetImagesOfApp(appDir).Return([]Service{
		{Name: "sampleapp", Image: "ocelot/sampleapp", Tag: "1.0.0"},
	}, nil)
	dockerHubClientMock.EXPECT().listImageTags("ocelot/sampleapp").Return([]string{"1.0.0"}, nil)

	appUpdate, err := singleAppUpdaterReal.update(appDir)
	assert.Nil(t, err)
	assert.False(t, appUpdate.WasUpdateFound)
	assert.Equal(t, 0, len(appUpdate.ServiceUpdates))
}

func TestAppUpdater_FilterLatestImageTagFails(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	singleAppUpdateFileSystemOperatorMock.EXPECT().GetImagesOfApp(appDir).Return([]Service{
		{Name: "sampleapp", Image: "ocelot/sampleapp", Tag: "invalid-tag"},
	}, nil)
	dockerHubClientMock.EXPECT().listImageTags("ocelot/sampleapp").Return([]string{"1.0.0", "1.0.1"}, nil)

	_, err := singleAppUpdaterReal.update(appDir)
	assert.Equal(t, "integer conversion failed", err.Error())
}

/* TODO !!
func TestUpdater_PerformUpdate_GetDockerComposeFileContentFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetDockerComposeFileContent(appDir).Return(nil, errors.New("some error"))

	performUpdateAndAssertFailedAppReport(t, updater, "Failed to get docker-compose file content", "some error")
}
*/

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

// TODO main: if not all apps are healthy in report, exit with code 1
// TODO introduce mutation testing in every component
