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
	"fmt"
	"strings"

	"github.com/ghodss/yaml"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/name"
	"istio.io/operator/pkg/tpath"
	"istio.io/operator/pkg/util"
	"istio.io/operator/pkg/version"
	"istio.io/pkg/log"
)

// ValueYAMLTranslator is a set of mappings to translate between values.yaml and API paths, charts, k8s paths.
type ValueYAMLTranslator struct {
	Version version.MinorVersion
	// APIMapping is Values.yaml path to API path mapping using longest prefix match. If the path is a non-leaf node,
	// the output path is the matching portion of the path, plus any remaining output path.
	APIMapping map[string]*Translation
	// KubernetesMapping defines mappings from an  k8s resource paths to IstioControlPlane API paths.
	KubernetesMapping map[string]*Translation
	// ValuesToFeatureComponentName defines mapping from value path to feature and component name in API paths.
	ValuesToComponentName map[string]name.ComponentName
	// NamespaceMapping maps namespace defined in value.yaml to that in API spec.
	NamespaceMapping map[string][]string
}

var (
	// ValueYAMLTranslators is Translator for value.yaml
	ValueYAMLTranslators = map[version.MinorVersion]*ValueYAMLTranslator{
		version.NewMinorVersion(1, 2): {
			APIMapping: map[string]*Translation{},
			KubernetesMapping: map[string]*Translation{
				// TODO use template for podaffinity
				"{{.ValueComponentName}}.podAntiAffinityLabelSelector": {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8s.Affinity", nil},
				"{{.ValueComponentName}}.env":                          {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8s.Env", nil},
				"{{.ValueComponentName}}.autoscaleEnabled":             {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8s.HpaSpec", nil},
				"{{.ValueComponentName}}.imagePullPolicy":              {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8s.ImagePullPolicy", nil},
				"{{.ValueComponentName}}.nodeSelector":                 {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8s.NodeSelector", nil},
				"{{.ValueComponentName}}.podDisruptionBudget":          {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8s.PodDisruptionBudget", nil},
				"{{.ValueComponentName}}.podAnnotations":               {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8s.PodAnnotations", nil},
				"{{.ValueComponentName}}.priorityClassName":            {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8s.PriorityClassName", nil},
				// TODO check readinessProbe mapping
				"{{.ValueComponentName}}.readinessProbe": {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8s.ReadinessProbe", nil},
				"{{.ValueComponentName}}.replicaCount":   {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8s.ReplicaCount", nil},
				"{{.ValueComponentName}}.resources":      {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8s.Resources", nil},
			},
			ValuesToComponentName: map[string]name.ComponentName{
				"pilot":                         name.PilotComponentName,
				"galley":                        name.GalleyComponentName,
				"sidecarInjectorWebhook":        name.SidecarInjectorComponentName,
				"mixer.policy":                  name.PolicyComponentName,
				"mixer.telemetry":               name.TelemetryComponentName,
				"security":                      name.CitadelComponentName,
				"nodeagent":                     name.NodeAgentComponentName,
				"certmanager":                   name.CertManagerComponentName,
				"gateways.istio-ingressgateway": name.IngressComponentName,
				"gateways.istio-egressgateway":  name.EgressComponentName,
			},
			NamespaceMapping: map[string][]string{
				"global.istioNamespace":     {"security.components.namespace"},
				"global.telemetryNamespace": {"telemetry.components.namespace"},
				"global.policyNamespace":    {"policy.components.namespace"},
				"global.configNamespace":    {"configManagement.components.namespace"},
			},
		},
	}
	// Component enablement mapping. Ex "{{.ValueComponent}}.enabled": {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.enabled}", nil},
	// Component enablement mapping for gateway Ex "{{.ValueComponent}}.enabled": {"{{.FeatureName}}.Components.{{.ComponentName}}.Gateway.Common.enabled}", nil},
	// Feature enablement mapping. Ex: "{{.ValueComponent}}.enabled": {"{{.FeatureName}}.enabled}", nil},
	componentEnablementPattern   = "{{.FeatureName}}.Components.{{.ComponentName}}.Common.Enabled"
	gWComponentEnablementPattern = "{{.FeatureName}}.Components.{{.ComponentName}}.Gateway.Common.Enabled"
)

// initAPIMapping generate the reverse mapping from original translator apiMapping
func (t *ValueYAMLTranslator) initAPIMapping(vs version.MinorVersion) {
	for valKey, outVal := range Translators[vs].APIMapping {
		t.APIMapping[outVal.outPath] = &Translation{valKey, nil}
	}
}

// initK8SMapping generates the k8s settings mapping for all components based on templates
func (t *ValueYAMLTranslator) initK8SMapping() {
	outputMapping := make(map[string]*Translation)
	for valKey, componentName := range t.ValuesToComponentName {
		featureName := name.ComponentNameToFeatureName[componentName]
		if featureName == "" {
			continue
		}
		for K8SValKey, outPathTmpl := range t.KubernetesMapping {
			newKey := renderComponentName(K8SValKey, valKey)
			newVal := renderFeatureComponentPathTemplate(outPathTmpl.outPath, featureName, componentName)
			outputMapping[newKey] = &Translation{newVal, nil}
		}
	}
	t.KubernetesMapping = outputMapping
}

// NewValueYAMLTranslator creates a new ValueYAMLTranslator for minorVersion and returns a ptr to it.
func NewValueYAMLTranslator(minorVersion version.MinorVersion) (*ValueYAMLTranslator, error) {
	t := ValueYAMLTranslators[minorVersion]
	if t == nil {
		return nil, fmt.Errorf("no translator available for version %s", minorVersion)
	}

	t.initAPIMapping(minorVersion)
	t.initK8SMapping()
	return t, nil
}

// TranslateFromValueToSpec translates from values struct to IstioControlPlaneSpec
func (t *ValueYAMLTranslator) TranslateFromValueToSpec(values *v1alpha2.Values) (controlPlaneSpec *v1alpha2.IstioControlPlaneSpec, err error) {
	// marshal value struct to yaml
	valueYaml, err := yaml.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("error when marshalling value struct %v", err.Error())
	}
	// unmarshal yaml to untyped tree
	var yamlTree = make(map[string]interface{})
	err = yaml.Unmarshal(valueYaml, &yamlTree)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshalling into untype tree %v", err.Error())
	}

	outputTree := make(map[string]interface{})
	err = t.TranslateTree(yamlTree, outputTree, nil)
	if err != nil {
		return nil, err
	}
	outputVal, err := yaml.Marshal(outputTree)
	if err != nil {
		return nil, err
	}

	var cpSpec = &v1alpha2.IstioControlPlaneSpec{}
	err = util.UnmarshalWithJSONPB(string(outputVal), cpSpec)

	if err != nil {
		return nil, fmt.Errorf("error when unmarshalling into control plane spec %v", err.Error())
	}

	return cpSpec, nil
}

// TranslateTree translates input value.yaml Tree to ControlPlaneSpec Tree
func (t *ValueYAMLTranslator) TranslateTree(valueTree map[string]interface{}, cpSpecTree map[string]interface{}, path util.Path) error {
	// translate enablement and namespace
	err := t.setEnablementAndNamespacesFromValue(valueTree, cpSpecTree)
	if err != nil {
		return fmt.Errorf("error when translating enablement and namespace from value.yaml tree %v", err.Error())
	}
	// translate with api mapping
	err = translateTree(valueTree, cpSpecTree, t.APIMapping)
	if err != nil {
		return fmt.Errorf("error when translating value.yaml tree with global mapping %v", err.Error())
	}
	// translate with k8s mapping
	err = translateTree(valueTree, cpSpecTree, t.KubernetesMapping)
	if err != nil {
		return fmt.Errorf("error when translating value.yaml tree with kubernetes mapping %v", err.Error())
	}
	return nil
}

// setEnablementAndNamespaces translates the enablement and namespace value of each component in the baseYAML values
// tree, based on feature/component inheritance relationship.
func (t *ValueYAMLTranslator) setEnablementAndNamespacesFromValue(valueSpec map[string]interface{}, root map[string]interface{}) error {
	for cnv, cni := range t.ValuesToComponentName {
		enabled, err := name.IsComponentEnabledFromValue(cnv, valueSpec)
		if err != nil {
			return err
		}
		featureName := name.ComponentNameToFeatureName[cni]
		tmpl := componentEnablementPattern
		if strings.Contains(cnv, "gateway") {
			tmpl = gWComponentEnablementPattern
		}
		ceVal := renderFeatureComponentPathTemplate(tmpl, featureName, cni)
		outCP := util.ToYAMLPath(ceVal)
		// set component enablement
		if err := tpath.WriteNode(root, outCP, enabled); err != nil {
			return err
		}
		// set feature enablement
		feVal := featureName + ".Enabled"
		outFP := util.ToYAMLPath(string(feVal))
		curEnabled, found, _ := name.GetFromTreePath(root, outFP)
		if !found {
			if err := tpath.WriteNode(root, outFP, enabled); err != nil {
				return err
			}
		} else if curEnabled == false && enabled {
			if err := tpath.WriteNode(root, outFP, enabled); err != nil {
				return err
			}
		}
	}
	// TODO: base component is always enabled

	// set namespace
	for vp, nsList := range t.NamespaceMapping {
		namespace, err := name.NamespaceFromValue(vp, valueSpec)
		if err != nil {
			return err
		}
		for _, ns := range nsList {
			if err := tpath.WriteNode(root, util.ToYAMLPath(ns), namespace); err != nil {
				return err
			}
		}
	}
	return nil
}

// translateTree is internal method for translating value.yaml tree
func translateTree(valueTree map[string]interface{},
	cpSpecTree map[string]interface{}, mapping map[string]*Translation) error {
	for inPath, v := range mapping {
		// Extra logics needed for K8s translation
		// HPA spec


		// Readiness Probe



		// Env, which is map[string]string originally then to []*
		log.Infof("Checking for path %s in helm Value.yaml tree", inPath)
		m, found, err := name.GetFromTreePath(valueTree, util.ToYAMLPath(inPath))
		if err != nil {
			return err
		}
		if !found {
			log.Infof("path %s not found in helm Value.yaml tree, skip mapping.", inPath)
			continue
		}
		if mstr, ok := m.(string); ok && mstr == "" {
			log.Infof("path %s is empty string, skip mapping.", inPath)
			continue
		}
		// Zero int values are due to proto3 compiling to scalars rather than ptrs. Skip these because values of 0 are
		// the default in destination fields and need not be set explicitly.
		if mint, ok := util.ToIntValue(m); ok && mint == 0 {
			log.Infof("path %s is int 0, skip mapping.", inPath)
			continue
		}

		path := util.ToYAMLPath(v.outPath)
		log.Infof("path has value in helm Value.yaml tree, mapping to output path %s", path)

		if err := tpath.WriteNode(cpSpecTree, path, m); err != nil {
			return err
		}
	}
	return nil
}

// renderComponentName renders a template of the form <path>{{.ComponentName}}<path> with
// the supplied parameters.
func renderComponentName(tmpl string, componentName string) string {
	type temp struct {
		ValueComponentName string
	}
	return renderTemplate(tmpl, temp{componentName})
}
