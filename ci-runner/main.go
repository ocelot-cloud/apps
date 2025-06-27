package main

import (
	tr "github.com/ocelot-cloud/task-runner"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var (
	projectDir = getParentDir()
	updaterDir = projectDir + "/updater"
)

func getParentDir() string {
	projectDirName := "apps"
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Cannot get current directory: %v", err)
	}
	for {
		if filepath.Base(dir) == projectDirName {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			log.Fatalf("Cannot find directory named '%s'", projectDirName)
		}
		dir = parent
	}
}

func main() {
	tr.HandleSignals()
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(testAllCmd)

	if err := rootCmd.Execute(); err != nil {
		tr.ColoredPrintln("\nError during execution: %s\n", err.Error())
		tr.CleanupAndExitWithError()
	}
}

var rootCmd = &cobra.Command{
	Use:   "ci-runner",
	Short: "Updater is a service that updates apps",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var testAllCmd = &cobra.Command{
	Use:   "test",
	Short: "Updater is a service that updates apps",
	Run: func(cmd *cobra.Command, args []string) {
		tr.ExecuteInDir(updaterDir, "go generate")
		tr.ExecuteInDir(updaterDir, "go test ./...")
	},
}
