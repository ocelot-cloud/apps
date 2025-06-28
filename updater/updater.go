package main

// TODO dont pass the appDir in any function. Just pass "app" and construct the appDir from the global appDir inside the methods
func (a *SingleAppUpdaterImpl) update(appDir string) (*AppUpdate, error) {
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
	appUpdate := &AppUpdate{
		WasUpdateFound: false,
		ServiceUpdates: []ServiceUpdate{},
	}

	var originalDockerComposeContent []byte
	var err error
	originalDockerComposeContent, err = u.fileSystemOperator.GetDockerComposeFileContent(appDir)
	if err != nil {
		report := getEmptyReport()
		report.UpdateErrorMessage = "Failed to get docker-compose file content: " + err.Error()
		return report
	}

	appUpdate, err = u.appUpdater.update(appDir)
	if err != nil {
		// TODO no need to reset the docker-compose file here, because it was never changed
		u.resetDockerComposeYamlToInitialContent(appDir, originalDockerComposeContent)
		report := getEmptyReport()
		report.UpdateErrorMessage = "Failed to update app: " + err.Error()
		return report
	}

	if !appUpdate.WasUpdateFound {
		report := getEmptyReport()
		report.WasSuccessful = true
		return report
	}

	// TODO !! I need to ensure at the beginning that there are no git diffs at least within the appsDir, because potential broken but up to date apps will not be healthchecked
	// TODO !! I am still missing to write the changes to the docker-compose file

	updatedDockerComposeContent, err := updateDockerComposeContent(originalDockerComposeContent, appUpdate.ServiceUpdates)
	if err != nil {
		// TODO To be covered
		u.resetDockerComposeYamlToInitialContent(appDir, originalDockerComposeContent)
		report := getEmptyReport()
		report.UpdateErrorMessage = "Failed to update docker-compose file content: " + err.Error()
		return report
	}
	println(string(updatedDockerComposeContent))

	appHealthReport := u.healthChecker.ConductHealthcheckForSingleApp(app)
	// TODO if app health report is not successful, I should reset the docker-compose file to the original content
	return AppUpdateReport{
		WasSuccessful:      true,
		WasUpdateAvailable: true,
		AppHealthReport:    &appHealthReport,
		AppUpdates:         appUpdate,
		UpdateErrorMessage: "",
	}
}

func updateDockerComposeContent(content []byte, updates []ServiceUpdate) ([]byte, error) {
	return nil, nil
}

func getEmptyReport() AppUpdateReport {
	return AppUpdateReport{
		AppName:            "sampleapp",
		WasSuccessful:      false,
		WasUpdateAvailable: false,
		AppHealthReport:    nil,
		AppUpdates:         nil,
		UpdateErrorMessage: "",
	}
}

// TODO can be inlined I guess
func (u *Updater) resetDockerComposeYamlToInitialContent(appDir string, originalDockerComposeContent []byte) {
	err := u.fileSystemOperator.WriteDockerComposeFileContent(appDir, originalDockerComposeContent)
	if err != nil {
		panic("Failed to write docker-compose file back to original state: " + err.Error())
	}
}
