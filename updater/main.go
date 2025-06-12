package main

import (
	"fmt"
	"github.com/ocelot-cloud/shared/utils"
	"os"
	"path/filepath"
	"time"

	tr "github.com/ocelot-cloud/task-runner"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
	logger.Info("sample log")

	rootCmd.AddCommand(testUnitsCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(testIntegrationCmd)
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

var testIntegrationCmd = &cobra.Command{
	Use:   "test-integration",
	Short: "run updater integration test on sampleapp",
	RunE: func(cmd *cobra.Command, args []string) error {
		tr.PrintTaskDescription("running integration test")
		composePath := filepath.Join(testDir, "sampleapp", "docker-compose.yml")
		data, err := os.ReadFile(composePath)
		if err != nil {
			return err
		}
		var compose map[string]any
		if err := yaml.Unmarshal(data, &compose); err != nil {
			return err
		}
		services := compose["services"].(map[string]any)
		svc := services["sampleapp"].(map[string]any)
		img, _ := svc["image"].(string)
		before := imageTag(img)

		mgr := buildManager()
		res, err := mgr.Update([]string{"sampleapp"})
		if err != nil {
			return err
		}
		if len(res) == 0 || !res[0].Success {
			return fmt.Errorf("update failed")
		}

		data, err = os.ReadFile(composePath)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(data, &compose); err != nil {
			return err
		}
		services = compose["services"].(map[string]any)
		svc = services["sampleapp"].(map[string]any)
		img, _ = svc["image"].(string)
		after := imageTag(img)
		expected := bumpTag(before)
		if after != expected {
			return fmt.Errorf("expected tag %s, got %s", expected, after)
		}
		fmt.Printf("sampleapp updated: %s -> %s\n", before, after)
		return nil
	},
}
