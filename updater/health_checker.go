package main

func (u *HealthCheckerImpl) PerformHealthChecks() (*HealthCheckReport, error) {
	apps, err := u.fileSystemOperator.GetListOfApps(u.appsDir)
	if err != nil {
		return nil, err
	}
	report := &HealthCheckReport{
		AllAppsHealthy: true,
	}
	for _, app := range apps {
		appReport := u.ConductHealthcheckForSingleApp(app)
		if !appReport.Healthy {
			report.AllAppsHealthy = false
		}
		report.AppHealthReports = append(report.AppHealthReports, appReport)
	}
	return report, nil
}

func getAppHealthReportWithError(app, errorMessage string, err error) AppHealthReport {
	return AppHealthReport{
		AppName:      app,
		Healthy:      false,
		ErrorMessage: errorMessage + ": " + err.Error(),
	}
}

func (u *HealthCheckerImpl) ConductHealthcheckForSingleApp(app string) AppHealthReport {
	appDir := u.appsDir + "/" + app
	port, path, err := u.fileSystemOperator.GetPortAndPathOfApp(appDir)
	if err != nil {
		return getAppHealthReportWithError(app, "Failed to get port", err)
	}

	injectedComposeYamlPath, err := u.fileSystemOperator.RunDockerCompose(appDir, port)
	if err != nil {
		return getAppHealthReportWithError(app, "Failed to run docker-compose", err)
	}
	defer u.fileSystemOperator.ShutdownStackAndDeleteComposeFile(injectedComposeYamlPath)
	err = u.endpointChecker.TryAccessingIndexPageOnLocalhost(port, path)
	if err != nil {
		return getAppHealthReportWithError(app, "Failed to access index page", err)
	}
	return AppHealthReport{
		AppName:      app,
		Healthy:      true,
		ErrorMessage: "",
	}
}
