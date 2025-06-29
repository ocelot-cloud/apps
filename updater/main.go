package main

import (
	tr "github.com/ocelot-cloud/task-runner"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var (
	deps Deps
)

func main() {
	tr.HandleSignals()

	rootCmd.AddCommand(healthCheckCmd, updateCmd)
	rootCmd.CompletionOptions = cobra.CompletionOptions{DisableDefaultCmd: true}
	rootCmd.PersistentFlags().StringVarP(&appsDir, "apps-directory", "d", ".", "Path to apps directory")
	cobra.OnInitialize(func() {
		abs, err := filepath.Abs(appsDir)
		if err != nil {
			logger.Fatal("failed to resolve apps directory: %v", err)
		}
		appsDir = abs
	})

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
	Use:   "Fetch",
	Short: "updates all apps and performs health checks",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("starting Fetch process")
		updateReport, err := deps.Updater.PerformUpdate()
		if err != nil {
			tr.ColoredPrintln("error: %v", err)
			os.Exit(1)
		}
		output := reportUpdate(*updateReport)
		print(output)
	},
}

var healthCheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "performs health checks",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("starting health check process")
		healthReport, err := deps.HealthChecker.PerformHealthChecks()
		if err != nil {
			tr.ColoredPrintln("error: %v", err)
			os.Exit(1)
		}
		output := reportHealth(*healthReport)
		print(output)
	},
}
