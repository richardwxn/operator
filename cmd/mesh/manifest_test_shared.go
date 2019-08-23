package mesh

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	repoRootDir string
	testDataDir string
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	repoRootDir = filepath.Join(wd, "../..")

	if err := syncCharts(); err != nil {
		panic(err)
	}
}

func runCommand(command string) (string, error) {
	var out bytes.Buffer
	rootCmd := GetRootCmd(strings.Split(command, " "))
	rootCmd.SetOutput(&out)

	if err := rootCmd.Execute(); err != nil {
		return "", err
	}
	return out.String(), nil
}

func syncCharts() error {
	cmd := exec.Command(filepath.Join(repoRootDir, "scripts/run_update_charts.sh"))
	return cmd.Run()
}

func readFile(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	return string(b), err
}
