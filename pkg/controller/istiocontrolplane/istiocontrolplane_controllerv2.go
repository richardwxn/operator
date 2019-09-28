package istiocontrolplane

import (
	"context"
	"github.com/go-logr/logr"
	"istio.io/operator/pkg/apis/istio/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"istio.io/operator/pkg/helmreconciler"
)

// ReconcileIstioControlPlane reconciles a v1alpha2 IstioControlPlane object
type ReconcileIstioControlPlaneV2 struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	factory  *helmreconciler.Factory
	instance runtime.Object
}

// AddV2 creates a new IstioControlPlane Controller V2 and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func AddV2(mgr manager.Manager) error {
	return add(mgr, newReconcilerV2(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconcilerV2(mgr manager.Manager) reconcile.Reconciler {
	factory := &helmreconciler.Factory{CustomizerFactory: &IstioRenderingCustomizerFactory{}}
	return &ReconcileIstioControlPlaneV2{client: mgr.GetClient(), scheme: mgr.GetScheme(), factory: factory}
}

func handleDeleted(g runtime.Object, reqLogger logr.Logger, client client.Client, factory helmreconciler.Factory) {
	instance, ok := g.(*v1alpha1.IstioControlPlane)
	if !ok {

	}
	deleted := instance.GetDeletionTimestamp() != nil
	finalizers := instance.GetFinalizers()
	finalizerIndex := indexOf(finalizers, finalizer)

	if deleted {
		if finalizerIndex < 0 {
			reqLogger.Info("IstioControlPlane deleted")
			return reconcile.Result{}, nil
		}
		reqLogger.Info("Deleting IstioControlPlane")

		reconciler, err := factory.NewISCPReconciler(instance, client, reqLogger)
		if err == nil {
			err = reconciler.Delete()
		} else {
			reqLogger.Error(err, "failed to create reconciler")
		}
		// XXX: for now, nuke the resources, regardless of errors
		finalizers = append(finalizers[:finalizerIndex], finalizers[finalizerIndex+1:]...)
		instance.SetFinalizers(finalizers)
		finalizerError := r.client.Update(context.TODO(), instance)
		for retryCount := 0; errors.IsConflict(finalizerError) && retryCount < 5; retryCount++ {
			// workaround for https://github.com/kubernetes/kubernetes/issues/73098 for k8s < 1.14
			reqLogger.Info("confilict during finalizer removal, retrying")
			_ = client.Get(context.TODO(), request.NamespacedName, instance)
			finalizers = instance.GetFinalizers()
			finalizerIndex = indexOf(finalizers, finalizer)
			finalizers = append(finalizers[:finalizerIndex], finalizers[finalizerIndex+1:]...)
			instance.SetFinalizers(finalizers)
			finalizerError = r.client.Update(context.TODO(), instance)
		}
		if finalizerError != nil {
			reqLogger.Error(finalizerError, "error removing finalizer")
		}
		return reconcile.Result{}, err
	} else if finalizerIndex < 0 {
		reqLogger.V(1).Info("Adding finalizer", "finalizer", finalizer)
		finalizers = append(finalizers, finalizer)
		instance.SetFinalizers(finalizers)
		err = r.client.Update(context.TODO(), instance)
		return reconcile.Result{}, err
	}

	reqLogger.Info("Updating IstioControlPlane")
	reconciler, err := factory.NewISCPReconciler(instance, r.client, reqLogger)
	if err == nil {
		err = reconciler.Reconcile()
	} else {
		reqLogger.Error(err, "failed to create reconciler")
	}

	return reconcile.Result{}, err
}

// Reconcile reads that state of the cluster for a v1alpha2 IstioControlPlane object and makes changes based on the state read
// and what is in the IstioControlPlane.Spec
func (r *ReconcileIstioControlPlaneV2) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling IstioControlPlane")

	// Fetch the IstioControlPlane instance
	instance := &v1alpha1.IstioControlPlane{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}


	reqLogger.Info("Updating IstioControlPlane")
	reconciler, err := r.factory.NewISCPReconciler(instance, r.client, reqLogger)
	if err == nil {
		err = reconciler.Reconcile()
	} else {
		reqLogger.Error(err, "failed to create reconciler")
	}

	return reconcile.Result{}, err
}
