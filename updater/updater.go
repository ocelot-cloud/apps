package main

/* TODO healthcheck
get app directories
iterate over app directories
  inject port from app yaml in docker-compose file
  run app
  run healthcheck, if not healthy, mark in report and continue
print summary
*/

//go:generate mockery
type FileSystemOperator interface {
	GetListOfApps(appsDir string) ([]string, error)
	InjectPortInDockerCompose(appDir string) error
	RunDockerCompose(appDir string) error
}

type Updater struct {
	appsDir            string
	fileSystemOperator FileSystemOperator
	dockerHubClient    DockerHubClient
}

// TODO update, same as above, but we update images before performing healthcheck
