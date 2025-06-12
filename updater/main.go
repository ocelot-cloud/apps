package main

import (
	"fmt"
	"github.com/ocelot-cloud/shared/utils"
	"log"
	"os"
	"path/filepath"
	"time"

	tr "github.com/ocelot-cloud/task-runner"
	"github.com/spf13/cobra"
)

var logger = utils.ProvideLogger("info")

var (
	updaterDir = getCurrentDir()
	projectDir = updaterDir + "/.."
	appsDir    = filepath.Join(projectDir, "apps/production")
)

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		tr.CleanupAndExitWithError()
	}
	return dir
}

func main() {
	tr.HandleSignals()
	logger.Info("sample log")

	rootCmd.AddCommand(testUnitsCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.CompletionOptions = cobra.CompletionOptions{DisableDefaultCmd: true}

	if err := rootCmd.Execute(); err != nil {
		tr.ColoredPrintln("error: %v", err)
		tr.CleanupAndExitWithError()
	}
}

var rootCmd = &cobra.Command{
	Use:   "updater",
	Short: "CI runner for apps",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var testUnitsCmd = &cobra.Command{
	Use:   "test",
	Short: "execute updater unit tests",
	Run: func(cmd *cobra.Command, args []string) {
		tr.PrintTaskDescription("execute unit tests")
		tr.ExecuteInDir(updaterDir, "go test -count=1 -coverprofile=coverage.out ./...")
		tr.ExecuteInDir(updaterDir, "go tool cover -func=coverage.out")
	},
}

func buildManager() AppManager {
	return AppManager{
		AppsDir: appsDir,
		Runner:  CmdRunner{},
		Waiter:  DefaultWaiter{Timeout: time.Minute},
	}
}

var healthCmd = &cobra.Command{
	Use:   "healthcheck [apps...]",
	Short: "run health checks for production apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		tr.PrintTaskDescription("running health checks")
		mgr := buildManager()
		healthy, err := mgr.HealthCheck(args)
		if err != nil {
			return err
		}
		fmt.Println("summary: all services healthy")
		for _, name := range healthy {
			fmt.Printf("- %s\n", name)
		}
		return nil
	},
}

// TODO update should be written to compose file when healthcheck is passed after update. When an update worked, simply write the new tag in the docker compose yaml. That is it. The developer will handle the persistence afterwards via manual "git commit".
var updateCmd = &cobra.Command{
	Use:   "update [apps...]",
	Short: "update docker images and run health checks",
	RunE: func(cmd *cobra.Command, args []string) error {
		tr.PrintTaskDescription("updating images")
		mgr := buildManager()
		results, err := mgr.Update(args)
		if err != nil {
			return err
		}
		for _, r := range results {
			if r.Success {
				log.Printf("%s: update successful", r.App)
			} else {
				log.Printf("%s: update failed: %v", r.App, r.Error)
			}
		}
		return nil
	},
}
