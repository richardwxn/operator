// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mesh

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"istio.io/operator/pkg/util"
	"os"
	"path/filepath"

	"istio.io/operator/pkg/object"
	"istio.io/pkg/log"
)

// YAMLSuffix is the suffix of a YAML file.
const YAMLSuffix = "yaml"

type manifestDiffArgs struct {
	// compareDir indicates comparison between directory.
	compareDir bool
}

func addManifestDiffFlags(cmd *cobra.Command, diffArgs *manifestDiffArgs) {
	cmd.PersistentFlags().BoolVarP(&diffArgs.compareDir, "directory", "r",
		false, "compare directory")
}

func manifestDiffCmd(rootArgs *rootArgs, diffArgs *manifestDiffArgs) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare manifests and generate diff.",
		Long:  "The diff-manifest subcommand is used to compare manifest from two files or directories.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if diffArgs.compareDir {
				compareManifestsFromDirs(args[0], args[1])
			} else {
				compareManifestsFromFiles(rootArgs, args)
			}
		}}
	return cmd
}

//compareManifestsFromFiles compares two manifest files
func compareManifestsFromFiles(rootArgs *rootArgs, args []string) {
	if err := configLogs(rootArgs); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Could not configure logs: %s", err)
		os.Exit(1)
	}
	a, err := ioutil.ReadFile(args[0])
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	b, err := ioutil.ReadFile(args[1])
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	diff, err := object.ManifestDiff(string(a), string(b))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	if diff == "" {
		fmt.Println("Manifests are identical")
	} else {
		fmt.Printf("Difference of manifests are:\n%s", diff)
	}
}

func yamlFileFilter(path string) bool {
	return filepath.Ext(path) == YAMLSuffix
}

//compareManifestsFromDirs compares manifests from two directories
func compareManifestsFromDirs(dirName1 string, dirName2 string) {
	mf1, err := util.ReadFromDir(dirName1, yamlFileFilter)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	mf2, err := util.ReadFromDir(dirName2, yamlFileFilter)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	diff, err := object.ManifestDiff(mf1, mf2)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	if diff == "" {
		fmt.Println("Manifests are identical")
	} else {
		fmt.Printf("Difference of manifests are:\n%s", diff)
	}
}
