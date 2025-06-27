package main

import (
	"github.com/ocelot-cloud/shared/utils"
	tr "github.com/ocelot-cloud/task-runner"
	"os"
	"path/filepath"
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
