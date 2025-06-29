package main

//go:generate mockery
type FileSystemOperator interface {
	GetListOfApps(appsDir string) ([]string, error)
	GetPortAndPathOfApp(appDir string) (string, string, error)
	RunDockerCompose(appDir, port string) (string, error)
	ShutdownStackAndDeleteComposeFile(composeFilePath string) error

	GetDockerComposeFileContent(appDir string) ([]byte, error)
	WriteDockerComposeFileContent(appDir string, content []byte) error
	WriteServiceUpdatesIntoComposeFile(appDir string, serviceUpdates []ServiceUpdate) error
}

//go:generate mockery
type SingleAppUpdateFileSystemOperator interface {
	GetAppServices(appDir string) ([]Service, error)
	WriteNewTagToTheDockerCompose(appDir, serviceName, newTag string) error
}

//go:generate mockery
type EndpointChecker interface {
	TryAccessingIndexPageOnLocalhost(port string, path string) error
}

//go:generate mockery
type SingleAppUpdater interface {
	fetchAppUpdate(appDir string) (*AppUpdate, error)
}

//go:generate mockery
type HealthChecker interface {
	PerformHealthChecks() (*HealthCheckReport, error)
	ConductHealthcheckForSingleApp(app string) AppHealthReport
}

//go:generate mockery
type DockerHubClient interface {
	listImageTags(image string) ([]string, error)
}
