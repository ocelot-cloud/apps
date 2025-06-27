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
	update(appDir string) error
}

type SingleAppUpdaterReal struct {
	fsOperator      SingleAppUpdateFileSystemOperator
	dockerHubClient DockerHubClient
}

/*
	TODO !! cases:

- not updated -> no healthcheck necessary
- updated -> healthcheck 1) successful, 2) failed
*/
type AppUpdate struct {
	AppDir         string
	WasUpdateFound bool
	// TODO apply update, do healthcheck; if passed keep update, else revert update
	ServiceUpdates []ServiceUpdate
}

type ServiceUpdate struct {
	ServiceName string
	OldTag      string
	NewTag      string
}

func (a *SingleAppUpdaterReal) update(appDir string) (*AppUpdate, error) {
	services, err := a.fsOperator.GetImagesOfApp(appDir)
	if err != nil {
		return nil, err
	}

	var appUpdate = &AppUpdate{}
	appUpdate.AppDir = appDir
	var serviceUpdates []ServiceUpdate
	for _, service := range services {

		latestTagsFromDockerHub, err := a.dockerHubClient.listImageTags(service.Image)
		if err != nil {
			return nil, err
		}
		newTag, err := FilterLatestImageTag(service.Tag, latestTagsFromDockerHub)
		if err != nil {
			return nil, err
		}

		if newTag != "" {
			newUpdate := ServiceUpdate{
				ServiceName: service.Name,
				OldTag:      service.Tag,
				NewTag:      newTag,
			}
			serviceUpdates = append(serviceUpdates, newUpdate)
		}
	}

	if len(serviceUpdates) > 0 {
		appUpdate.ServiceUpdates = serviceUpdates
		appUpdate.WasUpdateFound = true
	}

	return appUpdate, nil
}

type Service struct {
	Name  string
	Image string
	Tag   string
}

type Updater struct {
	appsDir            string
	fileSystemOperator FileSystemOperator
	endpointChecker    EndpointChecker
	appUpdater         SingleAppUpdater
}

type HealthCheckReport struct {
	AllAppsHealthy bool
	AppReports     []AppHealthReport
}

type AppHealthReport struct {
	AppName      string
	Healthy      bool
	ErrorMessage string
}

func (u *Updater) PerformHealthCheck() (*HealthCheckReport, error) {
	return u.conductLogic(false)
}

func getAppHealthReport(app, errorMessage string, err error) AppHealthReport {
	return AppHealthReport{
		AppName:      app,
		Healthy:      false,
		ErrorMessage: errorMessage + ": " + err.Error(),
	}
}

func (u *Updater) PerformUpdate() (*HealthCheckReport, error) {
	return u.conductLogic(true)
}

// TODO add argument []string, if empty perform on all apps, otherwise only on specified apps
func (u *Updater) conductLogic(conductTagUpdatesBeforeHealthcheck bool) (*HealthCheckReport, error) {
	apps, err := u.fileSystemOperator.GetListOfApps(u.appsDir)
	if err != nil {
		return nil, err
	}
	report := &HealthCheckReport{
		AllAppsHealthy: true,
	}
	for _, app := range apps {
		appReport := u.conductLogicForSingleApp(conductTagUpdatesBeforeHealthcheck, app)
		if !appReport.Healthy {
			report.AllAppsHealthy = false
		}
		report.AppReports = append(report.AppReports, appReport)
	}
	return report, nil
}

func (u *Updater) conductLogicForSingleApp(conductTagUpdatesBeforeHealthcheck bool, app string) AppHealthReport {
	appDir := u.appsDir + "/" + app
	var originalDockerComposeContent []byte
	if conductTagUpdatesBeforeHealthcheck {
		var err error
		originalDockerComposeContent, err = u.fileSystemOperator.GetDockerComposeFileContent(appDir)
		if err != nil {
			return getAppHealthReport(app, "Failed to get docker-compose file content", err)
		}

		err = u.appUpdater.update(appDir)
		if err != nil {
			u.resetDockerComposeYamlToInitialContent(appDir, originalDockerComposeContent)
			return getAppHealthReport(app, "Failed to update app", err)
		}
	}
	port, err := u.fileSystemOperator.GetPortOfApp(appDir)
	if err != nil {
		return getAppHealthReport(app, "Failed to get port", err)
	}
	err = u.fileSystemOperator.InjectPortInDockerCompose(appDir)
	if err != nil {
		return getAppHealthReport(app, "Failed to inject port in docker-compose", err)
	}

	err = u.fileSystemOperator.RunInjectedDockerCompose(appDir)
	if err != nil {
		return getAppHealthReport(app, "Failed to run docker-compose", err)
	}
	err = u.endpointChecker.TryAccessingIndexPageOnLocalhost(port)
	if err != nil {
		return getAppHealthReport(app, "Failed to access index page", err)
	}
	return AppHealthReport{
		AppName:      app,
		Healthy:      true,
		ErrorMessage: "",
	}
}

func (u *Updater) resetDockerComposeYamlToInitialContent(appDir string, originalDockerComposeContent []byte) {
	err := u.fileSystemOperator.WriteDockerComposeFileContent(appDir, originalDockerComposeContent)
	if err != nil {
		panic("Failed to write docker-compose file back to original state: " + err.Error())
	}
}

// TODO implement real file system operator and endpoint checker
// TODO in real implementation, I will create a "docker-compose-injected.yml" that will be deleted at the end; make an assertion that there is no such file at the end
