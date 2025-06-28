package main

import (
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

func TestReportHealthWorked(t *testing.T) {
	report := HealthCheckReport{
		AllAppsHealthy: true,
		AppHealthReports: []AppHealthReport{
			{
				AppName:      "sample-app",
				Healthy:      true,
				ErrorMessage: "",
			},
		},
	}
	output := reportHealth(report)
	assert.Equal(t, "sample-app: OK\nSummary: All apps are healthy", output)
}

func TestReportHealthFailed(t *testing.T) {
	report := HealthCheckReport{
		AllAppsHealthy: false,
		AppHealthReports: []AppHealthReport{
			{
				AppName:      "sample-app",
				Healthy:      false,
				ErrorMessage: "some-error",
			},
		},
	}
	output := reportHealth(report)
	assert.Equal(t, "sample-app: some-error\nSummary: Some apps are unhealthy", output)
}
