package helmreconciler

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	binversion "istio.io/operator/version"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/releaseutil"

	"istio.io/operator/pkg/component/controlplane"
	operatormanifest "istio.io/operator/pkg/manifest"
	"istio.io/operator/pkg/name"
	"istio.io/operator/pkg/translate"
)

var (
	kindRegex = regexp.MustCompile("kind:(.*)\n")
	componentToChartName = map[name.ComponentName]string{
		name.IstioBaseComponentName:       "istio",
		name.PilotComponentName:           "istio/charts/pilot",
		name.GalleyComponentName:          "istio/charts/galley",
		name.SidecarInjectorComponentName: "istio/charts/sidecarInjectorWebhook",
		name.PolicyComponentName:          "istio/charts/mixer-policy",
		name.TelemetryComponentName:       "istio/charts/mixer-telemetry",
		name.CitadelComponentName:         "istio/charts/security-citadel",
		name.CertManagerComponentName:     "istio/charts/security-certmanager",
		name.NodeAgentComponentName:       "istio/charts/security-nodeagent",
		name.IngressComponentName:         "istio/charts/gateways-istio-engress",
		name.EgressComponentName:          "istio/charts/gateways-istio-egress",
	}
)

func (rc *ISCPReconciler) renderManifests(input RenderingInputV2) (ChartManifestsMap, error) {
	cr, err := ioutil.ReadFile(input.GetCRPath())
	if err != nil {
		return nil, fmt.Errorf("could not read cr file %s: %s", input.GetCRPath(), err)
	}
	icps, _, err := operatormanifest.ParseK8SYAMLToIstioControlPlaneSpec(string(cr))
	if err != nil {
		return nil, fmt.Errorf("could not IstioControlPlane CR YAML:\n%s", cr)
	}
	t, err := translate.NewTranslator(binversion.OperatorBinaryVersion.MinorVersion)
	if err != nil {
		return nil, err
	}

	cp := controlplane.NewIstioControlPlane(icps, t)
	if err := cp.Run(); err != nil {
		return nil, fmt.Errorf("failed to create Istio control" +
			" plane with spec: \n%v\nerror: %s", icps, err)
	}

	manifests, errs := cp.RenderManifest()
	if errs != nil {
		return ChartManifestsMap{}, errs.ToError()
	}

	// convert to
	return convert(manifests)
}

// convert is helper function to converts from namespace:manifests mapping to chartname:[]manifests mapping.
func convert(manifestMap name.ManifestMap) (ChartManifestsMap, error) {
	var listManifests []manifest.Manifest
	// extract kind and name
	for cn, v := range manifestMap {
		match := kindRegex.FindStringSubmatch(v)
		h := "Unknown"
		if len(match) == 2 {
			h = strings.TrimSpace(match[1])
		}
		chartName := componentToChartName[cn]
		if chartName == "" {
			return nil, fmt.Errorf("no matching chart name for component name: %s", cn)
		}
		m := manifest.Manifest{Name: chartName, Content: v, Head: &releaseutil.SimpleHead{Kind: h}}
		listManifests = append(listManifests, m)
	}
	return CollectManifestsByChart(listManifests), nil
}

func (rc *ISCPReconciler) ProcessManifests(manifests []manifest.Manifest) error {
	allErrors := []error{}
	h := rc.Helmreconciler
	logger, origLogger := h.GetLogger(), h.GetLogger()
	defer func() { logger = origLogger }()
	for _, manifest := range manifests {
		logger = origLogger.WithValues("manifest", manifest.Name)
		logger.V(2).Info("Processing resources from manifest")
		// split the manifest into individual objects
		objects := releaseutil.SplitManifests(manifest.Content)
		for _, raw := range objects {
			rawJSON, err := yaml.YAMLToJSON([]byte(raw))
			if err != nil {
				logger.Error(err, "unable to convert raw data to JSON")
				allErrors = append(allErrors, err)
				continue
			}
			obj := &unstructured.Unstructured{}
			_, _, err = unstructured.UnstructuredJSONScheme.Decode(rawJSON, nil, obj)
			if err != nil {
				logger.Error(err, "unable to decode object into Unstructured")
				allErrors = append(allErrors, err)
				continue
			}
			if err = h.ProcessObject(obj); err != nil {
				allErrors = append(allErrors, err)
			}
		}
	}

	return utilerrors.NewAggregate(allErrors)
}
