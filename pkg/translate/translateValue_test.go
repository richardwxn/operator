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
		desc    string
		valueYAML string
		want    string
		wantErr string
	}{
		{
			desc: "global",
			want: `
hub: docker.io/istio
tag: 1.2.3
telemetry:
  components:
    namespace: istio-telemetry
  enabled:
    value: true
policy:
  components:
    namespace: istio-policy
  enabled:
    value: true
`,
			valueYAML: `
certmanager:
  enabled: false
galley:
  enabled: false
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
nodeagent:
  enabled: false
pilot:
  enabled: false
security:
  enabled: false

`,
		},
	}
	tr := ValueTranslators[version.MinorVersion{Major: 1, Minor: 2}]

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
			cpYaml, err := yaml.Marshal(got)
			if want := tt.want; !util.IsYAMLEqual(string(cpYaml), want) {
				t.Errorf("ValuesToProto(%s): got:\n%s\n\nwant:\n%s\nDiff:\n%s\n", tt.desc, string(cpYaml), want, util.YAMLDiff(string(cpYaml), want))
			}
		})
	}
}
