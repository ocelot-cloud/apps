package main

import (
	"fmt"
	"strings"
)

func reportHealth(h HealthCheckReport) string {
	var b strings.Builder
	for _, a := range h.AppHealthReports {
		if a.Healthy {
			fmt.Fprintf(&b, "%s: OK\n", a.AppName)
		} else {
			fmt.Fprintf(&b, "%s: %s\n", a.AppName, a.ErrorMessage)
		}
	}
	if h.AllAppsHealthy {
		b.WriteString("Summary: All apps are healthy")
	} else {
		b.WriteString("Summary: Some apps are unhealthy")
	}
	return b.String()
}

func reportUpdate(u UpdateReport) string {
	var b strings.Builder
	for _, a := range u.AppUpdateReport {
		switch {
		case !a.WasSuccessful:
			if a.AppUpdates != nil && a.AppUpdates.ErrorMessage != "" {
				fmt.Fprintf(&b, "%s: service update error - %s\n", a.AppName, a.AppUpdates.ErrorMessage)
			} else {
				fmt.Fprintf(&b, "%s: update failed - %s\n", a.AppName, a.UpdateErrorMessage)
			}
		case !a.WasUpdateAvailable:
			fmt.Fprintf(&b, "%s: no update available\n", a.AppName)
		default:
			if a.AppUpdates != nil && a.AppUpdates.ErrorMessage != "" {
				fmt.Fprintf(&b, "%s: service update error - %s\n", a.AppName, a.AppUpdates.ErrorMessage)
				break
			}
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
			fmt.Fprintf(&b, "%s: %s%s\n", a.AppName, strings.Join(parts, "; "), health)
		}
	}
	if u.WasSuccessful {
		b.WriteString("Summary: Update successful")
	} else {
		b.WriteString("Summary: Update failed")
	}
	return b.String()
}
