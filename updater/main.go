package main

import (
	tr "github.com/ocelot-cloud/task-runner"
	"github.com/spf13/cobra"
	"os"
)

var (
	updaterDir = getCurrentDir()
	projectDir = updaterDir + "/.."
	appsDir    = projectDir + "/apps/production"
)

func getCurrentDir() string {
	currentDir, err := os.Getwd()
	if err != nil {
		tr.CleanupAndExitWithError()
	}
	return currentDir
}

func main() {
	tr.HandleSignals()
	tr.PrintTaskDescription("This is a test task for the CI runner")

	rootCmd.AddCommand(testUnitsCmd)
	rootCmd.CompletionOptions = cobra.CompletionOptions{DisableDefaultCmd: true}

	tr.Execute("docker compose up -d")

	if err := rootCmd.Execute(); err != nil {
		tr.ColoredPrintln("error: %v", err)
		tr.CleanupAndExitWithError()
	}
}

var rootCmd = &cobra.Command{
	Use:   "ci-runner",
	Short: "CI runner for liquid",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var testUnitsCmd = &cobra.Command{
	Use:   "test",
	Short: "execute updater unit tests",
	Run: func(cmd *cobra.Command, args []string) {
		tr.PrintTaskDescription("execute unit tests")
		tr.ExecuteInDir(updaterDir, "go test -count=1 ./...")
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "updating production apps",
	Run: func(cmd *cobra.Command, args []string) {
		tr.PrintTaskDescription("updating production apps")
		// TODO conduct update
		// TODO conduct setup and stability test
		// TODO printing report
	},
}
