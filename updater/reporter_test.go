package main

import (
	"fmt"
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
	assertHealthReport(t, "- sample-app: OK\n", output, true)
}

func assertHealthReport(t *testing.T, expectedContent, actualReport string, wasHealthy bool) {
	var expectedSummary string
	if wasHealthy {
		expectedSummary = "All apps are healthy"
	} else {
		expectedSummary = "Some apps are unhealthy"
	}
	expectedReport := fmt.Sprintf("\nHealth Check Report:\n\n"+expectedContent+"\nSummary: %s\n", expectedSummary)
	assert.Equal(t, expectedReport, actualReport)

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
	assertHealthReport(t, "- sample-app: some-error\n", output, false)
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
	actualReport := reportUpdate(r)
	expectedContent := `
- sample-app: update worked
  - nginx: 1.0 -> 1.1
  - nginx2: 2.2 -> 2.3
`
	assertUpdateReport(t, expectedContent, actualReport, true)
}

func assertUpdateReport(t *testing.T, expectedContent, actualReport string, wasUpdateSuccessful bool) {
	var expectedSummary string
	if wasUpdateSuccessful {
		expectedSummary = "Update successful"
	} else {
		expectedSummary = "Update failed"
	}
	expectedReport := fmt.Sprintf("\nUpdate Report:\n"+expectedContent+"\nSummary: %s\n", expectedSummary)
	assert.Equal(t, expectedReport, actualReport)
}

func TestReportUpdateFail(t *testing.T) {
	r := UpdateReport{WasSuccessful: false, AppUpdateReport: []AppUpdateReport{
		newUpd(false, true, nil, nil, "some-update-error"),
	}}
	out := reportUpdate(r)
	expectedContent := `
- sample-app: update failed - some-update-error
`
	assertUpdateReport(t, expectedContent, out, false)
}

func TestReportUpdateNotAvailable(t *testing.T) {
	r := UpdateReport{WasSuccessful: true, AppUpdateReport: []AppUpdateReport{
		newUpd(true, false, nil, nil, ""),
	}}
	out := reportUpdate(r)
	expectedContent := `
- sample-app: no update available
`
	assertUpdateReport(t, expectedContent, out, true)
}

func TestReportUpdateServiceFailed(t *testing.T) {
	r := UpdateReport{WasSuccessful: false, AppUpdateReport: []AppUpdateReport{
		newUpd(false, true, nil,
			[]ServiceUpdate{{ServiceName: "nginx", OldTag: "1.0", NewTag: "1.1"}}, "service update failed"),
	}}
	out := reportUpdate(r)
	expectedContent := `
- sample-app: service update error - service update failed
`
	assertUpdateReport(t, expectedContent, out, false)
}

func TestReportUpdateWithHealthcheckFailed(t *testing.T) {
	r := UpdateReport{WasSuccessful: false, AppUpdateReport: []AppUpdateReport{
		newUpd(true, true, &AppHealthReport{
			AppName:      "sample-app2", // to be ignored in report as it is already in app update report
			Healthy:      false,
			ErrorMessage: "endpoint not available",
		},
			[]ServiceUpdate{{ServiceName: "nginx", OldTag: "1.0", NewTag: "1.1"}}, "service update failed"),
	}}
	out := reportUpdate(r)
	expectedContent := `
- sample-app: update failed - endpoint not available
  - nginx: 1.0 -> 1.1
`
	assertUpdateReport(t, expectedContent, out, false)
}
