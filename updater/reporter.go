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

func reportUpdate(updateReport UpdateReport) string {
	// TODO
	return ""
}
