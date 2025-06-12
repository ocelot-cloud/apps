package main

import (
	"fmt"
	"github.com/ocelot-cloud/shared/utils"
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
		results, err := mgr.HealthCheck(args)
		if err != nil {
			return err
		}
		fmt.Println("summary: all services healthy")
		for _, r := range results {
			if r.Success {
				fmt.Printf("- %s\n", r.App)
			}
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
				fmt.Printf("%s:\n", r.App)
				for _, s := range r.Services {
					fmt.Printf("  %s: %s -> %s\n", s.Service, s.Before, s.After)
				}
			} else {
				fmt.Printf("%s: update failed\n", r.App)
			}
		}
		return nil
	},
}

// TODO delete production/sampleapp; only use test/sampleapp; you may need to adda flag to cobra to specify the app folder to be used. By default/without that flag it is folder "production" used. but for local testing we need to implement sth like "-f ../apps/test" or so

/* TODO add a command "test-integration":
* get current sampleapp image tag in "sampleapp/docker-compose.yml"
* it runs "./updater update sampleapp" (maybe command on sampleapp), must be successful
* check that the updated/newer sampleapp image tag was written to "sampleapp/docker-compose.yml"
 */

// TODO delete "mocks" folder, install mockery v3.3.5, and "go generate ./..." the mocks dynamically before running the tests; we dont want to hard code mocks
