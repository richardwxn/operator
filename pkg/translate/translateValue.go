package translate

import (
	"fmt"

	"github.com/ghodss/yaml"

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
	// ComponentToHelmValuesName is the root component name used in values YAML files in component charts.
	ValuesToComponentName map[string]name.ComponentName
	//// NamespaceMapping maps namespace defined in value.yaml to that in API spec.
	NamespaceAndEnablementMapping map[string]*Translation
	// ComponentDirLayout is a mapping between the subdirectory of the component Chart a component name.
	ComponentDirLayout map[string]name.ComponentName
}


var (
	ValueTranslators = map[version.MinorVersion]*ValueYAMLTranslator{
		version.MinorVersion{Major: 1, Minor: 2}: {
			APIMapping: map[string]*Translation{
				"global.hub":                         {"Hub", nil},
				"global.tag":                         {"Tag", nil},
				"global.proxy":                       {"TrafficManagement.Components.Proxy.Common.Values", nil},
				"global.policyCheckFailOpen":         {"PolicyTelemetry.PolicyCheckFailOpen", nil},
				"global.outboundTrafficPolicy.mode":  {"PolicyTelemetry.OutboundTrafficPolicyMode", nil},
				"global.controlPlaneSecurityEnabled": {"Security.ControlPlaneMtls.Value", nil},
				"global.mtls.enabled":                {"Security.DataPlaneMtlsStrict.Value", nil},
			},
			KubernetesMapping: map[string]*Translation{
				"[Deployment:{{.ResourceName}}].spec.template.spec.containers.[name:{{.ContainerName}}].env":{" {.FeatureName}}.Components.{{.ComponentName}}.Common.K8S.Env", nil},
			},
			ValuesToComponentName: map[string]name.ComponentName{
				"global":                        name.IstioBaseComponentName,
				"pilot":                         name.PilotComponentName,
				"galley":                        name.GalleyComponentName,
				"sidecarInjectorWebhook":        name.SidecarInjectorComponentName,
				"mixer.policy":                  name.PolicyComponentName,
				"mixer.telemetry":               name.TelemetryComponentName,
				"citadel":                       name.CitadelComponentName,
				"nodeAgent":                     name.NodeAgentComponentName,
				"certManager":                   name.CertManagerComponentName,
				"gateways.istio-ingressgateway": name.IngressComponentName,
				"gateways.istio-egressgateway":  name.EgressComponentName,
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
			NamespaceAndEnablementMapping: map[string]*Translation{
				"global.telemetryNamespace": 					{"telemetry.Components.Namespace", nil},
				"mixer.telemetry.enabled": 						{"telemetry.enabled.value", nil},
				"mixer.policy.enabled": 						  {"policy.enabled.value", nil},
				"global.policyNamespace": 					  {"policy.Components.Namespace", nil},
				"global.configNamespace": 					  {"configManagement.Components.Namespace", nil},
				"global.securityNamespace": 					{"security.Components.Namespace", nil},
			},
		},
	}
)

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
	var cpSpec = &v1alpha2.IstioControlPlaneSpec{}
	err = yaml.Unmarshal(outputVal, cpSpec)

	if err != nil {
		return nil, fmt.Errorf("error when unmarshalling into control plane spec %v", err.Error())
	}

	return cpSpec, nil
}

// TranslateTree translates input value.yaml Tree to ControlPlaneSpec Tree
func (t *ValueYAMLTranslator) TranslateTree(valueTree map[string]interface{}, cpSpecTree map[string]interface{}, path util.Path) error {

	// translate with api mapping
	err := t.translateTree(valueTree, cpSpecTree, path, t.APIMapping)
	if err != nil {
		return fmt.Errorf("error when translating value.yaml tree with global mapping %v", err.Error())
	}
	// translate with k8s mapping
	err = t.translateTree(valueTree, cpSpecTree, path, t.KubernetesMapping)
	if err != nil {
		return fmt.Errorf("error when translating value.yaml tree with kubernetes mapping %v", err.Error())
	}
	// translate with namespace and enablement mapping
	err = t.translateTree(valueTree, cpSpecTree, path, t.NamespaceAndEnablementMapping)
	if err != nil {
		return fmt.Errorf("error when translating value.yaml tree with kubernetes mapping %v", err.Error())
	}

	return nil
}

//internal method for TranslateTree
func (t *ValueYAMLTranslator) translateTree(valueTree map[string]interface{},
	cpSpecTree map[string]interface{}, path util.Path, mapping map[string]*Translation) error {
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
				err := t.translateTree(test, cpSpecTree, newPath, mapping)
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
		errs = util.AppendErr(errs, defaultTranslationFunc(m, root, valuesPath, val))
	default:
		// Use a custom translation function.
		errs = util.AppendErr(errs, m.translationFunc(m, root, valuesPath, val))
	}
	return errs
}

// setTree sets the YAML path in the given Tree to the given value, creating any required intermediate nodes.
func setTree(root map[string]interface{}, path util.Path, value interface{}) error {
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
	setTree(root[path[0]].(map[string]interface{}), path[1:], value)
	return nil
}

