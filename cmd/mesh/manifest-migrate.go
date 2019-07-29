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
	"io/ioutil"
	"istio.io/operator/pkg/kubectlcmd"
	"istio.io/operator/pkg/util"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/translate"
	"istio.io/operator/pkg/version"
)

type manifestMigrateArgs struct {
	migrateDir bool
}

func addManifestMigrateFlags(cmd *cobra.Command, args *manifestMigrateArgs) {
	cmd.PersistentFlags().BoolVarP(&args.migrateDir, "migrateDir", "r", false, "translate directory")
}

func manifestMigrateCmd(rootArgs *rootArgs, mmArgs *manifestMigrateArgs) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrates a file containing Helm values to IstioControlPlane format.",
		Long:  "The migrate subcommand is used to migrate a configuration in Helm values format to IstioControlPlane format.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				migrateFromClusterConfig(rootArgs)
			} else {
				migrateFromFiles(rootArgs, args, mmArgs)
			}
		}}
}

func valueFileFilter(path string) bool {
	return filepath.Base(path) == "values.yaml"
}

// migrateFromFiles handles migration for local values.yaml files
func migrateFromFiles(rootArgs *rootArgs, args [] string, mmArgs *manifestMigrateArgs) {
	if err := configLogs(rootArgs); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Could not configure logs: %s", err)
		os.Exit(1)
	}
	logAndPrintf(rootArgs, "translating input values.yaml file at: %s to new API", args[0])
	if mmArgs.migrateDir {
		value, err := util.ReadFromDir(args[0], valueFileFilter)
		if err != nil {
			logAndFatalf(rootArgs, err.Error())
		}
		translateFunc(rootArgs, []byte(value))
	} else {
		value, err := ioutil.ReadFile(args[0])
		if err != nil {
			logAndFatalf(rootArgs, err.Error())
		}
		translateFunc(rootArgs, value)
	}
}

// translateFunc translates the input values and output the result
func translateFunc(rootArgs *rootArgs, values []byte) {
	ts, err := translate.NewValueYAMLTranslator(version.NewMinorVersion(1, 3))
	if err != nil {
		logAndFatalf(rootArgs, "failed to create values.yaml translator: %s", err.Error())
	}

	valueStruct := v1alpha2.Values{}
	err = yaml.Unmarshal(values, &valueStruct)
	if err != nil {
		logAndFatalf(rootArgs, "failed to unmarshall values.yaml into value struct : %s", err.Error())
	}

	isCPSpec, err := ts.TranslateFromValueToSpec(&valueStruct)
	if err != nil {
		logAndFatalf(rootArgs, "failed to translate values.yaml: %s", err.Error())
	}
	isCPYaml, err := yaml.Marshal(isCPSpec)
	if err != nil {
		logAndFatalf(rootArgs, "failed to marshal translated IstioControlPlaneSpec: %s", err.Error())
	}

	fmt.Println(string(isCPYaml))
}

//TODO use default in cluster if file not specified
// migrateFromClusterConfig handles migration for incluster config.
func migrateFromClusterConfig(rootArgs *rootArgs) {
	if err := configLogs(rootArgs); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Could not configure logs: %s", err)
		os.Exit(1)
	}
	logAndPrintf(rootArgs, "translating input incluster specs")

	c := kubectlcmd.New()
	output, stderr, err := c.GetConfig("istio-sidecar-injector", "istio-control", "jsonpath='{.data.values}'")
	if err != nil {

	}

	value, err := yaml.JSONToYAML([]byte(output))
	if err != nil {
	}

	translateFunc(rootArgs, value)
}
