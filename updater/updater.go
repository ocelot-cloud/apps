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
	// needed for healthchecks and updates
	GetListOfApps(appsDir string) ([]string, error)
	GetPortOfApp(appDir string) (string, error)
	InjectPortInDockerCompose(appDir string) error
	RunInjectedDockerCompose(appDir string) error

	// needed for updates
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
	fileSystemOperator FileSystemOperator
	dockerHubClient    DockerHubClient
}

func (a *SingleAppUpdaterReal) update(appDir string) error {
	services, err := a.fileSystemOperator.GetImagesOfApp(appDir)
	if err != nil {
		// TODO "Failed to get images of app"
		return err
	}

	for _, service := range services {
		latestTagsFromDockerHub, err := a.dockerHubClient.listImageTags(service.Image)
		if err != nil {
			// TODO "Failed to get latest tags from Docker Hub for service "+service.Name
			return err
		}
		newTag, wasNewerTagFound, err := FilterLatestImageTag(service.Tag, latestTagsFromDockerHub)
		if err != nil {
			// TODO "Failed to filter latest image tag for service "+service.Name
			return err
		}
		if wasNewerTagFound {
			err = a.fileSystemOperator.WriteNewTagToDockerCompose(appDir, service.Name, newTag)
			if err != nil {
				// TODO "Failed to write new tag to docker-compose for service "+service.Name
				return err
			}
		} else {
			logger.Info("No newer tag found for service " + service.Name + " in app dir: " + appDir)
		}
	}
	return nil
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
