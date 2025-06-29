package main

import (
	"fmt"
	"strings"
)

func reportHealth(h HealthCheckReport) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "\nHealth Check Report:\n\n")
	for _, a := range h.AppHealthReports {
		if a.Healthy {
			fmt.Fprintf(&builder, "- %s: OK\n", a.AppName)
		} else {
			fmt.Fprintf(&builder, "- %s: %s\n", a.AppName, a.ErrorMessage)
		}
	}
	if h.AllAppsHealthy {
		builder.WriteString("\nSummary: All apps are healthy\n")
	} else {
		builder.WriteString("\nSummary: Some apps are unhealthy\n")
	}
	return builder.String()
}

// TODO make pretty
func reportUpdate(updateReport UpdateReport) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "\nUpdate Report:\n\n")
	for _, appUpdateReport := range updateReport.AppUpdateReport {
		addUpdateReportLine(appUpdateReport, &builder)
	}
	if updateReport.WasSuccessful {
		fmt.Fprintf(&builder, "\nSummary: Update successful\n")
	} else {
		fmt.Fprintf(&builder, "\nSummary: Update failed\n")
	}
	return builder.String()
}

func addUpdateReportLine(a AppUpdateReport, b *strings.Builder) {
	switch {
	case !a.WasSuccessful:
		if a.AppUpdates != nil && len(a.AppUpdates.ServiceUpdates) > 0 && a.AppUpdates.ErrorMessage != "" {
			fmt.Fprintf(b, "- %s: service fetchAppUpdate error - %s\n", a.AppName, a.AppUpdates.ErrorMessage)
		} else {
			fmt.Fprintf(b, "- %s: fetchAppUpdate failed - %s\n", a.AppName, a.UpdateErrorMessage)
		}
	case !a.WasUpdateAvailable:
		fmt.Fprintf(b, "- %s: no fetchAppUpdate available\n", a.AppName)
	default:
		parts := make([]string, 0, len(a.AppUpdates.ServiceUpdates))
		for _, s := range a.AppUpdates.ServiceUpdates {
			parts = append(parts, fmt.Sprintf("%s: %s -> %s", s.ServiceName, s.OldTag, s.NewTag))
		}
		health := ""
		if a.AppHealthReport != nil {
			if a.AppHealthReport.Healthy {
				health = "fetchAppUpdate worked"
			} else {
				health = "fetchAppUpdate failed - " + a.AppHealthReport.ErrorMessage
			}
		}
		fmt.Fprintf(b, "- %s: %s\n", a.AppName, health)
		for _, part := range parts {
			fmt.Fprintf(b, "  - %s\n", part)
		}
	}
}
