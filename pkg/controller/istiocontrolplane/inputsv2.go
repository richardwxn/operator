package istiocontrolplane

import (
	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/helmreconciler"
)

// IstioRenderingInput is a RenderingInput specific to an v1alpha2 IstioControlPlane instance.
type IstioRenderingInputV2 struct {
	instance *v1alpha2.IstioControlPlane
	crPath   string
}

// NewIstioRenderingInput creates a new IstioRenderiongInput for the specified instance.
func NewIstioRenderingInputV2(instance *v1alpha2.IstioControlPlane) *IstioRenderingInputV2 {
	return &IstioRenderingInputV2{instance: instance}
}

// GetCRPath returns the path of IstioControlPlane CR.
func (i *IstioRenderingInputV2) GetCRPath() string {
	return i.crPath
}

// GetProcessingOrder returns the order in which the rendered charts should be processed.
func (i *IstioRenderingInputV2) GetProcessingOrder(manifests helmreconciler.ChartManifestsMap) ([]string, error) {
	seen := map[string]struct{}{}
	ordering := make([]string, 0, len(manifests))
	// known ordering
	for _, chart := range defaultProcessingOrder {
		if _, ok := manifests[chart]; ok {
			ordering = append(ordering, chart)
			seen[chart] = struct{}{}
		}
	}
	// everything else to the end
	for chart := range manifests {
		if _, ok := seen[chart]; !ok {
			ordering = append(ordering, chart)
			seen[chart] = struct{}{}
		}
	}
	return ordering, nil
}
