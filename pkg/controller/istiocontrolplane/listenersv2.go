package istiocontrolplane

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/helmreconciler"
)

// IstioStatusUpdater is a RenderingListener that updates the status field on the IstioControlPlane
// instance based on the results of the Reconcile operation.
type IstioStatusUpdaterV2 struct {
	*helmreconciler.DefaultRenderingListener
	instance   *v1alpha2.IstioControlPlane
	reconciler *helmreconciler.ISCPReconciler
}

// NewIstioRenderingListener returns a new IstioRenderingListener, which is a composite that includes IstioStatusUpdater
// and IstioChartCustomizerListener.
func NewIstioRenderingListenerV2(instance *v1alpha2.IstioControlPlane) *IstioRenderingListener {
	return &IstioRenderingListener{
		&helmreconciler.CompositeRenderingListener{
			Listeners: []helmreconciler.RenderingListener{
				NewChartCustomizerListener(),
				NewIstioStatusUpdaterV2(instance),
			},
		},
	}
}

// NewIstioStatusUpdater returns a new IstioStatusUpdater instance for the specified IstioControlPlane
func NewIstioStatusUpdaterV2(instance *v1alpha2.IstioControlPlane) helmreconciler.RenderingListener {
	return &IstioStatusUpdaterV2{
		DefaultRenderingListener: &helmreconciler.DefaultRenderingListener{},
		instance:                 instance,
	}
}

// EndReconcile updates the status field on the IstioControlPlane instance based on the resulting err parameter.
//TODO
func (u *IstioStatusUpdaterV2) EndReconcile(_ runtime.Object, err error) error {
	status := u.instance.Status
	vstatus := &v1alpha2.InstallStatus_VersionStatus{
		Status: v1alpha2.InstallStatus_HEALTHY,
	}
	if err != nil {
		vstatus.Status = v1alpha2.InstallStatus_ERROR
	}
	status.TrafficManagement = vstatus
	status.ConfigManagement = vstatus
	status.PolicyTelemetry = vstatus
	status.Security = vstatus
	status.IngressGateway = vstatus
	status.EgressGateway = vstatus
	return u.reconciler.Helmreconciler.GetClient().Status().Update(context.TODO(), u.instance)
}

// RegisterReconciler registers the HelmReconciler with this object
func (u *IstioStatusUpdaterV2) RegisterReconciler(reconciler *helmreconciler.ISCPReconciler) {
	u.reconciler = reconciler
}
