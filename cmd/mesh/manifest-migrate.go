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
	"os"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/translate"
	"istio.io/operator/pkg/version"
)

type manifestMigrateArgs struct {
	// inFilename is the path to the input values.yaml file.
	inFilename string
	// outFilename is the path to the translated output filename.
	outFilename string
}

func addManifestMigrateFlags(cmd *cobra.Command, args *manifestMigrateArgs) {
	cmd.PersistentFlags().StringVarP(&args.inFilename, "filename", "f", "", filenameFlagHelpStr)
	cmd.PersistentFlags().StringVarP(&args.outFilename, "output", "o", "", "Translation output file path.")
}

func manifestMigrateCmd(rootArgs *rootArgs, mmArgs *manifestMigrateArgs) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrates a file containing Helm values to IstioControlPlane format.",
		Long:  "The migrate subcommand is used to migrate a configuration in Helm values format to IstioControlPlane format.",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			manifestMigrate(rootArgs, mmArgs)
		}}
}

func manifestMigrate(rootArgs *rootArgs, mmArgs *manifestMigrateArgs) {
	if err := configLogs(rootArgs); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Could not configure logs: %s", err)
		os.Exit(1)
	}
	f, err := ioutil.ReadFile(mmArgs.inFilename)
	if err != nil {
		logAndFatalf(rootArgs, "failed to open values.yaml file: %s", err.Error())
	}

	ts, err := translate.NewValueYAMLTranslator(version.NewMinorVersion(1, 3))
	if err != nil {
		logAndFatalf(rootArgs, "failed to create values.yaml translator: %s", err.Error())
	}
	logAndPrintf(rootArgs, "translating input values.yaml file at: %s to new API", mmArgs.inFilename)

	valueStruct := v1alpha2.Values{}
	err = yaml.Unmarshal(f, &valueStruct)
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

	if mmArgs.outFilename == "" {
		fmt.Println(string(isCPYaml))
	} else {
		if err := ioutil.WriteFile(mmArgs.outFilename, isCPYaml, os.ModePerm); err != nil {
			logAndFatalf(rootArgs, err.Error())
		}
	}
}

//TODO use default in cluster if file not specified
