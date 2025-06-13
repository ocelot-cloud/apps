package main

import "path/filepath"

// ImageUpdater updates container image tags for an app before running it.
type ImageUpdater interface {
	UpdateImages(appDir string) *AppError
}

type basicImageUpdater struct {
	fs     FileSystemOperator
	docker DockerHubClient
}

func (u *basicImageUpdater) UpdateImages(appDir string) *AppError {
	services, err := u.fs.GetImagesOfApp(appDir)
	if err != nil {
		return &AppError{"Failed to get images of app", err}
	}
	for _, service := range services {
		tags, err := u.docker.listImageTags(service.Image)
		if err != nil {
			return &AppError{"Failed to get latest tags from Docker Hub for service " + service.Name, err}
		}
		newTag, found, err := FilterLatestImageTag(service.Tag, tags)
		if err != nil {
			return &AppError{"Failed to filter latest image tag for service " + service.Name, err}
		}
		if found {
			err = u.fs.WriteNewTagToDockerCompose(appDir, service.Name, newTag)
			if err != nil {
				return &AppError{"Failed to write new tag to docker-compose for service " + service.Name, err}
			}
		} else {
			logger.Info("No newer tag found for service " + service.Name + " in app " + filepath.Base(appDir))
		}
	}
	return nil
}
