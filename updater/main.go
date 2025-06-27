package main

import (
	"fmt"
	tr "github.com/ocelot-cloud/task-runner"
	"github.com/spf13/cobra"
	"log"
)

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
