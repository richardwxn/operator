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
	"github.com/ghodss/yaml"
	"istio.io/operator/pkg/tpath"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/name"
	"istio.io/operator/pkg/util"
	"istio.io/operator/pkg/version"
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
	// FeatureEnablementMapping maps component enablement in value.yaml to feature enablement in API spec.
	FeatureEnablementMapping map[string]*Translation
	// ComponentEnablementMapping maps component enablement in value.yaml to component enablement in API spec.
	ComponentEnablementMapping map[string]*Translation
}

// FeatureComponent represent featureName and componentName
type FeatureComponent struct {
	featureName   name.FeatureName
	componentName name.ComponentName
}

var (
	// ValueYAMLTranslators is Translator for value.yaml
	ValueYAMLTranslators = map[version.MinorVersion]*ValueYAMLTranslator{
		version.NewMinorVersion(1, 2): {
			APIMapping: map[string]*Translation{},
			KubernetesMapping: map[string]*Translation{
				// TODO use template for podaffinity
				"{{.ValueComponentName}}.podAntiAffinityLabelSelector": {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8S.Affinity", nil},
				"{{.ValueComponentName}}.env":                          {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8S.Env", nil},
				"{{.ValueComponentName}}.autoscaleEnabled":             {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8S.HpaSpec", nil},
				"{{.ValueComponentName}}.imagePullPolicy":              {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8S.ImagePullPolicy", nil},
				"{{.ValueComponentName}}.nodeSelector":                 {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8S.NodeSelector", nil},
				"{{.ValueComponentName}}.podDisruptionBudget":          {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8S.PodDisruptionBudget", nil},
				"{{.ValueComponentName}}.podAnnotations":               {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8S.PodAnnotations", nil},
				"{{.ValueComponentName}}.priorityClassName":            {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8S.PriorityClassName", nil},
				// TODO check readinessProbe mapping
				"{{.ValueComponentName}}.readinessProbe": {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8S.ReadinessProbe", nil},
				"{{.ValueComponentName}}.replicaCount":   {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8S.ReplicaCount", nil},
				"{{.ValueComponentName}}.resources":      {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.K8S.Resources", nil},
			},
			ValuesToComponentName: map[string]name.ComponentName{
				"pilot":                         name.PilotComponentName,
				"galley":                        name.GalleyComponentName,
				"sidecarInjectorWebhook":        name.SidecarInjectorComponentName,
				"mixer.policy":                  name.PolicyComponentName,
				"mixer.telemetry":               name.TelemetryComponentName,
				"citadel":                       name.CitadelComponentName,
				"nodeagent":                     name.NodeAgentComponentName,
				"certmanager":                   name.CertManagerComponentName,
				"gateways.istio-ingressgateway": name.IngressComponentName,
				"gateways.istio-egressgateway":  name.EgressComponentName,
			},
			// 1:N mapping possible for ns
			NamespaceMapping: map[string][]string{
				"global.istioNamespace":     {"security.components.namespace"},
				"global.telemetryNamespace": {"telemetry.components.namespace"},
				"global.policyNamespace":    {"policy.components.namespace"},
				"global.configNamespace":    {"configManagement.components.namespace"},
			},
			// Ex: "{{.ValueComponent}}.enabled": {"{{.FeatureName}}.enabled}", nil},
			FeatureEnablementMapping: map[string]*Translation{},
			// Ex "{{.ValueComponent}}.enabled": {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.enabled}", nil},
			// For gateway Ex "{{.ValueComponent}}.enabled": {"{{.FeatureName}}.Components.{{.ComponentName}}.Gateway.Common.enabled}", nil},
			ComponentEnablementMapping: map[string]*Translation{},
		},
	}
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

// initEnablementMapping generates the feature and component enablement mapping based on templates
func (t *ValueYAMLTranslator) initEnablementMapping() {
	ceMapping := make(map[string]*Translation)
	for valKey, componentName := range t.ValuesToComponentName {
		featureName := name.ComponentNameToFeatureName[componentName]
		// construct component enablement mapping
		newKey := valKey + ".enabled"
		newCEVal := renderFeatureComponentPathTemplate(componentEnablementPattern, featureName, componentName)
		ceMapping[newKey] = &Translation{newCEVal, nil}
	}
	t.ComponentEnablementMapping = ceMapping
}

// NewValueYAMLTranslator creates a new ValueYAMLTranslator for minorVersion and returns a ptr to it.
func NewValueYAMLTranslator(minorVersion version.MinorVersion) (*ValueYAMLTranslator, error) {
	t := ValueYAMLTranslators[minorVersion]
	if t == nil {
		return nil, fmt.Errorf("no translator available for version %s", minorVersion)
	}

	t.initAPIMapping(minorVersion)
	t.initK8SMapping()
	t.initEnablementMapping()
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

	// translate with api mapping
	err := translateTree(valueTree, cpSpecTree, path, t.APIMapping)
	if err != nil {
		return fmt.Errorf("error when translating value.yaml tree with global mapping %v", err.Error())
	}
	// translate with k8s mapping
	err = translateTree(valueTree, cpSpecTree, path, t.KubernetesMapping)
	if err != nil {
		return fmt.Errorf("error when translating value.yaml tree with kubernetes mapping %v", err.Error())
	}
	// translate enablement and namespace
	err = t.setEnablementAndNamespacesFromValue(cpSpecTree, valueTree)
	if err != nil {
		return fmt.Errorf("error when translating enablement and namespace from value.yaml tree %v", err.Error())
	}
	return nil
}

func firstCharToLowerPath(input string) util.Path {
	var path util.Path
	for _, p := range util.PathFromString(input) {
		p = firstCharToLower(p)
		path = append(path, p)
	}
	return path
}

// setEnablementAndNamespaces translates the enablement and namespace value of each component in the baseYAML values
// tree, based on feature/component inheritance relationship.
func (t *ValueYAMLTranslator) setEnablementAndNamespacesFromValue(root map[string]interface{}, valueSpec map[string]interface{}) error {
	for vp, fe := range t.FeatureEnablementMapping {
		enabled := name.IsComponentEnabledFromValue(vp, valueSpec)
		// set feature enablement
		if fe == nil || fe.outPath == "" {
			continue
		}
		newP := firstCharToLowerPath(fe.outPath)
		// Value.yaml component to IstioFeature is N:1, so if the feature is enabled by other component already, skip setting
		curEnabled, found, _ := name.GetFromValuePath(root, newP)
		if !found || curEnabled == false {
			if err := tpath.WriteNode(root, newP, enabled); err != nil {
				return err
			}
		}
		// set component enablement
		ce := t.ComponentEnablementMapping[vp]
		if ce == nil || ce.outPath == "" {
			continue
		}
		outP := firstCharToLowerPath(ce.outPath)
		if err := tpath.WriteNode(root, outP, enabled); err != nil {
			return err
		}
	}
	// set namespace
	for vp, nsList := range t.NamespaceMapping {
		namespace := name.NamespaceFromValue(vp, valueSpec)
		for _, ns := range nsList {
			if err := tpath.WriteNode(root, util.PathFromString(ns), namespace); err != nil {
				return err
			}
		}
	}
	return nil
}

//translateTree is internal method for translating value.yaml tree
func translateTree(valueTree map[string]interface{},
	cpSpecTree map[string]interface{}, path util.Path, mapping map[string]*Translation) error {
	// translate input valueTree
	for key, val := range valueTree {
		newPath := append(path, key)
		// leaf
		if val == nil {
			err := insertLeaf(cpSpecTree, newPath, val, mapping)
			if err != nil {
				return err
			}
		} else {
			switch node := val.(type) {
			case map[string]interface{}:
				err := translateTree(node, cpSpecTree, newPath, mapping)
				if err != nil {
					return err
				}
			case []interface{}:
				for _, newNode := range node {
					newMap, ok := newNode.(map[string]interface{})
					if !ok {
						return fmt.Errorf("fail to translate slice")
					}
					err := translateTree(newMap, cpSpecTree, newPath, mapping)
					if err != nil {
						return err
					}
				}
			// leaf
			default:
				err := insertLeaf(cpSpecTree, newPath, val, mapping)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func insertLeaf(root map[string]interface{}, newPath util.Path,
	val interface{}, mapping map[string]*Translation) (errs util.Errors) {
	// Must be a scalar leaf. See if we have a mapping.
	valuesPath, m := getValuesPathMapping(mapping, newPath)
	switch {
	case m == nil:
		break
	case m.translationFunc == nil:
		// Use default translation which just maps to a different part of the tree.
		errs = util.AppendErr(errs, defaultTranslationFunc(m, root, valuesPath, val))
	default:
		// Use a custom translation function.
		errs = util.AppendErr(errs, m.translationFunc(m, root, valuesPath, val))
	}
	return errs
}

// renderComponentName renders a template of the form <path>{{.ComponentName}}<path> with
// the supplied parameters.
func renderComponentName(tmpl string, componentName string) string {
	type temp struct {
		ValueComponentName string
	}
	return renderTemplate(tmpl, temp{componentName})
}
