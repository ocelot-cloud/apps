//go:build wireinject
// +build wireinject

package main

import (
	"fmt"
	"github.com/google/wire"
	"github.com/ocelot-cloud/shared/utils"
	"log"
	"os"
	"path/filepath"

	tr "github.com/ocelot-cloud/task-runner"
	"github.com/spf13/cobra"
)

var logger = utils.ProvideLogger("info")

var (
	updaterDir = getCurrentDir()
	projectDir = updaterDir + "/.."
	appsDir    string
	testDir    = projectDir + "/apps/test"
)

func init() {
	defaultDir := filepath.Join(projectDir, "apps/production")
	rootCmd.PersistentFlags().StringVarP(&appsDir, "path-apps-dir", "p", defaultDir, "directory containing app definitions")
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		tr.CleanupAndExitWithError()
	}
	return dir
}

func main() {
	tr.HandleSignals()

	rootCmd.AddCommand(testUnitsCmd)
	rootCmd.CompletionOptions = cobra.CompletionOptions{DisableDefaultCmd: true}

	// TODO needs be connected with cobra commands: healthcheckCmd, updateCmd
	deps, err := Initialize()
	if err != nil {
		log.Fatalf("failed to initialize updater: %v", err)
	}
	fmt.Printf("deps: %v\n", deps.Updater)

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

	},
}

type Deps struct {
	Updater       *Updater
	HealthChecker HealthChecker
}

func Initialize() (Deps, error) {
	wire.Build(
		NewUpdater,
		NewFileSystemOperator,
		NewSingleAppUpdater,
		NewHealthChecker,
		NewDockerHubClient,
		NewEndpointChecker,
		NewFileSystemUpdateOperator,
		wire.Struct(new(Deps), "*"),
	)
	return Deps{}, nil
}

func NewUpdater(fs FileSystemOperator, appUpdater SingleAppUpdater, checker HealthChecker, client DockerHubClient) *Updater {
	return &Updater{
		appsDir:            "",
		fileSystemOperator: fs,
		appUpdater:         appUpdater,
		healthChecker:      checker,
		dockerHubClient:    client,
	}
}

func NewFileSystemOperator() FileSystemOperator {
	return &FileSystemOperatorImpl{}
}

func NewSingleAppUpdater(fs SingleAppUpdateFileSystemOperator, client DockerHubClient) SingleAppUpdater {
	return &SingleAppUpdaterImpl{
		fsOperator:      fs,
		dockerHubClient: client,
	}
}

func NewHealthChecker(fs FileSystemOperator, checker EndpointChecker) HealthChecker {
	return &HealthCheckerImpl{
		appsDir:            "",
		fileSystemOperator: fs,
		endpointChecker:    checker,
	}
}

func NewDockerHubClient() DockerHubClient {
	return &DockerHubClientImpl{}
}

func NewEndpointChecker() EndpointChecker {
	return &EndpointCheckerImpl{}
}

func NewFileSystemUpdateOperator() SingleAppUpdateFileSystemOperator {
	return &SingleAppUpdateFileSystemOperatorImpl{}
}
