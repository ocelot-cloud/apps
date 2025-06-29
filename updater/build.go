//go:build wireinject
// +build wireinject

package main

import "github.com/google/wire"

type Deps struct {
	Updater       *Updater
	HealthChecker HealthChecker
}

func Initialize() (Deps, error) {
	wire.Build(
		NewUpdater,
		NewFileSystemOperator,
		NewHealthChecker,
		NewEndpointChecker,
		NewSingleAppUpdater,
		NewDockerHubClient,
		NewUpdateApplier,
		wire.Struct(new(Deps), "*"),
	)
	return Deps{}, nil
}

func NewUpdater(fs FileSystemOperator, checker HealthChecker, applier UpdateApplier) *Updater {
	return &Updater{
		fileSystemOperator: fs,
		healthChecker:      checker,
		updateApplier:      applier,
	}
}

func NewFileSystemOperator() FileSystemOperator {
	return &FileSystemOperatorImpl{}
}

func NewSingleAppUpdater(fs FileSystemOperator, client DockerHubClient) AppUpdateFetcher {
	return &AppUpdateFetcherImpl{
		fsOperator:      fs,
		dockerHubClient: client,
	}
}

func NewHealthChecker(fs FileSystemOperator, checker EndpointChecker) HealthChecker {
	return &HealthCheckerImpl{
		fileSystemOperator: fs,
		endpointChecker:    checker,
	}
}

func NewDockerHubClient() DockerHubClient {
	return &DockerHubClientImpl{}
}

func NewEndpointChecker() EndpointChecker {
	return &EndpointCheckerImpl{}
}

func NewUpdateApplier(appUpdater AppUpdateFetcher, fileSystemOperator FileSystemOperator) UpdateApplier {
	return &UpdateApplierImpl{
		appUpdater:         appUpdater,
		fileSystemOperator: fileSystemOperator,
	}
}
