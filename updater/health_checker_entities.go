package main

type HealthCheckerImpl struct {
	appsDir            string
	fileSystemOperator FileSystemOperator
	endpointChecker    EndpointChecker
}

type HealthCheckReport struct {
	AllAppsHealthy   bool
	AppHealthReports []AppHealthReport
}

type AppHealthReport struct {
	AppName      string
	Healthy      bool
	ErrorMessage string
}
