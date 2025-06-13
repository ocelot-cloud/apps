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
	appsDir            string
	fileSystemOperator FileSystemOperator
	endpointChecker    EndpointChecker
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
	apps, err := u.fileSystemOperator.GetListOfApps(u.appsDir)
	if err != nil {
		// TODO to be covered
		return nil, err
	}

	report := HealthCheckReport{
		AllAppsHealthy: true,
	}
	for _, app := range apps {
		appDir := u.appsDir + "/" + app
		port, err := u.fileSystemOperator.GetPortOfApp(appDir)
		if err != nil {
			report.AllAppsHealthy = false
			report.AppReports = append(report.AppReports, AppHealthReport{
				AppName:      app,
				Healthy:      false,
				ErrorMessage: "Failed to get port: " + err.Error(),
			})
			continue
		}
		err = u.fileSystemOperator.InjectPortInDockerCompose(appDir)
		if err != nil {
			// TODO write to report
			continue
		}
		err = u.fileSystemOperator.RunInjectedDockerCompose(appDir)
		if err != nil {
			// TODO write to report
			continue
		}
		err = u.endpointChecker.TryAccessingIndexPageOnLocalhost(port)
		if err != nil {
			// TODO write to report
			continue
		}
		report.AppReports = append(report.AppReports, AppHealthReport{
			AppName:      app,
			Healthy:      true,
			ErrorMessage: "",
		})
	}
	return &report, nil
}

// TODO update, same as above, but we update images before performing healthcheck
func (u *Updater) PerformUpdate() (*HealthCheckReport, error) {
	return nil, nil
}
