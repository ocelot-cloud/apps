package main

//go:generate mockery
type FileSystemOperator interface {
	// needed for healthchecks and updates
	GetListOfApps(appsDir string) ([]string, error)
	GetPortOfApp(appDir string) (string, error)
	InjectPortInDockerCompose(appDir string) error
	RunInjectedDockerCompose(appDir string) error
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

func (a *SingleAppUpdaterReal) update(appDir string) (bool, error) {
	services, err := a.fsOperator.GetImagesOfApp(appDir)
	if err != nil {
		return false, err
	}

	wasAnyServiceUpdated := false
	for _, service := range services {
		latestTagsFromDockerHub, err := a.dockerHubClient.listImageTags(service.Image)
		if err != nil {
			return false, err
		}
		newTag, wasNewerTagFound, err := FilterLatestImageTag(service.Tag, latestTagsFromDockerHub)
		if err != nil {
			return false, err
		}
		if wasNewerTagFound {
			err = a.fsOperator.WriteNewTagToDockerCompose(appDir, service.Name, newTag)
			if err != nil {
				return false, err
			}
			wasAnyServiceUpdated = true
		}
	}
	if wasAnyServiceUpdated {
		return true, nil
	} else {
		logger.Info("No updates found for app at: %s", appDir)
		return false, nil
	}
}

type Service struct {
	Name  string
	Image string // TODO not sure if needed
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

func addErrorToReport(report *HealthCheckReport, app, errorMessage string, err error) {
	report.AllAppsHealthy = false
	report.AppReports = append(report.AppReports, AppHealthReport{
		AppName:      app,
		Healthy:      false,
		ErrorMessage: errorMessage + ": " + err.Error(),
	})
}

// TODO update, same as above, but we update images before performing healthcheck
// TODO resolve duplication
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
		appDir := u.appsDir + "/" + app
		if conductTagUpdatesBeforeHealthcheck {
			err := u.appUpdater.update(appDir)
			if err != nil {
				addErrorToReport(report, app, "Failed to update app", err)
				continue
			}
		}

		port, err := u.fileSystemOperator.GetPortOfApp(appDir)
		if err != nil {
			addErrorToReport(report, app, "Failed to get port", err)
			continue
		}
		err = u.fileSystemOperator.InjectPortInDockerCompose(appDir)
		if err != nil {
			addErrorToReport(report, app, "Failed to inject port in docker-compose", err)
			continue
		}

		err = u.fileSystemOperator.RunInjectedDockerCompose(appDir)
		if err != nil {
			addErrorToReport(report, app, "Failed to run docker-compose", err)
			continue
		}
		err = u.endpointChecker.TryAccessingIndexPageOnLocalhost(port)
		if err != nil {
			addErrorToReport(report, app, "Failed to access index page", err)
			// TODO if conductTagUpdatesBeforeHealthcheck -> set tag back to previous version
			continue
		}
		report.AppReports = append(report.AppReports, AppHealthReport{
			AppName:      app,
			Healthy:      true,
			ErrorMessage: "",
		})
	}
	return report, nil
}

// TODO implement real file system operator and endpoint checker
