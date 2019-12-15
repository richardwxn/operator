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
	"encoding/json"
	"fmt"
	"path/filepath"

	"istio.io/operator/pkg/helm"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/spf13/cobra"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/kubectlcmd"
	"istio.io/operator/pkg/translate"
	"istio.io/operator/pkg/util"
	binversion "istio.io/operator/version"
)

const (
	defaultNamespace   = "istio-system"
	defaultProfileName = "default"
)

type manifestMigrateArgs struct {
	// namespace is the namespace to get the in cluster configMap
	namespace string
	// profile is the base profile used when translating values.yaml
	profile string
}

func addManifestMigrateFlags(cmd *cobra.Command, args *manifestMigrateArgs) {
	cmd.PersistentFlags().StringVarP(&args.namespace, "namespace", "n", defaultNamespace,
		" Default namespace for output IstioControlPlane CustomResource")
	cmd.PersistentFlags().StringVarP(&args.profile, "profile", "p", defaultProfileName,
		" Default base profile name")
}

func manifestMigrateCmd(rootArgs *rootArgs, mmArgs *manifestMigrateArgs) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate [<filepath>]",
		Short: "Migrates a file containing Helm values to IstioControlPlane format",
		Long:  "The migrate subcommand migrates a configuration from Helm values format to IstioControlPlane format.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("migrate accepts optional single filepath")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			l := NewLogger(rootArgs.logToStdErr, cmd.OutOrStdout(), cmd.ErrOrStderr())

			if len(args) == 0 {
				return migrateFromClusterConfig(rootArgs, mmArgs, l)
			}

			return migrateFromFiles(rootArgs, args, mmArgs, l)
		}}
}

func valueFileFilter(path string) bool {
	return filepath.Base(path) == "values.yaml" || filepath.Base(path) == "global.yaml"
}

// migrateFromFiles handles migration for local values.yaml files
func migrateFromFiles(rootArgs *rootArgs, args []string, mmArgs *manifestMigrateArgs, l *Logger) error {
	initLogsOrExit(rootArgs)
	value, err := util.ReadFilesWithFilter(args[0], valueFileFilter)
	if err != nil {
		return err
	}
	if value == "" {
		l.logAndPrint("no valid value.yaml file specified")
		return nil
	}
	return translateFunc([]byte(value), mmArgs.profile, l)
}

// translateFunc translates the input values and output the result
func translateFunc(values []byte, profile string, l *Logger) error {
	ts, err := translate.NewReverseTranslator(binversion.OperatorBinaryVersion.MinorVersion)
	if err != nil {
		return fmt.Errorf("error creating values.yaml translator: %s", err)
	}

	translatedYAML, _, err := ts.TranslateFromValueToSpec(values)
	if err != nil {
		return fmt.Errorf("error translating values.yaml: %s", err)
	}
	profileYAML, err := genICPSFromProfile(profile, "")
	if err != nil {
		return fmt.Errorf("error generating profile yaml: %s", err)
	}
	mergedYAML, err := util.OverlayYAML(profileYAML, translatedYAML)

	mergedICPS, err := unmarshalAndValidateICPS(mergedYAML, true, l)
	if err != nil {
		return err
	}

	isCP := &v1alpha2.IstioControlPlane{Spec: mergedICPS, Kind: "IstioControlPlane", ApiVersion: "install.istio.io/v1alpha2"}

	ms := jsonpb.Marshaler{}
	gotString, err := ms.MarshalToString(isCP)
	if err != nil {
		return fmt.Errorf("error marshaling translated IstioControlPlane: %s", err)
	}

	isCPYaml, err := yaml.JSONToYAML([]byte(gotString))
	if err != nil {
		return fmt.Errorf("error converting JSON: %s\n%s", gotString, err)
	}

	l.print(string(isCPYaml) + "\n")
	return nil
}

func genICPSFromProfile(profile string, ver string) (string, error) {
	overlayYAML := ""

	if ver != "" && !util.IsFilePath(profile) {
		pkgPath, err := fetchInstallPackage(helm.InstallURLFromVersion(ver))
		if err != nil {
			return "", err
		}
		if helm.IsDefaultProfile(profile) {
			profile = filepath.Join(pkgPath, helm.ProfilesFilePath, helm.DefaultProfileFilename)
		} else {
			profile = filepath.Join(pkgPath, helm.ProfilesFilePath, profile+YAMLSuffix)
		}
	}

	// This contains the IstioControlPlane CR.
	baseCRYAML, err := helm.ReadProfileYAML(profile)
	if err != nil {
		return "", fmt.Errorf("could not read the profile values for %s: %s", profile, err)
	}

	if !helm.IsDefaultProfile(profile) {
		// Profile definitions are relative to the default profile, so read that first.
		dfn, err := helm.DefaultFilenameForProfile(profile)
		if err != nil {
			return "", err
		}
		defaultYAML, err := helm.ReadProfileYAML(dfn)
		if err != nil {
			return "", fmt.Errorf("could not read the default profile values for %s: %s", dfn, err)
		}
		baseCRYAML, err = util.OverlayYAML(defaultYAML, baseCRYAML)
		if err != nil {
			return "", fmt.Errorf("could not overlay the profile over the default %s: %s", profile, err)
		}
	}

	_, baseYAML, err := unmarshalAndValidateICP(baseCRYAML, true)
	if err != nil {
		return "", err
	}

	// Merge base and overlay.
	mergedYAML, err := util.OverlayYAML(baseYAML, overlayYAML)
	if err != nil {
		return "", fmt.Errorf("could not overlay user config over base: %s", err)
	}

	return mergedYAML, nil
}

// migrateFromClusterConfig handles migration for in cluster config.
func migrateFromClusterConfig(rootArgs *rootArgs, mmArgs *manifestMigrateArgs, l *Logger) error {
	initLogsOrExit(rootArgs)

	l.logAndPrint("translating in cluster specs\n")

	c := kubectlcmd.New()
	opts := &kubectlcmd.Options{
		Namespace: mmArgs.namespace,
		ExtraArgs: []string{"jsonpath='{.data.values}'"},
	}
	output, stderr, err := c.GetConfigMap("istio-sidecar-injector", opts)
	if err != nil {
		return err
	}
	if stderr != "" {
		l.logAndPrint("error: ", stderr, "\n")
	}
	var value map[string]interface{}
	if len(output) > 1 {
		output = output[1 : len(output)-1]
	}
	err = json.Unmarshal([]byte(output), &value)
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON to untyped map %s", err)
	}
	res, err := yaml.Marshal(value)
	if err != nil {
		return fmt.Errorf("error marshaling untyped map to YAML: %s", err)
	}
	// no concept of profile when translating in-cluster configs
	return translateFunc(res, "", l)
}
