package main

import (
	tr "github.com/ocelot-cloud/task-runner"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var deps Deps

func main() {
	tr.HandleSignals()

	rootCmd.AddCommand(healthCheckCmd, updateCmd)
	rootCmd.CompletionOptions = cobra.CompletionOptions{DisableDefaultCmd: true}

	var err error
	deps, err = Initialize()
	if err != nil {
		log.Fatalf("failed to initialize updater: %v", err)
	}

	if err := rootCmd.Execute(); err != nil {
		tr.ColoredPrintln("error: %v", err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "updater",
	Short: "CI runner for apps",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "updates all apps and performs health checks",
	Run: func(cmd *cobra.Command, args []string) {
		tr.PrintTaskDescription("updating apps")
		updateReport, err := deps.Updater.PerformUpdate()
		if err != nil {
			tr.ColoredPrintln("error: %v", err)
			os.Exit(1)
		}
		reportUpdate(*updateReport)
	},
}

var healthCheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "performs health checks",
	Run: func(cmd *cobra.Command, args []string) {
		tr.PrintTaskDescription("performing healthchecks")
		healthReport, err := deps.HealthChecker.PerformHealthChecks()
		if err != nil {
			tr.ColoredPrintln("error: %v", err)
			os.Exit(1)
		}
		reportHealth(*healthReport)
	},
}
