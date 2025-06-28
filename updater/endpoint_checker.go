package main

type EndpointCheckerImpl struct{}

// TODO add path (from app yaml), expected content (from updater.yml), add timeout (two minutes?)
func (e EndpointCheckerImpl) TryAccessingIndexPageOnLocalhost(port string, path string) error {
	//TODO implement me
	panic("implement me")
}
