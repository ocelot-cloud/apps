package main

import "path/filepath"

// HealthChecker validates if an app is healthy by starting it and checking the endpoint.
type HealthChecker interface {
	CheckApp(appDir string) *AppError
}

type basicHealthChecker struct {
	fs      FileSystemOperator
	checker EndpointChecker
}

func (h *basicHealthChecker) CheckApp(appDir string) *AppError {
	port, err := h.fs.GetPortOfApp(appDir)
	if err != nil {
		return &AppError{"Failed to get port", err}
	}
	if err = h.fs.InjectPortInDockerCompose(appDir); err != nil {
		return &AppError{"Failed to inject port in docker-compose", err}
	}
	if err = h.fs.RunInjectedDockerCompose(appDir); err != nil {
		return &AppError{"Failed to run docker-compose", err}
	}
	if err = h.checker.TryAccessingIndexPageOnLocalhost(port); err != nil {
		return &AppError{"Failed to access index page", err}
	}
	logger.Info("Health check successful for app " + filepath.Base(appDir))
	return nil
}
