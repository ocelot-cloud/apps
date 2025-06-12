package main

import (
	"fmt"
	tr "github.com/ocelot-cloud/task-runner"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
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

	rootCmd.AddCommand(testUnitsCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.CompletionOptions = cobra.CompletionOptions{DisableDefaultCmd: true}

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
	Short: "run health checks for production apps",
	Run: func(cmd *cobra.Command, args []string) {
		tr.PrintTaskDescription("running health checks")
		runHealthChecks()
	},
}

func runHealthChecks() {
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		tr.ColoredPrintln("error reading apps dir: %v", err)
		tr.CleanupAndExitWithError()
	}

	var healthy []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		appName := entry.Name()
		appDir := filepath.Join(appsDir, appName)
		port := readAppPort(appDir)
		if port == "" {
			continue
		}
		composeFile := filepath.Join(appDir, "docker-compose.yml")
		composeWithPort := injectPort(composeFile, appName, port)

		composeCmd := fmt.Sprintf("docker compose -f %s up -d", filepath.Base(composeWithPort))
		tr.ExecuteInDir(appDir, composeCmd)
		tr.WaitUntilPortIsReady(port)
		tr.WaitForWebPageToBeReady("http://localhost:" + port)
		healthy = append(healthy, appName)
		downCmd := fmt.Sprintf("docker compose -f %s down -v", filepath.Base(composeWithPort))
		tr.ExecuteInDir(appDir, downCmd)
		os.Remove(composeWithPort)
	}

	fmt.Println("summary: all services healthy")
	for _, name := range healthy {
		fmt.Printf("- %s\n", name)
	}
}

func readAppPort(appDir string) string {
	data, err := os.ReadFile(filepath.Join(appDir, "app.yml"))
	if err != nil {
		return ""
	}
	var obj map[string]any
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return ""
	}
	if p, ok := obj["port"].(int); ok {
		return fmt.Sprintf("%d", p)
	}
	if p, ok := obj["port"].(string); ok {
		return p
	}
	return ""
}

func injectPort(composePath, service, port string) string {
	data, err := os.ReadFile(composePath)
	if err != nil {
		tr.ColoredPrintln("error reading compose file: %v", err)
		tr.CleanupAndExitWithError()
	}

	var compose map[string]any
	if err := yaml.Unmarshal(data, &compose); err != nil {
		tr.ColoredPrintln("error parsing compose file: %v", err)
		tr.CleanupAndExitWithError()
	}

	services, ok := compose["services"].(map[string]any)
	if !ok {
		tr.ColoredPrintln("compose file has no services")
		tr.CleanupAndExitWithError()
	}

	svc, ok := services[service].(map[string]any)
	if !ok {
		for _, v := range services {
			svc, ok = v.(map[string]any)
			if ok {
				break
			}
		}
	}

	if svc == nil {
		tr.ColoredPrintln("no service found to inject port")
		tr.CleanupAndExitWithError()
	}

	ports, ok := svc["ports"].([]any)
	if !ok {
		ports = []any{}
	}
	mapping := fmt.Sprintf("%s:%s", port, port)
	already := false
	for _, p := range ports {
		if ps, ok := p.(string); ok && strings.HasPrefix(ps, port+":") {
			already = true
		}
	}
	if !already {
		ports = append(ports, mapping)
		svc["ports"] = ports
	}
	services[service] = svc
	compose["services"] = services

	out, err := yaml.Marshal(compose)
	if err != nil {
		tr.ColoredPrintln("error marshaling compose: %v", err)
		tr.CleanupAndExitWithError()
	}

	temp := composePath + ".temp"
	if err := os.WriteFile(temp, out, 0644); err != nil {
		tr.ColoredPrintln("error writing temp compose: %v", err)
		tr.CleanupAndExitWithError()
	}
	return temp
}
