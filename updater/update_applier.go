package main

type UpdateApplierImpl struct {
	appUpdater         AppUpdateFetcher
	fileSystemOperator FileSystemOperator
}

// TODO write tests for this
func (u *UpdateApplierImpl) ApplyUpdate(appDir string) (*AppUpdate, error) {
	appUpdate, err := u.appUpdater.Fetch(appDir)
	if err != nil {
		return nil, err
	}

	if !appUpdate.WasUpdateFound {
		return appUpdate, nil
	}

	err = u.fileSystemOperator.WriteServiceUpdatesIntoComposeFile(appDir, appUpdate.ServiceUpdates)
	if err != nil {
		return nil, err
	}
	return appUpdate, nil
}
