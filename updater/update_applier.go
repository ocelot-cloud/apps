package main

type UpdateApplierImpl struct {
	appUpdater         SingleAppUpdater
	fileSystemOperator FileSystemOperator
}

// TODO write tests for this
// TODO NewUpdateApplier function should take dependencies as parameters
func (u *UpdateApplierImpl) ApplyUpdate(appDir string) (*AppUpdate, error) {
	appUpdate, err := u.appUpdater.fetchAppUpdate(appDir)
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
