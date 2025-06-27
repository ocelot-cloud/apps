package main

type Updater struct {
	appsDir            string
	fileSystemOperator FileSystemOperator
	appUpdater         SingleAppUpdater
	healthChecker      HealthChecker
	dockerHubClient    DockerHubClient
}
type UpdateReport struct {
	WasSuccessful   bool
	AppUpdateReport []AppUpdateReport
}

type AppUpdateReport struct {
	WasSuccessful      bool
	AppHealthReport    *AppHealthReport
	AppUpdates         *AppUpdate
	UpdateErrorMessage string
}
