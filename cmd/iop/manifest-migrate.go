package iop

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/translate"
	"istio.io/operator/pkg/version"
)

func manifestMigrateCmd(rootArgs *rootArgs) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrates a file containing Helm values to IstioControlPlane format.",
		Long:  "The migrate subcommand is used to migrate a configuration in Helm values format to IstioControlPlane format.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			manifestMigrate(rootArgs, args)
		}}

}

func manifestMigrate(rootArgs *rootArgs, args []string) {

	ts, err := translate.NewValueYAMLTranslator(version.NewMinorVersion(1, 2))

	// if file exist
	fmt.Println(args[0])
	f, err := ioutil.ReadFile(args[0])
	if err != nil {
		logAndFatalf(rootArgs, "fail to open values.yaml file: %s", err.Error())
	}

	// else use default in cluster
	valueStruct := v1alpha2.Values{}
	err = yaml.Unmarshal(f, &valueStruct)
	if err != nil {
		logAndFatalf(rootArgs, "fail to unmarshall values.yaml into struct : %s", err.Error())
	}

	isCP, err := ts.TranslateFromValueToSpec(&valueStruct)

	isCP.

}
