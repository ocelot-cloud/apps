package main

import "testing"

var (
	updateApplier UpdateApplier
)

func TestAppUpdater_Success(t *testing.T) {
	setup(t)
	defer assertAppUpdaterMockExpectations(t)

	//updateApplier.ApplyUpdate(sampleAppDir)
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
