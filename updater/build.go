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
		NewSingleAppUpdater,
		NewHealthChecker,
		NewDockerHubClient,
		NewEndpointChecker,
		NewFileSystemUpdateOperator,
		wire.Struct(new(Deps), "*"),
	)
	return Deps{}, nil
}

func NewUpdater(fs FileSystemOperator, appUpdater SingleAppUpdater, checker HealthChecker, client DockerHubClient) *Updater {
	return &Updater{
		fileSystemOperator: fs,
		appUpdater:         appUpdater,
		healthChecker:      checker,
		dockerHubClient:    client,
	}
}

func NewFileSystemOperator() FileSystemOperator {
	return &FileSystemOperatorImpl{}
}

func NewSingleAppUpdater(fs SingleAppUpdateFileSystemOperator, client DockerHubClient) SingleAppUpdater {
	return &SingleAppUpdaterImpl{
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

func NewFileSystemUpdateOperator() SingleAppUpdateFileSystemOperator {
	return &SingleAppUpdateFileSystemOperatorImpl{}
}
