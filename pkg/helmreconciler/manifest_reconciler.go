package helmreconciler

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ISCPReconciler reconciles resources rendered by a set of helm charts for a specific instances of a custom resource,
// or deletes all resources associated with a specific instance of a custom resource.
type ISCPReconciler struct {
	Helmreconciler *HelmReconciler
}

// NewISCPReconciler Returns a new ISCPReconciler for the custom resource.
// instance is the custom resource to be reconciled/deleted.
// client is the kubernetes client
// logger is the logger
func (f *Factory) NewISCPReconciler(instance runtime.Object, client client.Client, logger logr.Logger) (*ISCPReconciler, error) {
	h, err := f.New(instance, client, logger)
	if err != nil {
		return &ISCPReconciler{}, err
	}
	reconciler := &ISCPReconciler{Helmreconciler: h}
	return reconciler, nil
}


// Reconcile the resources associated with the custom resource instance.
func (rc *ISCPReconciler) Reconcile() error {
	h := rc.Helmreconciler
	log := h.GetLogger()
	// any processing required before processing the charts
	if err := h.GetCustomizer().Listener().BeginReconcile(h.GetInstance()); err != nil {
		return err
	}

	// render component manifests mapping
	manifestMap, err := rc.renderManifests(h.GetCustomizer().Input())
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("error rendering charts"))
		listenerErr := h.GetCustomizer().Listener().EndReconcile(h.GetInstance(), err)
		if listenerErr != nil {
			log.Error(listenerErr, "unexpected error invoking EndReconcile")
		}
		return err
	}

	// determine processing order
	chartOrder, err := h.customizer.Input().GetProcessingOrder(manifestMap)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("error ordering charts"))
		listenerErr := h.customizer.Listener().EndReconcile(h.instance, err)
		if listenerErr != nil {
			h.logger.Error(listenerErr, "unexpected error invoking EndReconcile")
		}
		return err
	}

	// collect the errors.  from here on, we'll process everything with the assumption that any error is not fatal.
	allErrors := []error{}

	// process the charts
	for _, chartName := range chartOrder {
		chartManifests, ok := manifestMap[chartName]
		if !ok {
			// TODO: log warning about missing chart
			continue
		}
		chartManifests, err := h.GetCustomizer().Listener().BeginChart(chartName, chartManifests)
		if err != nil {
			allErrors = append(allErrors, err)
		}
		err = h.ProcessManifests(chartManifests)
		if err != nil {
			allErrors = append(allErrors, err)
		}
		err = h.GetCustomizer().Listener().EndChart(chartName)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	// delete any obsolete resources
	err = h.GetCustomizer().Listener().BeginPrune(false)
	if err != nil {
		allErrors = append(allErrors, err)
	}
	err = h.Prune(false)
	if err != nil {
		allErrors = append(allErrors, err)
	}
	err = h.GetCustomizer().Listener().EndPrune()
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// any post processing required after updating
	err = h.GetCustomizer().Listener().EndReconcile(h.GetInstance(), utilerrors.NewAggregate(allErrors))
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// return any errors
	return utilerrors.NewAggregate(allErrors)
}

// Delete resources associated with the custom resource instance
func (rc *ISCPReconciler) Delete() error {
	return rc.Helmreconciler.Delete()
}
