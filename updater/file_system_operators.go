package main

type FileSystemOperatorImpl struct{}

func (f FileSystemOperatorImpl) GetListOfApps(appsDir string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (f FileSystemOperatorImpl) GetPortOfApp(appDir string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (f FileSystemOperatorImpl) InjectPortInDockerCompose(appDir string) error {
	//TODO implement me
	panic("implement me")
}

func (f FileSystemOperatorImpl) RunInjectedDockerCompose(appDir string) error {
	//TODO implement me
	panic("implement me")
}

func (f FileSystemOperatorImpl) GetDockerComposeFileContent(appDir string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (f FileSystemOperatorImpl) WriteDockerComposeFileContent(appDir string, content []byte) error {
	//TODO implement me
	panic("implement me")
}

type SingleAppUpdateFileSystemOperatorImpl struct{}

func (s *SingleAppUpdateFileSystemOperatorImpl) GetImagesOfApp(appDir string) ([]Service, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SingleAppUpdateFileSystemOperatorImpl) WriteNewTagToDockerCompose(appDir, serviceName, newTag string) error {
	//TODO implement me
	panic("implement me")
}
