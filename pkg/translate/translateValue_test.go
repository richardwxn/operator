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

package translate

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/kr/pretty"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/util"
	"istio.io/operator/pkg/version"
)

func TestValueToProto(t *testing.T) {
	tests := []struct {
		desc      string
		valueYAML string
		want      string
		wantErr   string
	}{
		{
			desc: "All Enabled",
			want: `
hub: docker.io/istio
tag: 1.2.3
telemetry:
  components:
    namespace: istio-telemetry
    telemetry:
      common:
        enabled:
          value: true
  enabled:
    value: true
policy:
  components:
    namespace: istio-policy
    policy:
      common:
        enabled:
          value: true
  enabled:
    value: true
config_management:
  components:
    galley:
      common:
        enabled:
          value: true
  enabled:
    value: true 
security:
  components:
    namespace: istio-system
    cert_manager:
      common:
        enabled:
          value: true
    node_agent:
      common:
        enabled:
          value: true
  enabled:
    value: true
`,
			valueYAML: `
certManager:
  enabled: true
galley:
  enabled: true
global:
  hub: docker.io/istio
  istioNamespace: istio-system
  policyNamespace: istio-policy
  tag: 1.2.3
  telemetryNamespace: istio-telemetry
mixer:
  policy:
    enabled: true
  telemetry:
    enabled: true
nodeAgent:
  enabled: true
`,
		},
		{
			desc: "Some components Disabled",
			want: `
hub: docker.io/istio
tag: 1.2.3
telemetry:
  components:
    namespace: istio-telemetry
    telemetry:
      common:
        enabled:  {}
  enabled:  {}
policy:
  components:
    namespace: istio-policy
    policy:
      common:
        enabled:
          value: true
  enabled:
    value: true
config_management:
  components:
    galley:
      common:
        enabled:  {}
  enabled:  {} 
security:
  components:
    namespace: istio-system
`,
			valueYAML: `
galley:
  enabled: false
pilot:
  enabled: true
global:
  hub: docker.io/istio
  istioNamespace: istio-system
  policyNamespace: istio-policy
  tag: 1.2.3
  telemetryNamespace: istio-telemetry
mixer:
  policy:
    enabled: true
  telemetry:
    enabled: false
`,
		},
	}
	tr := ValueTranslators[version.NewMinorVersion(1, 2)]

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			valueStruct := v1alpha2.Values{}
			err := yaml.Unmarshal([]byte(tt.valueYAML), &valueStruct)
			if err != nil {
				t.Fatalf("unmarshal(%s): got error %s", tt.desc, err)
			}
			dbgPrint("ispec: \n%s\n", pretty.Sprint(valueStruct))
			got, err := tr.TranslateFromValueToSpec(&valueStruct)
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr {
				t.Errorf("ValuesToProto(%s)(%v): gotErr:%s, wantErr:%s", tt.desc, tt.valueYAML, gotErr, wantErr)
			}
			cpYaml, _ := yaml.Marshal(got)
			if want := tt.want; !util.IsYAMLEqual(string(cpYaml), want) {
				t.Errorf("ValuesToProto(%s): got:\n%s\n\nwant:\n%s\nDiff:\n%s\n", tt.desc, string(cpYaml), want, util.YAMLDiff(string(cpYaml), want))
			}
		})
	}
}
