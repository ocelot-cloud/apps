package main

import (
	"bytes"
	tr "github.com/ocelot-cloud/task-runner"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	removeAllContainers()
	removeAllVolumes()

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

func removeAllContainers() {
	logger.Info("removing all containers, as they could interfere with the update process")
	cmd := exec.Command("docker", "ps", "-q")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		logger.Fatal("failed to list running containers: %v", err)
	}
	ids := strings.Fields(out.String())
	if len(ids) == 0 {
		logger.Info("no running containers found to remove")
		return
	}
	args := append([]string{"rm", "-f"}, ids...)
	err := exec.Command("docker", args...).Run()
	if err != nil {
		logger.Fatal("failed to remove containers: %v", err)
	} else {
		logger.Info("removed all containers")
	}
}

func removeAllVolumes() {
	cmd := exec.Command("docker", "volume", "ls", "-q")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		logger.Fatal("failed to list volumes: %v", err)
	}
	vols := strings.Fields(out.String())
	if len(vols) == 0 {
		logger.Info("no volumes found to remove")
		return
	}
	args := append([]string{"volume", "rm", "-f"}, vols...)
	err := exec.Command("docker", args...).Run()
	if err != nil {
		logger.Fatal("failed to remove volumes: %v", err)
	} else {
		logger.Info("removed all volumes")
	}
}
