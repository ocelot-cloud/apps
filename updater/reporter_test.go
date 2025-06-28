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
	assert.Equal(t, "sample-app: OK\nSummary: All apps are healthy\n", output)
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
	assert.Equal(t, "sample-app: some-error\nSummary: Some apps are unhealthy\n", output)
}

func newAppHealth(n string, ok bool, e string) *AppHealthReport {
	if n == "" {
		return nil
	}
	return &AppHealthReport{AppName: n, Healthy: ok, ErrorMessage: e}
}

func newUpd(wasSuccessful, wasUpdateAvailable bool, appHealthReport *AppHealthReport, serviceUpdates []ServiceUpdate, err string) AppUpdateReport {
	var up *AppUpdate
	if wasUpdateAvailable {
		up = &AppUpdate{WasUpdateFound: wasUpdateAvailable, ServiceUpdates: serviceUpdates, ErrorMessage: err}
	}
	return AppUpdateReport{AppName: "sample-app", WasSuccessful: wasSuccessful, WasUpdateAvailable: wasUpdateAvailable, AppHealthReport: appHealthReport, AppUpdates: up, UpdateErrorMessage: err}
}

func TestReportUpdateWorked(t *testing.T) {
	r := UpdateReport{WasSuccessful: true, AppUpdateReport: []AppUpdateReport{
		newUpd(true, true, newAppHealth("sample-app", true, ""),
			[]ServiceUpdate{
				{ServiceName: "nginx", OldTag: "1.0", NewTag: "1.1"},
				{ServiceName: "nginx2", OldTag: "2.2", NewTag: "2.3"},
			}, ""),
	}}
	out := reportUpdate(r)
	assert.Equal(t, "sample-app: nginx 1.0->1.1; nginx2 2.2->2.3, Health: OK\nSummary: Update successful\n", out)
}

func TestReportUpdateFail(t *testing.T) {
	r := UpdateReport{WasSuccessful: false, AppUpdateReport: []AppUpdateReport{
		newUpd(false, true, nil, nil, "some-update-error"),
	}}
	out := reportUpdate(r)
	assert.Equal(t, "sample-app: update failed - some-update-error\nSummary: Update failed\n", out)
}

func TestReportUpdateNotAvailable(t *testing.T) {
	r := UpdateReport{WasSuccessful: true, AppUpdateReport: []AppUpdateReport{
		newUpd(true, false, nil, nil, ""),
	}}
	out := reportUpdate(r)
	assert.Equal(t, "sample-app: no update available\nSummary: Update successful\n", out)
}

func TestReportUpdateServiceFailed(t *testing.T) {
	r := UpdateReport{WasSuccessful: false, AppUpdateReport: []AppUpdateReport{
		newUpd(false, true, nil,
			[]ServiceUpdate{{ServiceName: "nginx", OldTag: "1.0", NewTag: "1.1"}}, "service update failed"),
	}}
	out := reportUpdate(r)
	assert.Equal(t, "sample-app: service update error - service update failed\nSummary: Update failed\n", out)
}

func TestReportUpdateWithHealthcheckFailed(t *testing.T) {
	r := UpdateReport{WasSuccessful: false, AppUpdateReport: []AppUpdateReport{
		newUpd(true, true, &AppHealthReport{
			AppName:      "sample-app2",
			Healthy:      false,
			ErrorMessage: "endpoint no available",
		},
			[]ServiceUpdate{{ServiceName: "nginx", OldTag: "1.0", NewTag: "1.1"}}, "service update failed"),
	}}
	out := reportUpdate(r)
	assert.Equal(t, "sample-app: nginx 1.0->1.1, Health: endpoint no available\nSummary: Update failed\n", out)
}
