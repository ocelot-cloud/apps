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
	AppName            string
	WasSuccessful      bool
	WasUpdateAvailable bool
	AppHealthReport    *AppHealthReport
	AppUpdates         *AppUpdate
	UpdateErrorMessage string
}

type AppUpdate struct {
	WasUpdateFound bool
	ServiceUpdates []ServiceUpdate
	ErrorMessage   string
}

type ServiceUpdate struct {
	ServiceName string
	OldTag      string
	NewTag      string
}

type Service struct {
	Name  string
	Image string
	Tag   string
}

type SingleAppUpdaterImpl struct {
	fsOperator      SingleAppUpdateFileSystemOperator
	dockerHubClient DockerHubClient
}
