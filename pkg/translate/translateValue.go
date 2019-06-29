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
	"bytes"
	"fmt"
	"html/template"

	"github.com/ghodss/yaml"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/name"
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
	ValuesToFeatureComponentName map[string]FeatureComponent
	// NamespaceMapping maps namespace defined in value.yaml to that in API spec.
	NamespaceMapping map[string]*Translation
	// FeatureEnablementMapping maps component enablement in value.yaml to feature enablement in API spec.
	FeatureEnablementMapping map[string]*Translation
	// ComponentEnablementMapping maps component enablement in value.yaml to component enablement in API spec.
	ComponentEnablementMapping map[string]*Translation
	// ComponentDirLayout is a mapping between the subdirectory of the component Chart a component name.
	ComponentDirLayout map[string]name.ComponentName
}

type FeatureComponent struct {
	featureName   name.FeatureName
	componentName name.ComponentName
}

type mappingType string

const (
	K8SMappingType      mappingType = "k8sMapping"
	APIMappingType      mappingType = "APIMapping"
	NSENABLEMappingType mappingType = "NamespaceEnableMapping"
)

var (
	ValueTranslators = map[version.MinorVersion]*ValueYAMLTranslator{
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
			ValuesToFeatureComponentName: map[string]FeatureComponent{
				"pilot":                         {name.TrafficManagementFeatureName, name.PilotComponentName},
				"galley":                        {name.ConfigManagementFeatureName, name.GalleyComponentName},
				"sidecarInjectorWebhook":        {name.AutoInjectionFeatureName, name.SidecarInjectorComponentName},
				"mixer.policy":                  {name.PolicyFeatureName, name.PolicyComponentName},
				"mixer.telemetry":               {name.TelemetryFeatureName, name.TelemetryComponentName},
				"citadel":                       {name.SecurityFeatureName, name.CitadelComponentName},
				"nodeagent":                     {name.SecurityFeatureName, name.NodeAgentComponentName},
				"certmanager":                   {name.SecurityFeatureName, name.CertManagerComponentName},
				"gateways.istio-ingressgateway": {"", name.IngressComponentName},
				"gateways.istio-egressgateway":  {"", name.EgressComponentName},
			},
			ComponentDirLayout: map[string]name.ComponentName{
				"istio-control/istio-discovery":  name.PilotComponentName,
				"istio-control/istio-config":     name.GalleyComponentName,
				"istio-control/istio-autoinject": name.SidecarInjectorComponentName,
				"istio-policy":                   name.PolicyComponentName,
				"istio-telemetry":                name.TelemetryComponentName,
				"security/citadel":               name.CitadelComponentName,
				"security/nodeagent":             name.NodeAgentComponentName,
				"security/certmanager":           name.CertManagerComponentName,
				"gateways/istio-ingress":         name.IngressComponentName,
				"gateways/istio-egress":          name.EgressComponentName,
			},
			NamespaceMapping: map[string]*Translation{
				// only security components use it
				"global.istioNamespace":     {"security.Components.Namespace", nil},
				"global.telemetryNamespace": {"telemetry.Components.Namespace", nil},
				"global.policyNamespace":    {"policy.Components.Namespace", nil},
				"global.configNamespace":    {"configManagement.Components.Namespace", nil},
			},
			// Ex: "{{.ValueComponent}}.enabled": {"{{.FeatureName}}.enabled.value}", nil},
			FeatureEnablementMapping: map[string]*Translation{},
			// Ex "{{.ValueComponent}}.enabled": {"{{.FeatureName}}.Components.{{.ComponentName}}.Common.enabled.value}", nil},
			ComponentEnablementMapping: map[string]*Translation{},
		},
	}
	ComponentEnablementPattern = "{{.FeatureName}}.Components.{{.ComponentName}}.Common.enabled.value"
)

// init initialize different mappings
func init() {
	vs := version.NewMinorVersion(1, 2)
	vTranslator := ValueTranslators[vs]
	vTranslator.initAPIMapping(vs)
	vTranslator.initK8SMapping()
	vTranslator.initEnablementMapping()
}

// initAPIMapping generate the reverse mapping from original translator apimapping
func (t *ValueYAMLTranslator) initAPIMapping(vs version.MinorVersion) {
	originalAPIMap := Translators[vs].APIMapping
	for valKey, outVal := range originalAPIMap {
		t.APIMapping[outVal.outPath] = &Translation{valKey, nil}
	}
}

// initK8SMapping generates the k8s settings mapping for all components based on templates
func (t *ValueYAMLTranslator) initK8SMapping() {
	outPutMapping := make(map[string]*Translation)
	for valKey, featureComponent := range t.ValuesToFeatureComponentName {
		if featureComponent.featureName == "" {
			continue
		}
		for K8SValKey, outPathTmpl := range t.KubernetesMapping {
			newKey := componentString(K8SValKey, valKey)
			newVal := featureComponentString(outPathTmpl.outPath, featureComponent.featureName, featureComponent.componentName)
			outPutMapping[newKey] = &Translation{newVal, nil}
		}
	}
	t.KubernetesMapping = outPutMapping
}

// initEnablementMapping generates the feature and component enablement mapping based on templates
func (t *ValueYAMLTranslator) initEnablementMapping() {
	feMapping := make(map[string]*Translation)
	ceMapping := make(map[string]*Translation)
	for valKey, featureComponent := range t.ValuesToFeatureComponentName {
		// construct feature enablement mapping
		newKey := valKey + ".enabled"
		newFEVal := string(featureComponent.featureName) + ".enabled.value"
		feMapping[newKey] = &Translation{newFEVal, nil}
		// construct component enablement mapping
		newCEVal := featureComponentString(ComponentEnablementPattern, featureComponent.featureName, featureComponent.componentName)
		ceMapping[newKey] = &Translation{newCEVal, nil}
	}
	t.FeatureEnablementMapping = feMapping
	t.ComponentEnablementMapping = ceMapping
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
	err = yaml.Unmarshal(outputVal, &cpSpec)

	if err != nil {
		return nil, fmt.Errorf("error when unmarshalling into control plane spec %v", err.Error())
	}

	return cpSpec, nil
}

// TranslateTree translates input value.yaml Tree to ControlPlaneSpec Tree
func (t *ValueYAMLTranslator) TranslateTree(valueTree map[string]interface{}, cpSpecTree map[string]interface{}, path util.Path) error {

	// translate with api mapping
	err := t.translateTree(valueTree, cpSpecTree, path, t.APIMapping, APIMappingType)
	if err != nil {
		return fmt.Errorf("error when translating value.yaml tree with global mapping %v", err.Error())
	}
	// translate with k8s mapping
	err = t.translateTree(valueTree, cpSpecTree, path, t.KubernetesMapping, K8SMappingType)
	if err != nil {
		return fmt.Errorf("error when translating value.yaml tree with kubernetes mapping %v", err.Error())
	}
	// translate with namespace and enablement mapping
	err = t.translateTree(valueTree, cpSpecTree, path, t.NamespaceMapping, NSENABLEMappingType)
	if err != nil {
		return fmt.Errorf("error when translating value.yaml tree with kubernetes mapping %v", err.Error())
	}
	err = t.translateTree(valueTree, cpSpecTree, path, t.FeatureEnablementMapping, NSENABLEMappingType)
	if err != nil {
		return fmt.Errorf("error when translating value.yaml tree with kubernetes mapping %v", err.Error())
	}
	err = t.translateTree(valueTree, cpSpecTree, path, t.ComponentEnablementMapping, NSENABLEMappingType)
	if err != nil {
		return fmt.Errorf("error when translating value.yaml tree with kubernetes mapping %v", err.Error())
	}

	return nil
}

//internal method for TranslateTree
func (t *ValueYAMLTranslator) translateTree(valueTree map[string]interface{},
	cpSpecTree map[string]interface{}, path util.Path, mapping map[string]*Translation, mpType mappingType) error {
	// translate input valueTree
	for key, val := range valueTree {
		newPath := append(path, key)
		// leaf
		if val == nil {
			err := t.insertLeaf(cpSpecTree, newPath, val, mapping)
			if err != nil {
				return err
			}
		} else {
			switch test := val.(type) {
			case map[string]interface{}:
				err := t.translateTree(test, cpSpecTree, newPath, mapping, mpType)
				if err != nil {
					return err
				}
			default:
				err := t.insertLeaf(cpSpecTree, newPath, val, mapping)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *ValueYAMLTranslator) insertLeaf(root map[string]interface{}, newPath util.Path,
	val interface{}, mapping map[string]*Translation) (errs util.Errors) {
	// Must be a scalar leaf. See if we have a mapping.
	valuesPath, m := getValuesPathMapping(mapping, newPath)
	switch {
	case m == nil:
		break
	case m.translationFunc == nil:
		// Use default translation which just maps to a different part of the tree.
		errs = util.AppendErr(errs, defaultValueTranslationFunc(m, root, valuesPath, val))
	default:
		// Use a custom translation function.
		errs = util.AppendErr(errs, m.translationFunc(m, root, valuesPath, val))
	}
	return errs
}

// defaultTranslationFunc is the default translation to values. It maps a Go data path into a YAML path.
func defaultValueTranslationFunc(m *Translation, root map[string]interface{}, valuesPath string, value interface{}) error {
	var path []string

	if util.IsEmptyString(value) {
		dbgPrint("Skip empty string value for path %s", m.outPath)
		return nil
	}
	if valuesPath == "" {
		dbgPrint("Not mapping to values, resources path is %s", m.outPath)
		return nil
	}

	for _, p := range util.PathFromString(valuesPath) {
		path = append(path, firstCharToLower(p))
	}

	return setUntypedTree(root, path, value)
}

// componentString renders a template of the form <path>{{.ComponentName}}<path> with
// the supplied parameters.
func componentString(tmpl string, componentName string) string {
	type temp struct {
		ValueComponentName string
	}
	return renderTemplate(tmpl, temp{componentName})
}

// helper method to render template
func renderTemplate(tmpl string, ts interface{}) string {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		log.Error(err.Error())
		return err.Error()
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, ts)
	if err != nil {
		log.Error(err.Error())
		return err.Error()
	}
	return buf.String()
}

// setTree sets the YAML path in the given Tree to the given value, creating any required intermediate nodes.
func setUntypedTree(root map[string]interface{}, path util.Path, value interface{}) error {
	dbgPrint("setTree %s:%v", path, value)
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}
	if len(path) == 1 {
		root[path[0]] = value
		return nil
	}
	if root[path[0]] == nil {
		root[path[0]] = make(map[string]interface{})
	}
	return setUntypedTree(root[path[0]].(map[string]interface{}), path[1:], value)
}
