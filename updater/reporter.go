package main

import (
	"fmt"
	"strings"
)

func reportHealth(h HealthCheckReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\nHealth Check Report:\n\n")
	for _, a := range h.AppHealthReports {
		if a.Healthy {
			fmt.Fprintf(&b, "- %s: OK\n", a.AppName)
		} else {
			fmt.Fprintf(&b, "- %s: %s\n", a.AppName, a.ErrorMessage)
		}
	}
	if h.AllAppsHealthy {
		b.WriteString("\nSummary: All apps are healthy\n")
	} else {
		b.WriteString("\nSummary: Some apps are unhealthy\n")
	}
	return b.String()
}

// TODO make pretty
func reportUpdate(updateReport UpdateReport) string {
	var builder strings.Builder
	for _, appUpdateReport := range updateReport.AppUpdateReport {
		addUpdateReportLine(appUpdateReport, &builder)
	}
	if updateReport.WasSuccessful {
		fmt.Fprintf(&builder, "Summary: Update successful\n")
	} else {
		fmt.Fprintf(&builder, "Summary: Update failed\n")
	}
	return builder.String()
}

func addUpdateReportLine(a AppUpdateReport, b *strings.Builder) {
	switch {
	case !a.WasSuccessful:
		if a.AppUpdates != nil && len(a.AppUpdates.ServiceUpdates) > 0 && a.AppUpdates.ErrorMessage != "" {
			fmt.Fprintf(b, "%s: service update error - %s\n", a.AppName, a.AppUpdates.ErrorMessage)
		} else {
			fmt.Fprintf(b, "%s: update failed - %s\n", a.AppName, a.UpdateErrorMessage)
		}
	case !a.WasUpdateAvailable:
		fmt.Fprintf(b, "%s: no update available\n", a.AppName)
	default:
		parts := make([]string, 0, len(a.AppUpdates.ServiceUpdates))
		for _, s := range a.AppUpdates.ServiceUpdates {
			parts = append(parts, fmt.Sprintf("%s %s->%s", s.ServiceName, s.OldTag, s.NewTag))
		}
		health := ""
		if a.AppHealthReport != nil {
			if a.AppHealthReport.Healthy {
				health = ", Health: OK"
			} else {
				health = ", Health: " + a.AppHealthReport.ErrorMessage
			}
		}
		fmt.Fprintf(b, "%s: %s%s\n", a.AppName, strings.Join(parts, "; "), health)
	}
}
