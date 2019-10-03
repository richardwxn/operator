package istiocontrolplane

import (
	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/helmreconciler"
)

// IstioRenderingInput is a RenderingInput specific to an v1alpha2 IstioControlPlane instance.
type IstioRenderingInput struct {
	instance *v1alpha2.IstioControlPlane
	crPath   string
}

// NewIstioRenderingInput creates a new IstioRenderiongInput for the specified instance.
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
func (i *IstioRenderingInput) GetProcessingOrder(manifests helmreconciler.ChartManifestsMap) ([]string, error) {
	ordering := make([]string, 0, len(manifests))

	// everything else to the end
	for chart := range manifests {
		ordering = append(ordering, chart)
	}
	return ordering, nil
}
