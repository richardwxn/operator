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
pilot:
  enabled: true
nodeAgent:
  enabled: true
gateways:
  enabled: true
  istio-ingressgateway:
    enabled: true
sidecarInjectorWebhook:
  enabled: false
`,
			want: `
hub: docker.io/istio
tag: 1.2.3
default_namespace_prefix: istio-system
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
    citadel:
      common:
        enabled: {}
  enabled:
    value: true
traffic_management:
   components:
     pilot:
       common:
         enabled: 
           value: true
   enabled: 
     value: true
auto_injection:
  components:
    injector:
      common:
        enabled: {}
  enabled: {}
gateways:
  components:
    ingress_gateway:
      - gateway:
          common:
           enabled:
             value: true
    egress_gateway:
      - gateway:
          common:
           enabled: {} 
  enabled:
    value: true
`,
		},
		{
			desc: "Some components Disabled",
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
			want: `
hub: docker.io/istio
tag: 1.2.3
default_namespace_prefix: istio-system
telemetry:
 components:
   namespace: istio-telemetry
   telemetry:
     common:
       enabled: {}
 enabled: {}
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
       enabled: {}
 enabled: {}
security:
 components:
   namespace: istio-system
   cert_manager:
     common:
       enabled: {}
   node_agent:
     common:
       enabled: {}
   citadel:
     common:
       enabled: {}
 enabled: {}
gateways:
 components:
   ingress_gateway:
     - gateway:
         common:
          enabled: {}
   egress_gateway:
     - gateway:
         common:
          enabled: {}
 enabled: {}
traffic_management:
 components:
   pilot:
     common:
       enabled:
         value: true
 enabled:
   value: true
auto_injection:
 components:
   injector:
      common:
       enabled: {}
 enabled: {}
`,
		},
		{
			desc: "K8s resources translation",
			valueYAML: `
galley:
  enabled: false
pilot:
  enabled: true
  resources:
    requests:
      cpu: 1000m
      memory: 1G
  replicaCount: 1
  nodeSelector:
    beta.kubernetes.io/os: linux
  autoscaleEnabled: true
  autoscaleMax: 3
  autoscaleMin: 1
  env:
    GODEBUG: gctrace=1
  podAntiAffinityLabelSelector:
    - labelSelector:
        matchLabels:
          testK1: testV1
global:
  hub: docker.io/istio
  istioNamespace: istio-system
  policyNamespace: istio-policy
  tag: 1.2.3
  telemetryNamespace: istio-telemetry
  proxy: 
    ReadinessInitialDelaySeconds: 2
mixer:
  policy:
    enabled: true
  telemetry:
    enabled: false
`,
			want: `
hub: docker.io/istio
tag: 1.2.3
default_namespace_prefix: istio-system
telemetry:
  components:
    namespace: istio-telemetry
    telemetry:
      common:
        enabled: {}
  enabled: {}
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
        enabled: {}
  enabled: {}
security:
  components:
    namespace: istio-system
    cert_manager:
      common:
        enabled: {}
    node_agent:
      common:
        enabled: {}
    citadel:
      common:
        enabled: {}
  enabled: {}
gateways:
  components:
    ingress_gateway:
      - gateway:
          common:
           enabled: {}
    egress_gateway:
      - gateway:
          common:
           enabled: {}
  enabled: {}
traffic_management:
  components:
    pilot:
      common:
        enabled:
          value: true
        k8s:
          affinity:
            podAntiAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
              - labelSelector:
                    matchLabels:
                      testK1: testV1
                topologyKey: ""
          replica_count: 1
          env:
          - name: GODEBUG
            value: gctrace=1
          hpa_spec:
             maxReplicas: 3
             minReplicas: 1
             scaleTargetRef:
               kind: ""
               name: ""
          node_selector:
             beta.kubernetes.io/os: linux
          resources:
             requests:
               cpu: 1000m
               memory: 1G
    proxy:
      common:
        k8s:
          readiness_probe:
            initial_delaySeconds: 2
        values:
          readinessInitialDelaySeconds: 2
  enabled:
    value: true
auto_injection:
  components:
    injector:
       common:
        enabled: {}
  enabled: {}
`,
		}}
	tr, err := NewValueYAMLTranslator(version.NewMinorVersion(1, 2))
	if err != nil {
		t.Fatal("fail to get helm value.yaml translator")
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			valueStruct := v1alpha2.Values{}
			err := yaml.Unmarshal([]byte(tt.valueYAML), &valueStruct)
			if err != nil {
				t.Fatalf("unmarshal(%s): got error %s", tt.desc, err)
			}
			scope.Debugf("value struct: \n%s\n", pretty.Sprint(valueStruct))
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
