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

/* TODO additional cases for reportUpdate:
1) update was available but failed (no health report present, so skip this info), fill UpdateErrorMessage
2) no update available (counts as success, also no healthcheck present/necessary then)
3) update was available, but updating a specific service failed so fill ErrorMessage in ServiceUpdate (no healthcheck present); UpdateErrorMessage should be sth like "error occurred when updating a service" or so
*/

func newAppHealth(n string, ok bool, e string) *AppHealthReport {
	if n == "" {
		return nil
	}
	return &AppHealthReport{AppName: n, Healthy: ok, ErrorMessage: e}
}

func newUpd(ok, avail bool, h *AppHealthReport, svc []ServiceUpdate, err string) AppUpdateReport {
	var up *AppUpdate
	if avail {
		up = &AppUpdate{WasUpdateFound: avail, ServiceUpdates: svc, ErrorMessage: err}
	}
	return AppUpdateReport{AppName: "sample-app", WasSuccessful: ok, WasUpdateAvailable: avail, AppHealthReport: h, AppUpdates: up, UpdateErrorMessage: err}
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
	assert.Equal(t, "sample-app: nginx 1.0->1.1; nginx2 2.2->2.3, Health: OK\nSummary: Update successful", out)
}

/* TODO !!
func TestReportUpdateFail(t *testing.T) {
	r := UpdateReport{WasSuccessful: false, AppUpdateReport: []AppUpdateReport{
		newUpd(false, true, nil, nil, "some-update-error"),
	}}
	out := reportUpdate(r)
	assert.Equal(t, "sample-app: update failed - some-update-error\nSummary: Update failed", out)
}
*/

func TestReportUpdateNotAvailable(t *testing.T) {
	r := UpdateReport{WasSuccessful: true, AppUpdateReport: []AppUpdateReport{
		newUpd(true, false, nil, nil, ""),
	}}
	out := reportUpdate(r)
	assert.Equal(t, "sample-app: no update available\nSummary: Update successful", out)
}

func TestReportUpdateServiceFailed(t *testing.T) {
	r := UpdateReport{WasSuccessful: false, AppUpdateReport: []AppUpdateReport{
		newUpd(false, true, nil,
			[]ServiceUpdate{{ServiceName: "nginx", OldTag: "1.0", NewTag: "1.1"}}, "service update failed"),
	}}
	out := reportUpdate(r)
	assert.Equal(t, "sample-app: service update error - service update failed\nSummary: Update failed", out)
}
