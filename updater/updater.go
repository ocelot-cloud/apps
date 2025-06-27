package main

type SingleAppUpdaterReal struct {
	fsOperator      SingleAppUpdateFileSystemOperator
	dockerHubClient DockerHubClient
}

func (a *SingleAppUpdaterReal) update(appDir string) (*AppUpdate, error) {
	services, err := a.fsOperator.GetImagesOfApp(appDir)
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

// TODO must return an update report
func (u *Updater) PerformUpdate() (*UpdateReport, error) {
	apps, err := u.fileSystemOperator.GetListOfApps(u.appsDir)
	if err != nil {
		return nil, err
	}
	report := &UpdateReport{
		WasSuccessful: true,
	}
	for _, app := range apps {
		appUpdateReport := u.conductUpdateForSingleApp(app)
		if !appUpdateReport.WasSuccessful {
			report.WasSuccessful = false
		}
		report.AppUpdateReport = append(report.AppUpdateReport, appUpdateReport)
	}
	return report, nil
}

// TODO add argument []string, if empty perform on all apps, otherwise only on specified apps

func (u *Updater) conductUpdateForSingleApp(app string) AppUpdateReport {
	appDir := u.appsDir + "/" + app
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

	appHealthReport := u.healthChecker.ConductHealthcheckForSingleApp(app)
	return AppUpdateReport{
		WasSuccessful:      true,
		WasUpdateAvailable: true,
		AppHealthReport:    &appHealthReport,
		AppUpdates:         appUpdate,
		UpdateErrorMessage: "",
	}
}

func getEmptyReport() AppUpdateReport {
	return AppUpdateReport{
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
		panic("Failed to write docker-compose file back to original state: " + err.Error())
	}
}

// TODO implement real file system operator and endpoint checker
// TODO in real implementation, I will create a "docker-compose-injected.yml" that will be deleted at the end; make an assertion that there is no such file at the end
