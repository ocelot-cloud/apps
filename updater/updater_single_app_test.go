package main

import (
	"errors"
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

func assertSingleAppUpdaterMockExpectations(t *testing.T) {
	fileSystemOperatorMock.AssertExpectations(t)
	dockerHubClientMock.AssertExpectations(t)
}

func setupSingleAppUpdater(t *testing.T) {
	fileSystemOperatorMock = NewFileSystemOperatorMock(t)
	dockerHubClientMock = NewDockerHubClientMock(t)
	singleAppUpdaterReal = &AppUpdateFetcherImpl{
		fsOperator:      fileSystemOperatorMock,
		dockerHubClient: dockerHubClientMock,
	}
}

func TestAppUpdaterSuccess(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetAppServices(sampleAppDir).Return([]Service{
		{Name: "sampleapp", Image: "ocelot/sampleapp", Tag: "1.0.0"},
	}, nil)
	dockerHubClientMock.EXPECT().listImageTags("ocelot/sampleapp").Return([]string{"1.0.0", "1.0.1"}, nil)

	appUpdate, err := singleAppUpdaterReal.Fetch(sampleAppDir)
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

	fileSystemOperatorMock.EXPECT().GetAppServices(sampleAppDir).Return(nil, errors.New("some error"))

	_, err := singleAppUpdaterReal.Fetch(sampleAppDir)
	assert.Equal(t, "some error", err.Error())
}

func TestAppUpdater_ListImageTagsFails(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetAppServices(sampleAppDir).Return([]Service{
		{Name: "sampleapp", Image: "ocelot/sampleapp", Tag: "1.0.0"},
	}, nil)
	dockerHubClientMock.EXPECT().listImageTags("ocelot/sampleapp").Return(nil, errors.New("some error"))

	_, err := singleAppUpdaterReal.Fetch(sampleAppDir)
	assert.Equal(t, "some error", err.Error())
}

func TestAppUpdater_SuccessButNoNewUpdateFound(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetAppServices(sampleAppDir).Return([]Service{
		{Name: "sampleapp", Image: "ocelot/sampleapp", Tag: "1.0.0"},
	}, nil)
	dockerHubClientMock.EXPECT().listImageTags("ocelot/sampleapp").Return([]string{"1.0.0"}, nil)

	appUpdate, err := singleAppUpdaterReal.Fetch(sampleAppDir)
	assert.Nil(t, err)
	assert.False(t, appUpdate.WasUpdateFound)
	assert.Equal(t, 0, len(appUpdate.ServiceUpdates))
}

func TestAppUpdater_FilterLatestImageTagFails(t *testing.T) {
	setupSingleAppUpdater(t)
	defer assertSingleAppUpdaterMockExpectations(t)

	fileSystemOperatorMock.EXPECT().GetAppServices(sampleAppDir).Return([]Service{
		{Name: "sampleapp", Image: "ocelot/sampleapp", Tag: "invalid-tag"},
	}, nil)
	dockerHubClientMock.EXPECT().listImageTags("ocelot/sampleapp").Return([]string{"1.0.0", "1.0.1"}, nil)

	_, err := singleAppUpdaterReal.Fetch(sampleAppDir)
	assert.Equal(t, "integer conversion failed", err.Error())
}
