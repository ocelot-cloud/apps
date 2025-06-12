package main

import (
	"github.com/ocelot-cloud/shared/utils"
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
		tr.ExecuteInDir(updaterDir, "go generate ./...")
		tr.ExecuteInDir(updaterDir, "go test -count=1 ./...")
	},
}
