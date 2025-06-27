package main

//go:generate mockery
type FileSystemOperator interface {
	// needed for healthchecks and updates
	GetListOfApps(appsDir string) ([]string, error)
	GetPortOfApp(appDir string) (string, error)
	InjectPortInDockerCompose(appDir string) error
	RunInjectedDockerCompose(appDir string) error
	GetDockerComposeFileContent(appDir string) ([]byte, error)
	WriteDockerComposeFileContent(appDir string, content []byte) error
}

//go:generate mockery
type SingleAppUpdateFileSystemOperator interface {
	GetImagesOfApp(appDir string) ([]Service, error)
	WriteNewTagToDockerCompose(appDir, serviceName, newTag string) error
}

//go:generate mockery
type EndpointChecker interface {
	TryAccessingIndexPageOnLocalhost(port string) error
}

//go:generate mockery
type SingleAppUpdater interface {
	update(appDir string) (*AppUpdate, error)
}
