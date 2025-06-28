package main

import (
	"fmt"
	tr "github.com/ocelot-cloud/task-runner"
	"github.com/spf13/cobra"
	"log"
)

var deps Deps

func main() {
	tr.HandleSignals()

	rootCmd.AddCommand(healthCheckCmd, updateCmd)
	rootCmd.CompletionOptions = cobra.CompletionOptions{DisableDefaultCmd: true}

	// TODO needs be connected with cobra commands: healthcheckCmd, updateCmd
	var err error
	deps, err = Initialize()
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

// TODO !! crate report functionality; maybe report -> string which can be asserted

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "updates all apps and performs health checks",
	Run: func(cmd *cobra.Command, args []string) {
		tr.PrintTaskDescription("updating apps")
		// TODO
	},
}

var healthCheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "performs health checks",
	Run: func(cmd *cobra.Command, args []string) {
		tr.PrintTaskDescription("performing healthchecks")
		// TODO
	},
}
