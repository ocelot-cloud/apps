package main

import "path/filepath"

func (a *AppUpdateFetcherImpl) Fetch(appDir string) (*AppUpdate, error) {
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
			logger.Info("Found new tag for service '%s': %s -> %s", service.Name, service.Tag, newTag)
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

	if appUpdate.WasUpdateFound {
		mainService := filepath.Base(appDir)
		var wasUpdateOfMainServiceFound bool
		for _, serviceUpdate := range serviceUpdates {
			if serviceUpdate.ServiceName == mainService {
				wasUpdateOfMainServiceFound = true
				break
			}
		}

		// We only apply an update when the apps main service is upgraded. For example, in a Gitea stack, changing the database container alone is ignored; an update only counts if the Gitea service (i.e. the web UI) is upgraded. This rule exists because the Ocelot app package is named after the main service’s version, e.g. 1.2.3.zip corresponds to gitea/gitea:1.2.3 in docker-compose.yml. Updating secondary services without changing the tag of the main service would leave the package version unchanged, and we must not publish multiple packages with the same version number.
		if !wasUpdateOfMainServiceFound {
			appUpdate.ServiceUpdates = nil
			appUpdate.WasUpdateFound = false
		}
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
		logger.Info("Performing update for app '%s'", app)
		appUpdateReport := u.conductUpdateForSingleApp(app)
		appUpdateReport.AppName = app
		if !appUpdateReport.WasSuccessful {
			report.WasSuccessful = false
		}
		logger.Info("Was update process for app '%s' successful: %t", app, appUpdateReport.WasSuccessful)
		report.AppUpdateReport = append(report.AppUpdateReport, appUpdateReport)
	}
	return report, nil
}

func (u *Updater) conductUpdateForSingleApp(app string) AppUpdateReport {
	appDir := appsDir + "/" + app

	// TODO I need to ensure at the beginning that there are no git diffs at least within the appsDir, because potential broken but up to date apps will not be healthchecked
	originalContent, err := u.fileSystemOperator.GetDockerComposeFileContent(appDir)
	if err != nil {
		report := getEmptyAppUpdateReport()
		report.UpdateErrorMessage = "Failed to read docker-compose.yml: " + err.Error()
		return report
	}

	appUpdate, err := u.updateApplier.ApplyUpdate(appDir)
	if err != nil {
		u.resetDockerComposeYamlToInitialContent(appDir, originalContent)
		report := getEmptyAppUpdateReport()
		report.UpdateErrorMessage = "Failed to apply update to docker-compose.yml: " + err.Error()
		return report
	}

	if !appUpdate.WasUpdateFound {
		logger.Info("No update available for app '%s'", app)
		report := getEmptyAppUpdateReport()
		report.WasSuccessful = true
		return report
	}

	appHealthReport := u.healthChecker.ConductHealthcheckForSingleApp(app)
	logger.Info("Was app healthy after update: %v", appHealthReport)
	if !appHealthReport.Healthy {
		u.resetDockerComposeYamlToInitialContent(appDir, originalContent)
		report := getEmptyAppUpdateReport()
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

func getEmptyAppUpdateReport() AppUpdateReport {
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
