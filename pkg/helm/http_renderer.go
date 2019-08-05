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

package helm

import (
	"fmt"
	"time"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"istio.io/pkg/log"
)

// HTTPTemplateRenderer is a helm template renderer for helm charts hosted on http url
type HTTPTemplateRenderer struct {
	namespace        string
	componentName    string
	helmChartDirPath string
	chart            *chart.Chart
	started          bool
	globalValues     string
	pollInterval     time.Duration
}

// NewHTTPTemplateRenderer creates a TemplateRenderer with the given parameters and returns a pointer to it.
// helmChartDirPath must be an absolute HTTP URL hosting root of helm charts.
func NewHTTPTemplateRenderer(helmChartDirPath, globalValues, componentName, namespace string, pollInterval time.Duration) *HTTPTemplateRenderer {
	log.Infof("NewHTTPTemplateRenderer with helmChart=%s, componentName=%s, pollInterval=%v", helmChartDirPath, componentName, pollInterval)
	return &HTTPTemplateRenderer{
		namespace:        namespace,
		componentName:    componentName,
		helmChartDirPath: helmChartDirPath,
		globalValues:     globalValues,
		pollInterval:     pollInterval,
	}
}

// Run implements the TemplateRenderer interface.
func (h *HTTPTemplateRenderer) Run() error {
	log.Infof("Run HTTPTemplateRenderer with helmChart=%s, componentName=%s", h.helmChartDirPath, h.componentName)
	if err := h.loadChart(); err != nil {
		return err
	}

	chartChanged, err := PollURL(h.helmChartDirPath, h.pollInterval)
	if err != nil {
		return err
	}

	go func() {
		for range chartChanged {
			if err := h.loadChart(); err != nil {
				log.Errorf("Failed to load chart: %v", err.Error())
			}
		}
	}()
	h.started = true
	return nil
}

// RenderManifest renders the current helm templates with the current values and returns the resulting YAML manifest string.
func (h *HTTPTemplateRenderer) RenderManifest(values string) (string, error) {
	if !h.started {
		return "", fmt.Errorf("HTTPTemplateRenderer for %s not started in renderChart", h.componentName)
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
