package helm

import (
	"fmt"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"istio.io/pkg/log"
)

// HTTPTemplateRenderer is a helm template renderer for a local filesystem.
type HTTPTemplateRenderer struct {
	namespace        string
	componentName    string
	helmChartDirPath string
	chart            *chart.Chart
	started          bool
	globalValues     string
}

// NewFileTemplateRenderer creates a TemplateRenderer with the given parameters and returns a pointer to it.
// helmChartDirPath must be an absolute file path to the root of the helm charts.
func NewHTTPTemplateRenderer(helmChartDirPath, globalValues, componentName, namespace string) *FileTemplateRenderer {
	log.Infof("NewFileTemplateRenderer with helmChart=%s, componentName=%s", helmChartDirPath, componentName)
	return &FileTemplateRenderer{
		namespace:        namespace,
		componentName:    componentName,
		helmChartDirPath: helmChartDirPath,
		globalValues:     globalValues,
	}
}

// Run implements the TemplateRenderer interface.
func (h *HTTPTemplateRenderer) Run() error {
	//



	h.started = true
	return nil
}

// RenderManifest renders the current helm templates with the current values and returns the resulting YAML manifest string.
func (h *HTTPTemplateRenderer) RenderManifest(values string) (string, error) {
	if !h.started {
		return "", fmt.Errorf("fileTemplateRenderer for %s not started in renderChart", h.componentName)
	}
	return renderChart(h.namespace, h.globalValues, values, h.chart)
}

// loadChart implements the TemplateRenderer interface.
func (h *HTTPTemplateRenderer) loadChart() error {
	var err error
	if h.chart, err = chartutil.Load(h.helmChartDirPath); err != nil {
		return err
	}
	return nil
}

