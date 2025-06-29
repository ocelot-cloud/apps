package main

func (a *SingleAppUpdaterImpl) fetchAppUpdate(appDir string) (*AppUpdate, error) {
	services, err := a.fsOperator.GetAppServices(appDir)
	if err != nil {
		return nil, err
	}

	var appUpdate = &AppUpdate{}
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

func (u *Updater) PerformUpdate() (*UpdateReport, error) {
	apps, err := u.fileSystemOperator.GetListOfApps(appsDir)
	if err != nil {
		return nil, err
	}
	report := &UpdateReport{
		WasSuccessful: true,
	}
	for _, app := range apps {
		appUpdateReport := u.conductUpdateForSingleApp(app)
		appUpdateReport.AppName = app
		if !appUpdateReport.WasSuccessful {
			report.WasSuccessful = false
		}
		report.AppUpdateReport = append(report.AppUpdateReport, appUpdateReport)
	}
	return report, nil
}

func (u *Updater) conductUpdateForSingleApp(app string) AppUpdateReport {
	appDir := appsDir + "/" + app

	// TODO !! I need to ensure at the beginning that there are no git diffs at least within the appsDir, because potential broken but up to date apps will not be healthchecked
	originalContent, err := u.fileSystemOperator.GetDockerComposeFileContent(appDir)
	if err != nil {
		report := getEmptyUpdateReport()
		report.UpdateErrorMessage = "Failed to read docker-compose.yml: " + err.Error()
		return report
	}

	appUpdate, err := u.updateApplier.ApplyUpdate(appDir)
	if err != nil {
		// TODO !! block not covered yet
		u.resetDockerComposeYamlToInitialContent(appDir, originalContent)
		report := getEmptyUpdateReport()
		report.UpdateErrorMessage = "Failed to apply update to docker-compose.yml: " + err.Error()
		return report
	}

	if !appUpdate.WasUpdateFound {
		report := getEmptyUpdateReport()
		report.WasSuccessful = true
		return report
	}

	appHealthReport := u.healthChecker.ConductHealthcheckForSingleApp(app)
	if !appHealthReport.Healthy {
		u.resetDockerComposeYamlToInitialContent(appDir, originalContent)
		report := getEmptyUpdateReport()
		report.UpdateErrorMessage = "App health check failed: " + appHealthReport.ErrorMessage
		return report
	}
	return AppUpdateReport{
		WasSuccessful:      true,
		WasUpdateAvailable: true,
		AppHealthReport:    &appHealthReport,
		AppUpdates:         appUpdate,
		UpdateErrorMessage: "",
	}
}

func getEmptyUpdateReport() AppUpdateReport {
	return AppUpdateReport{
		AppName:            "sampleapp",
		WasSuccessful:      false,
		WasUpdateAvailable: false,
		AppHealthReport:    nil,
		AppUpdates:         nil,
		UpdateErrorMessage: "",
	}
}

func (u *Updater) resetDockerComposeYamlToInitialContent(appDir string, originalDockerComposeContent []byte) {
	err := u.fileSystemOperator.WriteDockerComposeFileContent(appDir, originalDockerComposeContent)
	if err != nil {
		panic("Failed to write docker-compose file back to original state, which should never happen: " + err.Error())
	}
}
