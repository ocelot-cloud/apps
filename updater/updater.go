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

type Service struct {
	Name  string
	Image string // TODO not sure if needed
	Tag   string
}

type Updater struct {
	appsDir       string
	fs            FileSystemOperator
	healthChecker HealthChecker
	imageUpdater  ImageUpdater
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
func (u *Updater) conductLogic(updateBeforeCheck bool) (*HealthCheckReport, error) {
	apps, err := u.fs.GetListOfApps(u.appsDir)
	if err != nil {
		return nil, err
	}
	report := &HealthCheckReport{
		AllAppsHealthy: true,
	}
	for _, app := range apps {
		appDir := u.appsDir + "/" + app

		if updateBeforeCheck {
			if opErr := u.imageUpdater.UpdateImages(appDir); opErr != nil {
				addErrorToReport(report, app, opErr.message, opErr.err)
				continue
			}
		}

		if opErr := u.healthChecker.CheckApp(appDir); opErr != nil {
			addErrorToReport(report, app, opErr.message, opErr.err)
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
