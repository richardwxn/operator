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

package iop

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

func manifestDiffCmd(rootArgs *rootArgs) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mandiff",
		Short: "Compare two manifest.",
		Long:  "The mandiff subcommand is used to compare manifest from files.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			//compareManifests(rootArgs)
			fmt.Println("Print: " + strings.Join(args, " "))
		}}
	//cmd.Flags().StringSlice
	return cmd
}

//func compareManifests(args *rootArgs) {
//	if err := configLogs(args); err != nil {
//		_, _ = fmt.Fprintf(os.Stderr, "Could not configure logs: %s", err)
//		os.Exit(1)
//	}
//
//	a, err := ioutil.ReadFile(filepath.Join(testDataDir, path))
//
//
//	b, err := ioutil.ReadFile(filepath.Join(testDataDir, path))
//
//	diff, err := util.ManifestDiff(string(a), string(b))
//
//}