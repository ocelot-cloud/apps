package main

import (
	"errors"
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

var (
	updateApplier UpdateApplier
)

func TestAppUpdater_Success(t *testing.T) {
	setup(t)
	defer assertAppUpdaterMockExpectations(t)

	sampleUpdate := getSampleAppUpdate()
	appUpdateFetcherMock.EXPECT().Fetch(sampleAppDir).Return(sampleUpdate, nil)
	fileSystemOperatorMock.EXPECT().WriteServiceUpdatesIntoComposeFile(sampleAppDir, sampleUpdate.ServiceUpdates).Return(nil)

	actualUpdate, err := updateApplier.ApplyUpdate(sampleAppDir)
	assert.Nil(t, err)
	assert.Equal(t, sampleUpdate, actualUpdate)
}

func TestAppUpdater_FetchFailed(t *testing.T) {
	setup(t)
	defer assertAppUpdaterMockExpectations(t)

	appUpdateFetcherMock.EXPECT().Fetch(sampleAppDir).Return(nil, errors.New("some error"))

	_, err := updateApplier.ApplyUpdate(sampleAppDir)
	assert.Equal(t, "some error", err.Error())
}

func TestAppUpdater_NoUpdateFound(t *testing.T) {
	setup(t)
	defer assertAppUpdaterMockExpectations(t)

	sampleUpdate := getSampleAppUpdate()
	sampleUpdate.WasUpdateFound = false
	appUpdateFetcherMock.EXPECT().Fetch(sampleAppDir).Return(sampleUpdate, nil)

	actualUpdate, err := updateApplier.ApplyUpdate(sampleAppDir)
	assert.Nil(t, err)
	assert.Equal(t, sampleUpdate, actualUpdate)
}

func TestAppUpdater_WriteFails(t *testing.T) {
	setup(t)
	defer assertAppUpdaterMockExpectations(t)

	sampleUpdate := getSampleAppUpdate()
	appUpdateFetcherMock.EXPECT().Fetch(sampleAppDir).Return(sampleUpdate, nil)
	fileSystemOperatorMock.EXPECT().WriteServiceUpdatesIntoComposeFile(sampleAppDir, sampleUpdate.ServiceUpdates).Return(errors.New("some error"))

	_, err := updateApplier.ApplyUpdate(sampleAppDir)
	assert.Equal(t, "some error", err.Error())
}

func setup(t *testing.T) {
	appUpdateFetcherMock = NewAppUpdateFetcherMock(t)
	fileSystemOperatorMock = NewFileSystemOperatorMock(t)

	updateApplier = NewUpdateApplier(
		appUpdateFetcherMock,
		fileSystemOperatorMock,
	)
}

func assertAppUpdaterMockExpectations(t *testing.T) {
	appUpdateFetcherMock.AssertExpectations(t)
	fileSystemOperatorMock.AssertExpectations(t)
}
