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

package gotoproto

// TODO: create remaining enum types.

type Values struct {
	Global *GlobalConfig `json:"global,omitempty" protobuf:"bytes,1,opt,name=global"`
}

type GlobalConfig struct {
	Arch *ArchConfig `json:"arch,omitempty" protobuf:"bytes,30,opt,name=arch"`
	// Configuration validation component namespace
	ConfigNamespace *string `json:"configNamespace,omitempty" protobuf:"bytes,1,opt,name=configNamespace"`
	// Enables server-side validation of configuration.
	ConfigValidation *bool `json:"configValidation,omitempty" protobuf:"varint,2,opt,name=configValidation"`
	//Enables MTLS for control plane.
	ControlPlaneSecurityEnabled *bool `json:"controlPlaneSecurityEnabled,omitempty" protobuf:"varint,3,opt,name=controlPlaneSecurityEnabled"`
	// K8s NodeSelector
	DefaultNodeSelector map[string]string `json:"defaultNodeSelector,omitempty" protobuf:"bytes,31,rep,name=defaultNodeSelector"`
	//k8s PodDisruptionBudget settings.
	DefaultPodDisruptionBudget *DefaultPodDisruptionBudgetConfig `json:"defaultPodDisruptionBudget,omitempty" protobuf:"bytes,4,opt,name=defaultPodDisruptionBudget"`
	//Selects whether policy enforcement is installed.
	DisablePolicyChecks *bool                   `json:"disablePolicyChecks,omitempty" protobuf:"varint,5,opt,name=disablePolicyChecks"`
	DefaultResources    *DefaultResourcesConfig `json:"defaultResources,omitempty" protobuf:"bytes,6,opt,name=defaultResources"`
	// Enable helm test
	EnableHelmTest *bool `json:"enableHelmTest,omitempty" protobuf:"varint,7,opt,name=enableHelmTest"`
	// Enables tracing.
	EnableTracing *bool `json:"enableTracing,omitempty" protobuf:"varint,8,opt,name=enableTracing"`
	// Root for docker image paths.
	Hub *string `json:"hub,omitempty" protobuf:"bytes,9,opt,name=hub"`
	// Default namespace.
	IstioNamespace    *string                  `json:"istioNamespace,omitempty" protobuf:"bytes,10,opt,name=istioNamespace"`
	LocalityLbSetting map[string]string        `json:"localityLbSetting,omitempty" protobuf:"bytes,32,rep,name=localityLbSetting"`
	KubernetesIngress *KubernetesIngressConfig `json:"k8sIngress,omitempty" protobuf:"bytes,11,opt,name=k8sIngress"`
	Logging           *GlobalLoggingConfig     `json:"logging,omitempty" protobuf:"bytes,12,opt,name=logging"`
	MeshExpansion     *MeshExpansionConfig     `json:"meshExpansion,omitempty" protobuf:"bytes,13,opt,name=meshExpansion"`
	MeshNetworks      map[string]string        `json:"meshNetworks,omitempty" protobuf:"bytes,33,rep,name=meshNetworks"`
	// Monitor port number for all control plane components.
	MonitoringPort *uint32             `json:"monitoringPort,omitempty" protobuf:"varint,14,opt,name=monitoringPort"`
	MTLS           *MTLSConfig         `json:"mtls,omitempty" protobuf:"bytes,15,opt,name=mtls"`
	MultiCluster   *MultiClusterConfig `json:"multiCluster,omitempty" protobuf:"bytes,16,opt,name=multiCluster"`
	// Restricts the applications namespace that the controller manages.
	OneNamespace          *bool                        `json:"oneNamespace,omitempty" protobuf:"varint,17,opt,name=oneNamespace"`
	OutboundTrafficPolicy *OutboundTrafficPolicyConfig `json:"outboundTrafficPolicy,omitempty" protobuf:"bytes,18,opt,name=outboundTrafficPolicy"`
	//If set, allows traffic in cases when the mixer policy service cannot be reached.
	PolicyCheckFailOpen *bool `json:"policyCheckFailOpen,omitempty" protobuf:"varint,19,opt,name=policyCheckFailOpen"`
	//Namespace of policy components
	PolicyNamespace *string `json:"policyNamespace,omitempty" protobuf:"bytes,20,opt,name=policyNamespace"`
	//k8s priorityClassName.
	PriorityClassName *string          `json:"priorityClassName,omitempty" protobuf:"bytes,21,opt,name=priorityClassName"`
	Proxy             *ProxyConfig     `json:"proxy,omitempty" protobuf:"bytes,22,opt,name=proxy"`
	ProxyInit         *ProxyInitConfig `json:"proxy_init,omitempty" protobuf:"bytes,23,opt,name=proxy_init,json=proxyInit"`
	SDS               *SDSConfig       `json:"sds,omitempty" protobuf:"bytes,24,opt,name=sds"`
	//Version tag for docker images.
	Tag *string `json:"tag,omitempty" protobuf:"bytes,25,opt,name=tag"`
	// Namespace of telemetry components
	TelemetryNamespace *string       `json:"telemetryNamespace,omitempty" protobuf:"bytes,26,opt,name=telemetryNamespace"` // test
	Tracer             *TracerConfig `json:"tracer,omitempty" protobuf:"bytes,27,opt,name=tracer"`
	//Specifies the trust domain that corresponds to the root cert of CA.
	TrustDomain *string `json:"trustDomain,omitempty" protobuf:"bytes,28,opt,name=trustDomain"`
	// Selects use of Mesh Configuration Protocol to configure Pilot.
	UseMCP *bool `json:"useMCP,omitempty" protobuf:"varint,29,opt,name=useMCP"`
}

// ArchConfig is described in istio.io documentation.
type ArchConfig struct {
	// Sets pod scheduling weight for amd64 arch
	Amd64 *uint32 `json:"amd64,omitempty" protobuf:"varint,1,opt,name=amd64"`
	// Sets pod scheduling weight for ppc64le arch.
	Ppc64le *uint32 `json:"ppc64le,omitempty" protobuf:"varint,2,opt,name=ppc64le"`
	// Sets pod scheduling weight for s390x arch.
	S390x *uint32 `json:"s390x,omitempty" protobuf:"varint,3,opt,name=s390x"`
}

// DefaultPodDisruptionBudgetConfig is described in istio.io documentation.
type DefaultPodDisruptionBudgetConfig struct {
	//k8s PodDisruptionBudget settings.
	Enabled *bool `json:"enabled,omitempty" protobuf:"varint,1,opt,name=enabled"`
}

// DefaultResourcesConfig is described in istio.io documentation.
type DefaultResourcesConfig struct {
	// k8s resources settings.
	Requests *ResourcesRequestsConfig `json:"requests,omitempty" protobuf:"bytes,1,opt,name=requests"`
}

// KubernetesIngressConfig represents the configuration for Kubernetes Ingress.
type KubernetesIngressConfig struct {
	//Enables gateway for legacy k8s Ingress.
	Enabled *bool `json:"enabled,omitempty" protobuf:"varint,1,opt,name=enabled"`
	//Enables gateway for legacy k8s Ingress.
	EnableHTTPS *bool `json:"enableHttps,omitempty" protobuf:"varint,2,opt,name=enableHttps"`
	//Sets the gateway name for legacy k8s Ingress.
	GatewayName *string `json:"gatewayName,omitempty" protobuf:"bytes,3,opt,name=gatewayName"`
}

// GlobalLoggingConfig is described in istio.io documentation.
type GlobalLoggingConfig struct {
	Level *string `json:"level,omitempty" protobuf:"bytes,1,opt,name=level"`
}

// MeshExpansionConfig is described in istio.io documentation.
type MeshExpansionConfig struct {
	// Exposes Pilot and Citadel mTLS on the ingress gateway.
	Enabled *bool `json:"enabled,omitempty" protobuf:"varint,1,opt,name=enabled"`
	// Exposes Pilot and Citadel mTLS and the plain text Pilot ports on an internal gateway.
	UseILB *bool `json:"useILB,omitempty" protobuf:"varint,2,opt,name=useILB"`
}

// MTLSConfig is described in istio.io documentation.
type MTLSConfig struct {
	// Enables MTLS for service to service traffic.
	Enabled *bool `json:"enabled,omitempty" protobuf:"varint,1,opt,name=enabled"`
}

// MultiClusterConfig is described in istio.io documentation.
type MultiClusterConfig struct {
	// Enables the connection between two kubernetes clusters via their respective ingressgateway services. Use if the pods in each cluster cannot directly talk to one another.
	Enabled *bool `json:"enabled,omitempty" protobuf:"varint,1,opt,name=enabled"`
}

// OutboundTrafficPolicyConfig is described in istio.io documentation.
type OutboundTrafficPolicyConfig struct {
	// Specifies the sidecar's default behavior when handling outbound traffic from the application.
	Mode string `json:"mode,omitempty" protobuf:"bytes,1,opt,name=mode"`
}

// ProxyConfig specifies how proxies are configured within Istio.
type ProxyConfig struct {
	// Specifies the path to write the sidecar access log file.
	AccessLogFile *string `json:"accessLogFile,omitempty" protobuf:"bytes,1,opt,name=accessLogFile"`
	// Configures how and what fields are displayed in sidecar access log.
	AccessLogFormat   *string `json:"accessLogFormat,omitempty" protobuf:"bytes,2,opt,name=accessLogFormat"`
	AccessLogEncoding *string `json:"accessLogEncoding,omitempty" protobuf:"bytes,3,opt,name=accessLogEncoding"`
	AutoInject        *string `json:"autoInject,omitempty" protobuf:"bytes,4,opt,name=autoInject"`
	// Domain for the cluster - defaults to .cluster.local, but k8s allows this to be customized, can be prod.example.com
	ClusterDomain     *string `json:"clusterDomain,omitempty" protobuf:"bytes,5,opt,name=clusterDomain"`
	ComponentLogLevel *string `json:"componentLogLevel,omitempty" protobuf:"bytes,6,opt,name=componentLogLevel"`
	// Controls number of proxy worker threads.
	Concurrency *uint32 `json:"concurrency,omitempty" protobuf:"varint,7,opt,name=concurrency"`
	// Configures the DNS refresh rate for Envoy cluster of type STRICT_DNS.
	DNSRefreshRate *string `json:"dnsRefreshRate,omitempty" protobuf:"bytes,8,opt,name=dnsRefreshRate"`
	// Enables core dumps for newly injected sidecars.
	EnableCoreDump      *bool               `json:"enableCoreDump,omitempty" protobuf:"varint,9,opt,name=enableCoreDump"`
	EnvoyMetricsService *EnvoyMetricsConfig `json:"envoyMetricsService,omitempty" protobuf:"bytes,10,opt,name=envoyMetricsService"`
	EnvoyStatsD         *EnvoyMetricsConfig `json:"envoyStatsd,omitempty" protobuf:"bytes,11,opt,name=envoyStatsd"`
	// Specifies the Istio ingress ports not to capture.
	ExcludeInboundPorts *string `json:"excludeInboundPorts,omitempty" protobuf:"bytes,12,opt,name=excludeInboundPorts"`
	// Lists the excluded IP ranges of Istio egress traffic that the sidecar captures.
	ExcludeIPRanges *string `json:"excludeIPRanges,omitempty" protobuf:"bytes,13,opt,name=excludeIPRanges"`
	// Image name or path for the proxy.
	Image *string `json:"image,omitempty" protobuf:"bytes,14,opt,name=image"`
	// Specifies the Istio ingress ports to capture.
	IncludeInboundPorts *string `json:"includeInboundPorts,omitempty" protobuf:"bytes,15,opt,name=includeInboundPorts"`
	// Lists the IP ranges of Istio egress traffic that the sidecar captures.
	IncludeIPRanges    *string `json:"includeIPRanges,omitempty" protobuf:"bytes,16,opt,name=includeIPRanges"`
	KubevirtInterfaces *string `json:"kubevirtInterfaces,omitempty" protobuf:"bytes,17,opt,name=kubevirtInterfaces"`
	LogLevel           *string `json:"logLevel,omitempty" protobuf:"bytes,18,opt,name=logLevel"`
	Privileged         *bool   `json:"privileged,omitempty" protobuf:"varint,19,opt,name=privileged"`
	// Sets the initial delay for readiness probes in seconds.
	ReadinessInitialDelaySeconds *uint32 `json:"readinessInitialDelaySeconds,omitempty" protobuf:"varint,20,opt,name=readinessInitialDelaySeconds"`
	// Sets the interval between readiness probes in seconds.
	ReadinessPeriodSeconds *uint32 `json:"readinessPeriodSeconds,omitempty" protobuf:"varint,21,opt,name=readinessPeriodSeconds"`
	// Sets the number of successive failed probes before indicating readiness failure.
	ReadinessFailureThreshold *uint32 `json:"readinessFailureThreshold,omitempty" protobuf:"varint,22,opt,name=readinessFailureThreshold"`
	// Default port used for the Pilot agent's health checks.
	StatusPort *uint32          `json:"statusPort,omitempty" protobuf:"varint,23,opt,name=statusPort"`
	Resources  *ResourcesConfig `json:"resources,omitempty" protobuf:"bytes,24,opt,name=resources"`
	// Specifies which tracer to use.
	Tracer *string `json:"tracer,omitempty" protobuf:"bytes,25,opt,name=tracer"`
}

// EnvoyMetricsConfig is described in istio.io documentation.
type EnvoyMetricsConfig struct {
	// Enables the Envoy Metrics Service.
	Enabled *bool `json:"enabled,omitempty" protobuf:"varint,1,opt,name=enabled"`
	// Sets the destination Envoy Metrics Service address in Envoy.
	Host *string `json:"host,omitempty" protobuf:"bytes,2,opt,name=host"`
	// Sets the destination Envoy Metrics Service port in Envoy.
	Port *int32 `json:"port,omitempty" protobuf:"varint,3,opt,name=port"`
}

// ProxyInitConfig is described in istio.io documentation.
type ProxyInitConfig struct {
	Image *string `json:"image,omitempty" protobuf:"bytes,1,opt,name=image"`
}

// PilotIngressConfig is described in istio.io documentation.
type PilotIngressConfig struct {
	IngressService        string `json:"ingressService,omitempty" protobuf:"bytes,1,opt,name=ingressService"`
	IngressControllerMode string `json:"ingressControllerMode,omitempty" protobuf:"bytes,2,opt,name=ingressControllerMode"`
	IngressClass          string `json:"ingressClass,omitempty" protobuf:"bytes,3,opt,name=ingressClass"`
}

// PilotPolicyConfig is described in istio.io documentation.
type PilotPolicyConfig struct {
	Enabled *bool `json:"enabled,omitempty" protobuf:"varint,1,opt,name=enabled"`
}

// PilotTelemetryConfig is described in istio.io documentation.
type PilotTelemetryConfig struct {
	Enabled *bool `json:"enabled,omitempty" protobuf:"varint,1,opt,name=enabled"`
}

// SDSConfig is described in istio.io documentation.
type SDSConfig struct {
	Enabled *bool `json:"enabled,omitempty" protobuf:"varint,1,opt,name=enabled"`
	// Specifies the Unix Domain Socket through which Envoy communicates with NodeAgent SDS to get key/cert for mTLS.
	UDSPath *string `json:"udsPath,omitempty" protobuf:"bytes,2,opt,name=udsPath"`
	// Enables SDS use of k8s sa normal JWT to request for certificates.
	UseNormalJWT *bool `json:"useNormalJwt,omitempty" protobuf:"varint,3,opt,name=useNormalJwt"`
	// Enables SDS use of trustworthy JWT to request for certificates.
	UseTrustworthyJWT *bool `json:"useTrustworthyJwt,omitempty" protobuf:"varint,4,opt,name=useTrustworthyJwt"`
}

// TracerConfig is described in istio.io documentation.
type TracerConfig struct {
	Datadog   *TracerDatadogConfig   `json:"datadog,omitempty" protobuf:"bytes,1,opt,name=datadog"`
	LightStep *TracerLightStepConfig `json:"lightstep,omitempty" protobuf:"bytes,2,opt,name=lightstep"`
	Zipkin    *TracerZipkinConfig    `json:"zipkin,omitempty" protobuf:"bytes,3,opt,name=zipkin"`
}

// TracerDatadogConfig is described in istio.io documentation.
type TracerDatadogConfig struct {
	Address *string `json:"address,omitempty" protobuf:"bytes,1,opt,name=address"`
}

// TracerLightStepConfig is described in istio.io documentation.
type TracerLightStepConfig struct {
	// Sets the lightstep satellite pool address.
	Address *string `json:"address,omitempty" protobuf:"bytes,1,opt,name=address"`
	// Sets the lightstep access token.
	AccessToken *string `json:"accessToken,omitempty" protobuf:"bytes,2,opt,name=accessToken"`
	// Sets path to the file containing the cacert to use when verifying TLS.
	CACertPath *string `json:"cacertPath,omitempty" protobuf:"bytes,3,opt,name=cacertPath"`
	// Enables lightstep secure connection.
	Secure *bool `json:"secure,omitempty" protobuf:"varint,4,opt,name=secure"`
}

// TracerZipkinConfig is described in istio.io documentation.
type TracerZipkinConfig struct {
	// Specifies address in host:port format for reporting trace data in zipkin format.
	Address *string `json:"address,omitempty" protobuf:"bytes,1,opt,name=address"`
}

// ResourcesConfig is described in istio.io documentation.
type ResourcesConfig struct {
	Requests *ResourcesRequestsConfig `json:"requests,omitempty" protobuf:"bytes,1,opt,name=requests"`
	Limits   *ResourcesRequestsConfig `json:"limits,omitempty" protobuf:"bytes,2,opt,name=limits"`
}

// ResourcesRequestsConfig is described in istio.io documentation.
type ResourcesRequestsConfig struct {
	CPU    *string `json:"cpu,omitempty" protobuf:"bytes,1,opt,name=cpu"`
	Memory *string `json:"memory,omitempty" protobuf:"bytes,2,opt,name=memory"`
}
