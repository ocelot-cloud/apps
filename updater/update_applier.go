package main

//go:generate mockery
type UpdateApplier interface {
	ApplyUpdate(appDir string) (*AppUpdate, error)
}

type UpdateApplierImpl struct{}

func (u *UpdateApplierImpl) ApplyUpdate(appDir string) (*AppUpdate, error) {
	/* TODO !!
	appUpdate, err = u.appUpdater.fetchAppUpdate(appDir)
	if err != nil {
		report := getEmptyUpdateReport()
		report.UpdateErrorMessage = "Failed to fetchAppUpdate app: " + err.Error()
		return report
	}

	if !appUpdate.WasUpdateFound {
		report := getEmptyUpdateReport()
		report.WasSuccessful = true
		return report
	}

	err = u.fileSystemOperator.WriteServiceUpdatesIntoComposeFile(appDir, appUpdate.ServiceUpdates)
	if err != nil {
		u.resetDockerComposeYamlToInitialContent(appDir, originalContent) // TODO ensure tests fail with this
		report := getEmptyUpdateReport()
		report.UpdateErrorMessage = "Failed to apply app fetchAppUpdate to docker-compose.yml: " + err.Error()
		return report
	}
	*/
	return nil, nil
}
