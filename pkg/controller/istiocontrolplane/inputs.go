package istiocontrolplane

import (
	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/helmreconciler"
	"istio.io/operator/pkg/name"
)

var (
	componentDependencies = helmreconciler.ComponentNameToListMap{
		name.IstioBaseComponentName: {
			name.PilotComponentName,
			name.PolicyComponentName,
			name.TelemetryComponentName,
			name.GalleyComponentName,
			name.CitadelComponentName,
			name.NodeAgentComponentName,
			name.CertManagerComponentName,
			name.SidecarInjectorComponentName,
			name.IngressComponentName,
			name.EgressComponentName,
		},
	}

	installTree      = make(helmreconciler.ComponentTree)
	dependencyWaitCh = make(helmreconciler.DependencyWaitCh)
)

func init() {
	buildInstallTree()
	for _, parent := range componentDependencies {
		for _, child := range parent {
			dependencyWaitCh[child] = make(chan struct{}, 1)
		}
	}

}

// IstioRenderingInput is a RenderingInput specific to an v1alpha2 IstioControlPlane instance.
type IstioRenderingInput struct {
	instance *v1alpha2.IstioControlPlane
	crPath   string
}

// NewIstioRenderingInput creates a new IstioRenderingInput for the specified instance.
func NewIstioRenderingInput(instance *v1alpha2.IstioControlPlane) *IstioRenderingInput {
	return &IstioRenderingInput{instance: instance}
}

// GetCRPath returns the path of IstioControlPlane CR.
func (i *IstioRenderingInput) GetChartPath() string {
	return i.crPath
}

func (i *IstioRenderingInput) GetInputConfig() interface{} {
	// Not used in this renderer,
	return nil
}

func (i *IstioRenderingInput) GetTargetNamespace() string {
	return i.instance.Spec.DefaultNamespace
}

// GetProcessingOrder returns the order in which the rendered charts should be processed.
func (i *IstioRenderingInput) GetProcessingOrder(_ helmreconciler.ChartManifestsMap) (helmreconciler.ComponentNameToListMap, helmreconciler.DependencyWaitCh) {
	return componentDependencies, dependencyWaitCh
}

func buildInstallTree() {
	// Starting with root, recursively insert each first level child into each node.
	helmreconciler.InsertChildrenRecursive(name.IstioBaseComponentName, installTree, componentDependencies)
}
