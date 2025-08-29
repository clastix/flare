// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time since creation"

// Intent is the Schema for the intents API.
type Intent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntentSpec   `json:"spec,omitempty"`
	Status IntentStatus `json:"status,omitempty"`
}

var (
	IntentStatusTypeSolver     = "Solver"
	IntentStatusTypeOffloading = "NamespaceOffloading"
	IntentStatusTypeDeploy     = "Deploy"
)

type IntentStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type IntentObject string

var (
	IntentObjectBalancedOptimization    = IntentObject("BalancedOptimization")
	IntentObjectPerformanceMaximization = IntentObject("PerformanceMaximization")
	IntentObjectCostMinimization        = IntentObject("CostMinimization")
	IntentObjectLatencyMinimization     = IntentObject("LatencyMinimization")
	IntentObjectEnergyEfficiency        = IntentObject("EnergyEfficiency")
)

type IntentWorkloadType string

var (
	IntentWorkloadTypeService = IntentWorkloadType("Service")
	IntentWorkloadTypeBatch   = IntentWorkloadType("Batch")
)

type IntentWorkloadPort struct {
	Port int32 `json:"port"`
	//+kubebuilder:default=TCP
	//+kubebuilder:validation:Enum=TCP;UDP
	Protocol string `json:"protocol,omitempty"`
	Expose   bool   `json:"expose,omitempty"`
	Domain   string `json:"domain,omitempty"`
}

type IntentWorkloadResourceGPU struct {
	//+kubebuilder:default=Any
	Model string `json:"model,omitempty"`
	//+kubebuilder:default=1
	Count         int64             `json:"count"`
	MemoryMin     resource.Quantity `json:"memoryMin,omitempty"`
	MemoryMax     resource.Quantity `json:"memoryMax,omitempty"`
	CoresMin      int64             `json:"coresMin,omitempty"`
	CoresMax      int64             `json:"coresMax,omitempty"`
	ClockSpeedMin resource.Quantity `json:"clockSpeedMin,omitempty"`
	//+kubebuilder:default=Any
	ComputeCapability string `json:"computeCapability,omitempty"`
	//+kubebuilder:default=Any
	Architecture string `json:"architecture,omitempty"`
	//+kubebuilder:default=Any
	Tier          string  `json:"tier,omitempty"`
	Shared        *bool   `json:"shared,omitempty"`
	Interconnect  string  `json:"interconnect,omitempty"`
	Interruptible *bool   `json:"interruptible,omitempty"`
	MultiInstance *bool   `json:"multiInstance,omitempty"`
	Dedicated     *bool   `json:"dedicated,omitempty"`
	FP32TFlops    float64 `json:"fp32TFlops,omitempty"`
	//+kubebuilder:validation:Enum=AllToAll;NvSwitch;Ring;Mesh
	Topology           *string `json:"topology,omitempty"`
	MultiGPUEfficiency float64 `json:"multiGPUEfficiency,omitempty"`
}

type IntentWorkloadResource struct {
	CPU    resource.Quantity         `json:"cpu"`
	Memory resource.Quantity         `json:"memory"`
	GPU    IntentWorkloadResourceGPU `json:"gpu"`
}

type IntentConstraintAvailabilityMaintenanceWindow struct {
	//+kubebuilder:validation:Pattern=`^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$`
	Start string `json:"start"`
	//+kubebuilder:validation:Pattern=`^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$`
	End string `json:"end"`
	//+kubebuilder:default=Weekly
	//+kubebuilder:validation:Enum=Weekly;Monthly
	Frequency string `json:"frequency"`
}

// +kubebuilder:validation:Enum=Mon;Tue;Wed;Thu;Fri;Sat;Sun
type DayOfWeek string

type IntentWorkloadConstraintAvailability struct {
	//+kubebuilder:validation:Pattern=`^(?:[01]\d|2[0-3]):[0-5]\d$`
	WindowStart string `json:"windowStart,omitempty"`
	//+kubebuilder:validation:Pattern=`^(?:[01]\d|2[0-3]):[0-5]\d$`
	WindowEnd  string      `json:"windowEnd,omitempty"`
	Timezone   string      `json:"timezone,omitempty"`
	DaysOfWeek []DayOfWeek `json:"daysOfWeek,omitempty"`
	//+kubebuilder:validation:items:Pattern=`^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$`
	BlackoutDates      []string                                        `json:"blackoutDates,omitempty"`
	MaintenanceWindows []IntentConstraintAvailabilityMaintenanceWindow `json:"maintenanceWindows,omitempty"`
}

type IntentWorkloadConstraintNegotiation struct {
	//+kubebuilder:default=3
	MaxNegotiationRounds int `json:"maxNegotiationRounds,omitempty"`
	//+kubebuilder:default=0.15
	PriceFlexibility float64 `json:"priceFlexibility,omitempty"`
	//+kubebuilder:default=0.3
	ResourceFlexibility float64 `json:"resourceFlexibility,omitempty"`
	//+kubebuilder:default=300
	TimeoutSeconds int64 `json:"timeoutSeconds,omitempty"`
	//+kubebuilder:default=queue
	FallbackStrategy string `json:"fallbackStrategy,omitempty"`
	//+kubebuilder:default=0.05
	AutoAcceptThreshold float64 `json:"autoAcceptThreshold,omitempty"`
}

type IntentWorkloadConstraintEnergy struct {
	MaxCarbonFootprint     string `json:"maxCarbonFootprint,omitempty"`
	RenewableEnergyOnly    bool   `json:"renewableEnergyOnly,omitempty"`
	EnergyEfficiencyRating string `json:"energyEfficiencyRating,omitempty"`
	//+kubebuilder:default=2.0
	PowerUsageEffectiveness float32 `json:"powerUsageEffectiveness,omitempty"`
	GreenCertifiedOnly      bool    `json:"greenCertifiedOnly,omitempty"`
}

// +kubebuilder:validation:Enum=ISO27001;SOC2
type Certification string

type IntentWorkloadConstraintCompliance struct {
	DataResidency       []string        `json:"dataResidency,omitempty"`
	Certifications      []Certification `json:"certifications,omitempty"`
	EncryptionAtRest    bool            `json:"encryptionAtRest,omitempty"`
	EncryptionInTransit bool            `json:"encryptionInTransit,omitempty"`
	AuditLogging        bool            `json:"auditLogging,omitempty"`
	GDPRCompliant       bool            `json:"gdprCompliant,omitempty"`
	HIPPACompliant      bool            `json:"hippaCompliant,omitempty"`
}

type IntentWorkloadConstraintPerformance struct {
	MinNetworkBandwidth resource.Quantity `json:"minNetworkBandwidth,omitempty"`
	//+kubebuilder:default=50
	MaxJitterMs int64 `json:"maxJitterMs,omitempty"`
	//+kubebuilder:default=99.0
	MinUptimePercent float64         `json:"minUptimePercent,omitempty"`
	MaxColdStartTime metav1.Duration `json:"maxColdStartTime,omitempty"`
	//+kubebuilder:default=0.80
	GpuUtilizationTarget float64 `json:"gpuUtilizationTarget,omitempty"`
	//+kubebuilder:default=0.80
	MemoryUtilizationTarget float64 `json:"memoryUtilizationTarget,omitempty"`
}

type IntentWorkloadConstraintSecurityFirewallRule struct {
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
	Source   string `json:"source"`
	//+kubebuilder:validation:Enum=Allow;Deny
	Action string `json:"action"`
}

type IntentWorkloadConstraintSecurity struct {
	//+kubebuilder:default:="Public"
	//+kubebuilder:validation:Enum=Public;Private
	NetworkIsolation      string                                         `json:"networkIsolation,omitempty"`
	FirewallRules         []IntentWorkloadConstraintSecurityFirewallRule `json:"firewallRules,omitempty"`
	VpnAccess             bool                                           `json:"vpnAccess,omitempty"`
	BastionHost           bool                                           `json:"bastionHost,omitempty"`
	IntrusionDetection    bool                                           `json:"intrusionDetection,omitempty"`
	VulnerabilityScanning bool                                           `json:"vulnerabilityScanning,omitempty"`
}

type IntentConstraint struct {
	MaxHourlyCost    float64 `json:"maxHourlyCost,omitempty"`
	MaxTotalCost     float64 `json:"maxTotalCost,omitempty"` //TODO(prometherion): advanced
	Location         string  `json:"location,omitempty"`     //TODO(prometherion): advanced
	AvailabilityZone string  `json:"availabilityZone,omitempty"`
	//+kubebuilder:default=100
	MaxLatencyMs int64                                `json:"maxLatencyMs,omitempty"` //TODO(prometherion): advanced
	Deadline     metav1.Time                          `json:"deadline,omitempty"`     //TODO(prometherion): advanced
	PreEmptible  bool                                 `json:"preEmptible,omitempty"`
	Providers    []string                             `json:"providers,omitempty"`
	Availability IntentWorkloadConstraintAvailability `json:"availability,omitempty"` //TODO(prometherion): advanced
	Negotiation  IntentWorkloadConstraintNegotiation  `json:"negotiation,omitempty"`  //TODO(prometherion): advanced
	Energy       IntentWorkloadConstraintEnergy       `json:"energy,omitempty"`       //TODO(prometherion): advanced
	Compliance   IntentWorkloadConstraintCompliance   `json:"compliance,omitempty"`   //TODO(prometherion): advanced
	Performance  IntentWorkloadConstraintPerformance  `json:"performance,omitempty"`  //TODO(prometherion): advanced
	Security     IntentWorkloadConstraintSecurity     `json:"security,omitempty"`     //TODO(prometherion): advanced
}

type IntentWorkloadBatch struct {
	//+kubebuilder:validation:Enum=All;Any
	CompletionPolicy string `json:"completionPolicy,omitempty"`
	//+kubebuilder:default=3
	MaxRetries int `json:"maxRetries,omitempty"`
	//+kubebuilder:default=1
	//+kubebuilder:validation:Minimum=1
	ParallelTasks int `json:"parallelTasks,omitempty"`
	//+kubebuilder:default="1h"
	Timeout metav1.Duration `json:"timeout,omitempty"`
}

type IntentWorkloadScaling struct {
	AutoScale bool `json:"autoScale,omitempty"`
	//+kubebuilder:default=10
	MaxReplicas int `json:"maxReplicas,omitempty"`
	//+kubebuilder:default=1
	MinReplicas int `json:"minReplicas,omitempty"`
	//+kubebuilder:default=70
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=100
	TargetCpuPercent int `json:"targetCPUPercent,omitempty"`
	//+kubebuilder:default=80
	//+kubebuilder:validation:Minimum=0
	//+kubebuilder:validation:Maximum=100
	TargetGpuPercent int `json:"targetGPUPercent,omitempty"`
}

type IntentWorkloadSecret struct {
	// Name is the Secret name.
	Name string `json:"name"`
	// Env is the environment variable the Secret will be injected in the workload.
	Env string `json:"env"`
}

type IntentWorkloadStorageVolumeSource struct {
	Credentials string `json:"credentials,omitempty"`
	//+kubebuilder:validation:Enum=S3;GCS;Azure
	Type string `json:"type,omitempty"`
	Uri  string `json:"uri,omitempty"`
}

type IntentWorkloadStorageVolume struct {
	Name   string                            `json:"name"`
	Path   string                            `json:"path"`
	Size   resource.Quantity                 `json:"size"`
	Source IntentWorkloadStorageVolumeSource `json:"source,omitempty"`
	//+kubebuilder:validation:Enum=Persistent;Temporary
	Type string `json:"type"`
}

type IntentWorkloadStorage struct {
	Volumes []IntentWorkloadStorageVolume `json:"volumes,omitempty"`
}

type IntentWorkload struct {
	//+kubebuilder:validation:Enum=Service;Batch
	Type IntentWorkloadType `json:"type"`
	//+kubebuilder:validation:Enum=AllReduce;Independent;Pipeline
	CommunicationPattern string `json:"communicationPattern,omitempty"` //TODO(prometherion): advanced
	//+kubebuilder:validation:Enum=Colocated;Distributed;Flexible
	DeploymentStrategy string              `json:"deploymentStrategy,omitempty"` //TODO(prometherion): advanced
	Batch              IntentWorkloadBatch `json:"batch,omitempty"`
	Name               string              `json:"name"`
	Image              string              `json:"image"`
	Commands           []string            `json:"commands,omitempty"`
	//+kubebuilder:validation:items:Pattern=`^[A-Z_][A-Z0-9_]*=[^\s]+$`
	Env       []string               `json:"env,omitempty"`
	Secrets   []IntentWorkloadSecret `json:"secrets,omitempty"`
	Storage   IntentWorkloadStorage  `json:"storage,omitempty"`
	Ports     []IntentWorkloadPort   `json:"ports,omitempty"`
	Scaling   IntentWorkloadScaling  `json:"scaling,omitempty"` //TODO(prometherion): advanced
	Resources IntentWorkloadResource `json:"resources,omitempty"`
}

type IntentSLA struct {
	Availability        string           `json:"availability,omitempty"`
	BackupStrategy      string           `json:"backupStrategy,omitempty"`
	MaxInterruptionTime *metav1.Duration `json:"maxInterruptionTime,omitempty"`
}

type IntentSpec struct {
	Constraints IntentConstraint `json:"contraints,omitempty"`
	Objective   IntentObject     `json:"objective"`     //TODO(prometherion): advanced
	SLA         IntentSLA        `json:"sla,omitempty"` //TODO(prometherion): advanced
	Workload    IntentWorkload   `json:"workload"`
}

//+kubebuilder:object:root=true

// IntentList contains a list of Intent instances.
type IntentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Intent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Intent{}, &IntentList{})
}
