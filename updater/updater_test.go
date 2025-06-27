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

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetPortOfApp(appDir).Return("8080", nil)
	fileSystemOperatorMock.EXPECT().InjectPortInDockerCompose(appDir).Return(nil)
	fileSystemOperatorMock.EXPECT().RunInjectedDockerCompose(appDir).Return(nil)
	endpointCheckerMock.EXPECT().TryAccessingIndexPageOnLocalhost("8080").Return(nil)

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
	singleAppUpdateFileSystemOperatorMock.AssertExpectations(t)
	dockerHubClientMock.AssertExpectations(t)
}

func TestUpdater_GetAppsFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, errors.New("some error"))

	report, err := updater.PerformHealthCheck()
	assert.Nil(t, report)
	assert.NotNil(t, err)
	assert.Equal(t, "some error", err.Error())
}

func TestUpdater_GetPortOfAppFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetPortOfApp(appDir).Return("", errors.New("some error"))

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

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetPortOfApp(appDir).Return("", nil)
	fileSystemOperatorMock.EXPECT().InjectPortInDockerCompose(appDir).Return(errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, updater, "Failed to inject port in docker-compose")
}

func TestUpdater_RunInjectedDockerComposeFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetPortOfApp(appDir).Return("8080", nil)
	fileSystemOperatorMock.EXPECT().InjectPortInDockerCompose(appDir).Return(nil)
	fileSystemOperatorMock.EXPECT().RunInjectedDockerCompose(appDir).Return(errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, updater, "Failed to run docker-compose")
}

func TestUpdater_TryAccessingIndexPageOnLocalhostFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetPortOfApp(appDir).Return("8080", nil)
	fileSystemOperatorMock.EXPECT().InjectPortInDockerCompose(appDir).Return(nil)
	fileSystemOperatorMock.EXPECT().RunInjectedDockerCompose(appDir).Return(nil)
	endpointCheckerMock.EXPECT().TryAccessingIndexPageOnLocalhost("8080").Return(errors.New("some error"))

	performHealthCheckAndAssertFailedAppReport(t, updater, "Failed to access index page")
}

// TODO assert update was conducted
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
	singleAppUpdaterMock.EXPECT().update(appDir).Return(appUpdate, nil) // TODO first nil must be appUpdate
	fileSystemOperatorMock.EXPECT().GetPortOfApp(appDir).Return("8080", nil)
	fileSystemOperatorMock.EXPECT().InjectPortInDockerCompose(appDir).Return(nil)
	fileSystemOperatorMock.EXPECT().RunInjectedDockerCompose(appDir).Return(nil)
	endpointCheckerMock.EXPECT().TryAccessingIndexPageOnLocalhost("8080").Return(nil)

	report, err := updater.PerformUpdate()
	assertHealthyReport(t, err, report)
	/* TODO
	singleAppReport := report.AppReports[0]
	assert.Equal(t, *appUpdate, *singleAppReport.AppUpdate)
	*/
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
	assertHealthyReport(t, err, report)
}

func TestUpdater_PerformUpdate_GetImagesFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetDockerComposeFileContent(appDir).Return([]byte("sample content"), nil)
	singleAppUpdaterMock.EXPECT().update(appDir).Return(nil, errors.New("some error"))
	fileSystemOperatorMock.EXPECT().WriteDockerComposeFileContent(appDir, []byte("sample content")).Return(nil)

	performUpdateAndAssertFailedAppReport(t, updater, "Failed to update app", "some error")
}

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

func TestUpdater_PerformUpdate_GetDockerComposeFileContentFails(t *testing.T) {
	setupUpdater(t)
	defer assertUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetListOfApps(mockAppsDir).Return([]string{"sampleapp"}, nil)
	fileSystemOperatorMock.EXPECT().GetDockerComposeFileContent(appDir).Return(nil, errors.New("some error"))

	performUpdateAndAssertFailedAppReport(t, updater, "Failed to get docker-compose file content", "some error")
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

// TODO main: if not all apps are healthy in report, exit with code 1
// TODO introduce mutation testing in every component
