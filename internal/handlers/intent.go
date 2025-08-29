// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	fluidosv1alpha1 "github.com/fluidos-project/node/apis/advertisement/v1alpha1"
	fluidosnodev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/labstack/echo/v4"
	"github.com/projectcapsule/capsule/pkg/indexer"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	flarev1alpha1 "github.com/clastix/flare-internal/api/v1alpha1"
	"github.com/clastix/flare-internal/internal/api"
)

//+kubebuilder:rbac:groups="",resources=namespaces,verbs=create;get;list;watch;delete
//+kubebuilder:rbac:groups=advertisement.fluidos.eu,resources=peeringcandidates,verbs=get;list;watch

type Intent struct {
	Client           client.Client
	IntentUIDIndexer indexer.CustomIndexer
	Helper           Helper
}

func (i *Intent) ListIntents(ctx echo.Context) error {
	user := ctx.Get("user").(authenticationv1.UserInfo)

	tnt, notFoundErr := i.Helper.RetrieveCapsuleTenant(ctx.Request().Context(), user)
	if notFoundErr != nil {
		if apierrors.IsNotFound(notFoundErr) {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "user is not assigned to any Tenant",
			})
		}

		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   notFoundErr.Error(),
			"context": "cannot retrieve list of Tenants",
		})
	}

	statuses := make([]api.IntentStatus, 0, len(tnt.Status.Namespaces))

	for _, ns := range tnt.Status.Namespaces {
		var intentList flarev1alpha1.IntentList

		if err := i.Client.List(ctx.Request().Context(), &intentList, client.InNamespace(ns)); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error":   err.Error(),
				"context": "cannot retrieve list of Intents",
			})
		}

		for _, intent := range intentList.Items {
			statuses = append(statuses, i.formatIntentToAPI(intent))
		}

	}

	return ctx.JSON(200, api.ListIntentsResponse{
		Tokens: &statuses,
	})
}

func (i *Intent) formatIntentToAPI(intent flarev1alpha1.Intent) api.IntentStatus {
	return api.IntentStatus{
		CurrentCost: func() *string {
			if intent.Spec.Constraints.MaxHourlyCost == float64(0) {
				return nil
			}

			runHours := intent.Spec.Constraints.MaxHourlyCost * time.Now().Sub(intent.CreationTimestamp.Time).Truncate(time.Hour).Hours()
			if runHours == 0 {
				runHours++
			}

			parsedFloat := strconv.FormatFloat(runHours, 'g', -1, 64)

			return ptr.To(parsedFloat + " EUR")
		}(),
		GpuUtilization: nil,
		IntentId:       ptr.To(string(intent.UID)),
		Message: func() *string {
			for _, condition := range intent.Status.Conditions {
				if condition.Status == metav1.ConditionFalse {
					return ptr.To(condition.Message + "(" + condition.Reason + ")")
				}
			}

			return ptr.To("Intent running successfully")
		}(),
		Runtime: ptr.To(time.Now().Sub(intent.CreationTimestamp.Time).Truncate(time.Second).String()),
		Status: func() *string {
			if len(intent.Status.Conditions) == 0 {
				return ptr.To("Pending")
			}

			for _, condition := range intent.Status.Conditions {
				if condition.Status == metav1.ConditionFalse {
					return ptr.To(condition.Type + "NotReady")
				}
			}

			return ptr.To("Ready")
		}(),
		WorkloadUrl: func() *string {
			for _, port := range intent.Spec.Workload.Ports {
				if !port.Expose || port.Domain == "" {
					continue
				}

				return ptr.To("https://" + port.Domain)
			}

			return nil
		}(),
	}
}

func (i *Intent) SubmitIntent(ctx echo.Context) error {
	user := ctx.Get("user").(authenticationv1.UserInfo)

	var body api.IntentSubmission
	if err := ctx.Bind(&body); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	if body.Intent == nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "missing intent key in body",
		})
	}

	tnt, notFoundErr := i.Helper.RetrieveCapsuleTenant(ctx.Request().Context(), user)
	if notFoundErr != nil {
		if apierrors.IsNotFound(notFoundErr) {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "user is not assigned to any Tenant",
			})
		}

		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   notFoundErr.Error(),
			"context": "cannot retrieve list of Tenants",
		})
	}

	var ns corev1.Namespace
	ns.GenerateName = fmt.Sprintf("%s-", tnt.Name)

	if err := controllerutil.SetControllerReference(tnt, &ns, i.Client.Scheme()); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   err.Error(),
			"context": "cannot set OwnerReference for Namespace",
		})
	}

	if err := i.Client.Create(ctx.Request().Context(), &ns); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   err.Error(),
			"context": "cannot create Namespace",
		})
	}

	var intent flarev1alpha1.Intent
	intent.Namespace = ns.Name
	intent.Name = strings.ReplaceAll(ns.Name, ns.GenerateName, "")

	if body.Intent.Constraints != nil {
		if body.Intent.Constraints.Availability != nil {
			if body.Intent.Constraints.Availability.BlackoutDates != nil {
				intent.Spec.Constraints.Availability.BlackoutDates = make([]string, 0, len(*body.Intent.Constraints.Availability.BlackoutDates))
				for _, date := range *body.Intent.Constraints.Availability.BlackoutDates {
					intent.Spec.Constraints.Availability.BlackoutDates = append(intent.Spec.Constraints.Availability.BlackoutDates, date.String())
				}
			}

			intent.Spec.Constraints.Availability.DaysOfWeek = make([]flarev1alpha1.DayOfWeek, 0, len(body.Intent.Constraints.Availability.DaysOfWeek))
			for _, day := range body.Intent.Constraints.Availability.DaysOfWeek {
				intent.Spec.Constraints.Availability.DaysOfWeek = append(intent.Spec.Constraints.Availability.DaysOfWeek, flarev1alpha1.DayOfWeek(day))
			}

			intent.Spec.Constraints.Availability.MaintenanceWindows = make([]flarev1alpha1.IntentConstraintAvailabilityMaintenanceWindow, 0, len(body.Intent.Constraints.Availability.MaintenanceWindows))
			for _, window := range body.Intent.Constraints.Availability.MaintenanceWindows {
				var maintenanceWindow flarev1alpha1.IntentConstraintAvailabilityMaintenanceWindow

				switch {
				case window.Frequency == nil:
					break
				case *window.Frequency == api.Monthly:
					maintenanceWindow.Frequency = "Monthly"
				case *window.Frequency == api.Weekly:
					maintenanceWindow.Frequency = "Weekly"
				default:
					return ctx.JSON(http.StatusBadRequest, map[string]string{
						"error":   "unhandled frequency enum",
						"context": string(*window.Frequency),
					})
				}

				maintenanceWindow.Start = window.Start.String()
				maintenanceWindow.End = window.End.String()

				intent.Spec.Constraints.Availability.MaintenanceWindows = append(intent.Spec.Constraints.Availability.MaintenanceWindows, maintenanceWindow)
			}

			intent.Spec.Constraints.Availability.Timezone = body.Intent.Constraints.Availability.Timezone
			intent.Spec.Constraints.Availability.WindowEnd = body.Intent.Constraints.Availability.WindowEnd
			intent.Spec.Constraints.Availability.WindowStart = body.Intent.Constraints.Availability.WindowStart
		}

		if body.Intent.Constraints.AvailabilityZone != nil {
			intent.Spec.Constraints.AvailabilityZone = *body.Intent.Constraints.AvailabilityZone
		}

		if body.Intent.Constraints.Compliance != nil {
			intent.Spec.Constraints.Compliance.AuditLogging = ptr.Deref(body.Intent.Constraints.Compliance.AuditLogging, false)

			if body.Intent.Constraints.Compliance.Certifications != nil {
				intent.Spec.Constraints.Compliance.Certifications = make([]flarev1alpha1.Certification, 0, len(*body.Intent.Constraints.Compliance.Certifications))
				for _, cert := range *body.Intent.Constraints.Compliance.Certifications {
					switch cert {
					case api.ISO27001, api.SOC2:
						intent.Spec.Constraints.Compliance.Certifications = append(intent.Spec.Constraints.Compliance.Certifications, flarev1alpha1.Certification(cert))
					default:
						return ctx.JSON(http.StatusBadRequest, map[string]string{
							"error":   "unhandled certification enum",
							"context": string(cert),
						})
					}
				}
			}

			if body.Intent.Constraints.Compliance.DataResidency != nil {
				intent.Spec.Constraints.Compliance.DataResidency = make([]string, 0, len(*body.Intent.Constraints.Compliance.DataResidency))
				for _, dataResidency := range *body.Intent.Constraints.Compliance.DataResidency {
					intent.Spec.Constraints.Compliance.DataResidency = append(intent.Spec.Constraints.Compliance.DataResidency, string(dataResidency))
				}
			}

			intent.Spec.Constraints.Compliance.EncryptionAtRest = ptr.Deref(body.Intent.Constraints.Compliance.EncryptionAtRest, false)
			intent.Spec.Constraints.Compliance.EncryptionInTransit = ptr.Deref(body.Intent.Constraints.Compliance.EncryptionInTransit, false)
			intent.Spec.Constraints.Compliance.GDPRCompliant = ptr.Deref(body.Intent.Constraints.Compliance.GdprCompliant, false)
			intent.Spec.Constraints.Compliance.HIPPACompliant = ptr.Deref(body.Intent.Constraints.Compliance.HipaaCompliant, false)
		}

		if body.Intent.Constraints.Deadline != nil {
			intent.Spec.Constraints.Deadline = metav1.Time{Time: *body.Intent.Constraints.Deadline}
		}

		if body.Intent.Constraints.Energy != nil {
			intent.Spec.Constraints.Energy.EnergyEfficiencyRating = ptr.Deref(body.Intent.Constraints.Energy.EnergyEfficiencyRating, "")
			intent.Spec.Constraints.Energy.GreenCertifiedOnly = ptr.Deref(body.Intent.Constraints.Energy.GreenCertifiedOnly, false)
			intent.Spec.Constraints.Energy.MaxCarbonFootprint = ptr.Deref(body.Intent.Constraints.Energy.MaxCarbonFootprint, "")
			intent.Spec.Constraints.Energy.PowerUsageEffectiveness = ptr.Deref(body.Intent.Constraints.Energy.PowerUsageEffectiveness, float32(0))
			intent.Spec.Constraints.Energy.RenewableEnergyOnly = ptr.Deref(body.Intent.Constraints.Energy.RenewableEnergyOnly, false)
		}

		intent.Spec.Constraints.Location = ptr.Deref(body.Intent.Constraints.Location, "")
		if body.Intent.Constraints.MaxHourlyCost != nil {
			maxHourlyCost := strings.ReplaceAll(*body.Intent.Constraints.MaxHourlyCost, "EUR", "")
			maxHourlyCost = strings.TrimSpace(maxHourlyCost)
			intent.Spec.Constraints.MaxHourlyCost, _ = strconv.ParseFloat(maxHourlyCost, 32)
		}

		intent.Spec.Constraints.MaxLatencyMs = int64(ptr.Deref(body.Intent.Constraints.MaxLatencyMs, 0))

		if body.Intent.Constraints.MaxTotalCost != nil {
			maxTotalCost := strings.ReplaceAll(*body.Intent.Constraints.MaxTotalCost, "EUR", "")
			maxTotalCost = strings.TrimSpace(maxTotalCost)
			intent.Spec.Constraints.MaxTotalCost, _ = strconv.ParseFloat(maxTotalCost, 32)
		}

		if body.Intent.Constraints.Negotiation != nil {
			intent.Spec.Constraints.Negotiation.AutoAcceptThreshold = float64(ptr.Deref(body.Intent.Constraints.Negotiation.AutoAcceptThreshold, float32(0)))
			intent.Spec.Constraints.Negotiation.FallbackStrategy = ptr.Deref(body.Intent.Constraints.Negotiation.FallbackStrategy, "")
			intent.Spec.Constraints.Negotiation.MaxNegotiationRounds = ptr.Deref(body.Intent.Constraints.Negotiation.MaxNegotiationRounds, 0)
			intent.Spec.Constraints.Negotiation.PriceFlexibility = float64(ptr.Deref(body.Intent.Constraints.Negotiation.PriceFlexibility, float32(0)))
			intent.Spec.Constraints.Negotiation.ResourceFlexibility = float64(ptr.Deref(body.Intent.Constraints.Negotiation.ResourceFlexibility, float32(0)))
			intent.Spec.Constraints.Negotiation.TimeoutSeconds = int64(ptr.Deref(body.Intent.Constraints.Negotiation.TimeoutSeconds, 0))
		}

		if body.Intent.Constraints.Performance != nil {
			intent.Spec.Constraints.Performance.GpuUtilizationTarget = float64(ptr.Deref(body.Intent.Constraints.Performance.GpuUtilizationTarget, float32(0)))

			if body.Intent.Constraints.Performance.MaxColdStartTime != nil {
				if d, err := time.ParseDuration(*body.Intent.Constraints.Performance.MaxColdStartTime); err != nil {
					return ctx.JSON(http.StatusBadRequest, map[string]string{
						"error":   "cannot parse max cold start time value",
						"context": *body.Intent.Constraints.Performance.MaxColdStartTime,
					})
				} else {
					intent.Spec.Constraints.Performance.MaxColdStartTime = metav1.Duration{Duration: d}
				}
			}

			intent.Spec.Constraints.Performance.MaxJitterMs = int64(ptr.Deref(body.Intent.Constraints.Performance.MaxJitterMs, 0))
			intent.Spec.Constraints.Performance.MemoryUtilizationTarget = float64(ptr.Deref(body.Intent.Constraints.Performance.MemoryUtilizationTarget, float32(0)))

			if body.Intent.Constraints.Performance.MinNetworkBandwidth != nil {
				qty, qErr := resource.ParseQuantity(*body.Intent.Constraints.Performance.MinNetworkBandwidth)
				if qErr != nil {
					return ctx.JSON(http.StatusBadRequest, map[string]string{
						"error":   "cannot parse quantity for minimum network bandwidth",
						"context": *body.Intent.Constraints.Performance.MaxColdStartTime,
					})
				}

				intent.Spec.Constraints.Performance.MinNetworkBandwidth = qty
			}

			intent.Spec.Constraints.Performance.MinUptimePercent = float64(ptr.Deref(body.Intent.Constraints.Performance.MinUptimePercent, float32(0)))
		}

		intent.Spec.Constraints.PreEmptible = ptr.Deref(body.Intent.Constraints.Preemptible, false)

		if body.Intent.Constraints.Providers != nil {
			intent.Spec.Constraints.Providers = make([]string, 0, len(*body.Intent.Constraints.Providers))
			for _, provider := range *body.Intent.Constraints.Providers {
				intent.Spec.Constraints.Providers = append(intent.Spec.Constraints.Providers, string(provider))
			}
		}

		if body.Intent.Constraints.Security != nil {
			intent.Spec.Constraints.Security.BastionHost = ptr.Deref(body.Intent.Constraints.Security.BastionHost, false)

			if body.Intent.Constraints.Security.FirewallRules != nil {
				intent.Spec.Constraints.Security.FirewallRules = make([]flarev1alpha1.IntentWorkloadConstraintSecurityFirewallRule, 0, len(*body.Intent.Constraints.Security.FirewallRules))

				for _, rule := range *body.Intent.Constraints.Security.FirewallRules {
					var r flarev1alpha1.IntentWorkloadConstraintSecurityFirewallRule

					r.Port = int32(rule.Port)
					r.Source = rule.Source
					r.Protocol = rule.Protocol

					switch rule.Action {
					case api.Allow, api.Deny:
						r.Action = string(rule.Action)
					default:
						return ctx.JSON(http.StatusBadRequest, map[string]string{
							"error":   "unhandled action enum",
							"context": string(rule.Action),
						})
					}
					intent.Spec.Constraints.Security.FirewallRules = append(intent.Spec.Constraints.Security.FirewallRules, r)
				}
			}

			intent.Spec.Constraints.Security.IntrusionDetection = ptr.Deref(body.Intent.Constraints.Security.IntrusionDetection, false)

			if body.Intent.Constraints.Security.NetworkIsolation != nil {
				switch *body.Intent.Constraints.Security.NetworkIsolation {
				case api.Private, api.Public:
					intent.Spec.Constraints.Security.NetworkIsolation = string(*body.Intent.Constraints.Security.NetworkIsolation)
				default:
					return ctx.JSON(http.StatusBadRequest, map[string]string{
						"error":   "unhandled network isolation enum",
						"context": string(*body.Intent.Constraints.Security.NetworkIsolation),
					})
				}

			}

			intent.Spec.Constraints.Security.VpnAccess = ptr.Deref(body.Intent.Constraints.Security.VpnAccess, false)
			intent.Spec.Constraints.Security.VulnerabilityScanning = ptr.Deref(body.Intent.Constraints.Security.VulnerabilityScanning, false)
		}
	}

	switch body.Intent.Objective {
	case api.BalancedOptimization:
		intent.Spec.Objective = flarev1alpha1.IntentObjectBalancedOptimization
	case api.CostMinimization:
		intent.Spec.Objective = flarev1alpha1.IntentObjectCostMinimization
	case api.EnergyEfficiency:
		intent.Spec.Objective = flarev1alpha1.IntentObjectEnergyEfficiency
	case api.LatencyMinimization:
		intent.Spec.Objective = flarev1alpha1.IntentObjectLatencyMinimization
	case api.PerformanceMaximization:
		intent.Spec.Objective = flarev1alpha1.IntentObjectPerformanceMaximization
	default:
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "unhandled objective type",
			"context": string(body.Intent.Objective),
		})
	}

	if body.Intent.Sla != nil {
		intent.Spec.SLA.Availability = ptr.Deref(body.Intent.Sla.Availability, "")
		intent.Spec.SLA.BackupStrategy = ptr.Deref(body.Intent.Sla.BackupStrategy, "")

		if body.Intent.Sla.MaxInterruptionTime != nil {
			if d, err := time.ParseDuration(*body.Intent.Sla.MaxInterruptionTime); err != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{
					"error":   "cannot parse max interruption time value",
					"context": *body.Intent.Sla.MaxInterruptionTime,
				})
			} else {
				intent.Spec.SLA.MaxInterruptionTime = &metav1.Duration{Duration: d}
			}
		}
	}

	if body.Intent.Workload.Commands != nil {
		intent.Spec.Workload.Commands = *body.Intent.Workload.Commands
	}

	if body.Intent.Workload.CommunicationPattern != nil {
		switch *body.Intent.Workload.CommunicationPattern {
		case api.AllReduce:
			intent.Spec.Workload.CommunicationPattern = "AllReduce"
		case api.Independent:
			intent.Spec.Workload.CommunicationPattern = "Independent"
		case api.Pipeline:
			intent.Spec.Workload.CommunicationPattern = "Pipeline"
		default:
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error":   "unhandled communication pattern enum",
				"context": string(*body.Intent.Workload.CommunicationPattern),
			})
		}
	}

	if body.Intent.Workload.DeploymentStrategy != nil {
		switch *body.Intent.Workload.DeploymentStrategy {
		case api.Colocated:
			intent.Spec.Workload.DeploymentStrategy = "Colocated"
		case api.Distributed:
			intent.Spec.Workload.DeploymentStrategy = "Distributed"
		case api.Flexibile:
			intent.Spec.Workload.DeploymentStrategy = "Flexible"
		default:
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error":   "unhandled deployment strategy enum",
				"context": string(*body.Intent.Workload.DeploymentStrategy),
			})
		}
	}

	if body.Intent.Workload.Env != nil {
		intent.Spec.Workload.Env = append(intent.Spec.Workload.Env, *body.Intent.Workload.Env...)
	}

	intent.Spec.Workload.Image = body.Intent.Workload.Image
	intent.Spec.Workload.Name = body.Intent.Workload.Name

	switch body.Intent.Workload.Type {
	case api.WorkloadTypeBatch:
		intent.Spec.Workload.Type = flarev1alpha1.IntentWorkloadTypeBatch

		if body.Intent.Workload.Batch != nil {
			switch *body.Intent.Workload.Batch.CompletionPolicy {
			case api.BatchCompletionPolicyAll, api.BatchCompletionPolicyAny:
				intent.Spec.Workload.Batch.CompletionPolicy = string(*body.Intent.Workload.Batch.CompletionPolicy)
			default:
				return ctx.JSON(http.StatusBadRequest, map[string]string{
					"error":   "unhandled completion policy enum",
					"context": string(*body.Intent.Workload.Batch.CompletionPolicy),
				})
			}

			intent.Spec.Workload.Batch.MaxRetries = ptr.Deref(body.Intent.Workload.Batch.ParallelTasks, 3)
			intent.Spec.Workload.Batch.ParallelTasks = ptr.Deref(body.Intent.Workload.Batch.ParallelTasks, 1)

			if body.Intent.Workload.Batch.Timeout != nil {
				if d, dErr := time.ParseDuration(*body.Intent.Workload.Batch.Timeout); dErr != nil {
					return ctx.JSON(http.StatusBadRequest, map[string]string{
						"error":   "cannot parse job timeout value",
						"context": *body.Intent.Workload.Batch.Timeout,
					})
				} else {
					intent.Spec.Workload.Batch.Timeout = metav1.Duration{Duration: d}
				}
			}
		}
	case api.WorkloadTypeService:
		intent.Spec.Workload.Type = flarev1alpha1.IntentWorkloadTypeService

		if body.Intent.Workload.Scaling != nil {
			intent.Spec.Workload.Scaling.AutoScale = ptr.Deref(body.Intent.Workload.Scaling.AutoScale, false)
			intent.Spec.Workload.Scaling.MaxReplicas = ptr.Deref(body.Intent.Workload.Scaling.MaxReplicas, 10)
			intent.Spec.Workload.Scaling.MinReplicas = ptr.Deref(body.Intent.Workload.Scaling.MinReplicas, 1)
			intent.Spec.Workload.Scaling.TargetCpuPercent = ptr.Deref(body.Intent.Workload.Scaling.TargetCpuPercent, 70)
			intent.Spec.Workload.Scaling.TargetGpuPercent = ptr.Deref(body.Intent.Workload.Scaling.TargetGpuPercent, 80)

		}
	default:
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error":   "unhandled workload type",
			"context": string(body.Intent.Workload.Type),
		})
	}

	if body.Intent.Workload.Ports != nil {
		for _, port := range *body.Intent.Workload.Ports {
			intent.Spec.Workload.Ports = append(intent.Spec.Workload.Ports, flarev1alpha1.IntentWorkloadPort{
				Port: int32(port.Port),
				Protocol: func() string {
					if port.Protocol != nil {
						return string(*port.Protocol)
					}

					return ""
				}(),
				Expose: ptr.Deref(port.Expose, false),
				Domain: ptr.Deref(port.Domain, ""),
			})
		}
	}

	if body.Intent.Workload.Resources.Cpu != nil {
		qty, qErr := resource.ParseQuantity(*body.Intent.Workload.Resources.Cpu)
		if qErr != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error":   "cannot parse quantity for CPU",
				"context": *body.Intent.Workload.Resources.Cpu,
			})
		}

		intent.Spec.Workload.Resources.CPU = qty
	}

	if body.Intent.Workload.Resources.Gpu != nil {
		switch {
		case body.Intent.Workload.Resources.Gpu.Architecture == nil, *body.Intent.Workload.Resources.Gpu.Architecture == "any":
			intent.Spec.Workload.Resources.GPU.Architecture = "Any"
		default:
			intent.Spec.Workload.Resources.GPU.Architecture = string(*body.Intent.Workload.Resources.Gpu.Architecture)
		}

		if body.Intent.Workload.Resources.Gpu.ClockSpeedMin != nil {
			qty, qErr := resource.ParseQuantity(*body.Intent.Workload.Resources.Gpu.ClockSpeedMin)
			if qErr != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{
					"error":   "cannot parse quantity for clock speed",
					"context": *body.Intent.Workload.Resources.Gpu.ClockSpeedMin,
				})
			}

			intent.Spec.Workload.Resources.GPU.ClockSpeedMin = qty
		}

		switch {
		case body.Intent.Workload.Resources.Gpu.ComputeCapability == nil, *body.Intent.Workload.Resources.Gpu.ComputeCapability == "any":
			intent.Spec.Workload.Resources.GPU.ComputeCapability = "Any"
		default:
			intent.Spec.Workload.Resources.GPU.ComputeCapability = string(*body.Intent.Workload.Resources.Gpu.ComputeCapability)
		}

		intent.Spec.Workload.Resources.GPU.CoresMax = int64(ptr.Deref(body.Intent.Workload.Resources.Gpu.CoresMax, 0))
		intent.Spec.Workload.Resources.GPU.CoresMin = int64(ptr.Deref(body.Intent.Workload.Resources.Gpu.CoresMin, 0))
		intent.Spec.Workload.Resources.GPU.Count = int64(ptr.Deref(body.Intent.Workload.Resources.Gpu.Count, 0))
		intent.Spec.Workload.Resources.GPU.Dedicated = body.Intent.Workload.Resources.Gpu.Dedicated
		intent.Spec.Workload.Resources.GPU.FP32TFlops = float64(ptr.Deref(body.Intent.Workload.Resources.Gpu.Fp32Tflops, float32(0)))
		intent.Spec.Workload.Resources.GPU.Interconnect = ptr.Deref(body.Intent.Workload.Resources.Gpu.Interconnect, "")
		intent.Spec.Workload.Resources.GPU.Interruptible = body.Intent.Workload.Resources.Gpu.Interruptible
		if body.Intent.Workload.Resources.Gpu.Interruptible != nil {
		}

		if body.Intent.Workload.Resources.Gpu.MemoryMax != nil {
			qty, qErr := resource.ParseQuantity(*body.Intent.Workload.Resources.Gpu.MemoryMax)
			if qErr != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{
					"error":   "cannot parse quantity for max memory",
					"context": *body.Intent.Workload.Resources.Gpu.MemoryMax,
				})
			}

			intent.Spec.Workload.Resources.GPU.MemoryMax = qty
		}

		if body.Intent.Workload.Resources.Gpu.MemoryMin != nil {
			qty, qErr := resource.ParseQuantity(*body.Intent.Workload.Resources.Gpu.MemoryMin)
			if qErr != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{
					"error":   "cannot parse quantity for min memory",
					"context": *body.Intent.Workload.Resources.Gpu.MemoryMin,
				})
			}

			intent.Spec.Workload.Resources.GPU.MemoryMin = qty
		}

		switch {
		case body.Intent.Workload.Resources.Gpu.Model == nil, *body.Intent.Workload.Resources.Gpu.Model == "any":
			intent.Spec.Workload.Resources.GPU.Model = "Any"
		default:
			intent.Spec.Workload.Resources.GPU.Model = string(*body.Intent.Workload.Resources.Gpu.Model)
		}

		intent.Spec.Workload.Resources.GPU.MultiGPUEfficiency = float64(ptr.Deref(body.Intent.Workload.Resources.Gpu.MultiGpuEfficiency, float32(0)))
		intent.Spec.Workload.Resources.GPU.MultiInstance = body.Intent.Workload.Resources.Gpu.MultiInstance
		intent.Spec.Workload.Resources.GPU.Shared = body.Intent.Workload.Resources.Gpu.Shared

		switch {
		case body.Intent.Workload.Resources.Gpu.Tier == nil, *body.Intent.Workload.Resources.Gpu.Tier == "any":
			intent.Spec.Workload.Resources.GPU.Tier = "Any"
		default:
			intent.Spec.Workload.Resources.GPU.Tier = string(*body.Intent.Workload.Resources.Gpu.Tier)
		}

		if body.Intent.Workload.Resources.Gpu.Topology != nil {
			switch *body.Intent.Workload.Resources.Gpu.Topology {
			case api.AllToAll:
				intent.Spec.Workload.Resources.GPU.Topology = ptr.To("AllToAll")
			case api.Mesh:
				intent.Spec.Workload.Resources.GPU.Topology = ptr.To("Mesh")
			case api.Nvswitch:
				intent.Spec.Workload.Resources.GPU.Topology = ptr.To("Nvswitch")
			case api.Ring:
				intent.Spec.Workload.Resources.GPU.Topology = ptr.To("Ring")
			default:
				return ctx.JSON(http.StatusBadRequest, map[string]string{
					"error":   "unhandled GPU topology enum",
					"context": string(*body.Intent.Workload.Resources.Gpu.Topology),
				})
			}
		}
	}

	if body.Intent.Workload.Resources.Memory != nil {
		qty, qErr := resource.ParseQuantity(*body.Intent.Workload.Resources.Memory)
		if qErr != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error":   "cannot parse quantity for memory",
				"context": *body.Intent.Workload.Resources.Memory,
			})
		}

		intent.Spec.Workload.Resources.Memory = qty
	}

	if body.Intent.Workload.Secrets != nil {
		intent.Spec.Workload.Secrets = make([]flarev1alpha1.IntentWorkloadSecret, 0, len(*body.Intent.Workload.Secrets))
		for _, secret := range *body.Intent.Workload.Secrets {
			intent.Spec.Workload.Secrets = append(intent.Spec.Workload.Secrets, flarev1alpha1.IntentWorkloadSecret{
				Name: secret.Name,
				Env:  secret.Env,
			})
		}
	}

	if body.Intent.Workload.Storage != nil && *body.Intent.Workload.Storage.Volumes != nil {
		intent.Spec.Workload.Storage.Volumes = make([]flarev1alpha1.IntentWorkloadStorageVolume, 0, len(*body.Intent.Workload.Storage.Volumes))
		for _, volume := range *body.Intent.Workload.Storage.Volumes {
			vol := flarev1alpha1.IntentWorkloadStorageVolume{
				Name: volume.Name,
				Path: volume.Path,
			}

			if qty, qErr := resource.ParseQuantity(volume.Size); qErr != nil {
				return ctx.JSON(http.StatusBadRequest, map[string]string{
					"error":   "cannot parse quantity for volume size",
					"context": volume.Size,
				})
			} else {
				vol.Size = qty
			}

			switch volume.Type {
			case api.Persistent:
				vol.Type = "Persistent"
			case api.Temporary:
				vol.Type = "Temporary"
			default:
				return ctx.JSON(http.StatusBadRequest, map[string]string{
					"error":   "unhandled volume type enum",
					"context": string(volume.Type),
				})
			}

			if volume.Source != nil {
				vol.Source = flarev1alpha1.IntentWorkloadStorageVolumeSource{
					Credentials: ptr.Deref(volume.Source.Credentials, ""),
					Type:        string(ptr.Deref(volume.Source.Type, "")),
					Uri:         ptr.Deref(volume.Source.Uri, ""),
				}
			}

			intent.Spec.Workload.Storage.Volumes = append(intent.Spec.Workload.Storage.Volumes, vol)
		}
	}

	if err := i.Client.Create(ctx.Request().Context(), &intent); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   err.Error(),
			"context": "cannot create Intent",
		})
	}

	return ctx.JSON(200, api.SubmitIntentResponse{
		EstimatedCost:      nil,
		EstimatedStartTime: nil,
		IntentId:           ptr.To(string(intent.UID)),
		Message:            ptr.To("Intent received and processing"),
		Status:             ptr.To("Pending"),
	})
}

//+kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=solvers,verbs=delete

func (i *Intent) CancelIntent(ctx echo.Context, intentId string) error {
	user := ctx.Get("user").(authenticationv1.UserInfo)

	tnt, notFoundErr := i.Helper.RetrieveCapsuleTenant(ctx.Request().Context(), user)
	if notFoundErr != nil {
		if apierrors.IsNotFound(notFoundErr) {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "user is not assigned to any Tenant",
			})
		}

		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   notFoundErr.Error(),
			"context": "cannot retrieve list of Tenants",
		})
	}

	var intentList flarev1alpha1.IntentList
	if err := i.Client.List(ctx.Request().Context(), &intentList, client.MatchingFields(fields.Set{i.IntentUIDIndexer.Field(): intentId})); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   err.Error(),
			"context": "cannot retrieve list of Intents by UID",
		})
	}

	if len(intentList.Items) == 0 {
		return ctx.JSON(http.StatusNotFound, map[string]string{
			"err": "intent not found",
		})
	}

	namespaceList := sets.New[string](tnt.Status.Namespaces...)
	for _, intent := range intentList.Items {
		if !namespaceList.Has(intent.Namespace) {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"err": "intent not found",
			})
		}

		if err := i.Client.Delete(ctx.Request().Context(), &intent); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error":   err.Error(),
				"context": "cannot delete Intent",
			})
		}

		var ns corev1.Namespace
		ns.Name = intent.Namespace
		if err := i.Client.Delete(ctx.Request().Context(), &ns); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error":   err.Error(),
				"context": "cannot delete Namespace",
			})
		}

		var solver fluidosnodev1alpha1.Solver
		solver.Name = intent.Namespace
		solver.Namespace = flags.FluidosNamespace

		if err := i.Client.Delete(ctx.Request().Context(), &solver); err != nil && !apierrors.IsNotFound(err) {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error":   err.Error(),
				"context": "cannot delete Solver",
			})
		}
	}

	return ctx.JSON(http.StatusAccepted, nil)
}

func (i *Intent) GetIntentStatus(ctx echo.Context, intentId string) error {
	user := ctx.Get("user").(authenticationv1.UserInfo)

	tnt, notFoundErr := i.Helper.RetrieveCapsuleTenant(ctx.Request().Context(), user)
	if notFoundErr != nil {
		if apierrors.IsNotFound(notFoundErr) {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "user is not assigned to any Tenant",
			})
		}

		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   notFoundErr.Error(),
			"context": "cannot retrieve list of Tenants",
		})
	}

	var intentList flarev1alpha1.IntentList
	if err := i.Client.List(ctx.Request().Context(), &intentList, client.MatchingFields(fields.Set{i.IntentUIDIndexer.Field(): intentId})); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   err.Error(),
			"context": "cannot retrieve list of Intents by UID",
		})
	}

	if len(intentList.Items) == 0 {
		return ctx.JSON(http.StatusNotFound, map[string]string{
			"err": "intent not found",
		})
	}

	namespaceList := sets.New[string](tnt.Status.Namespaces...)
	for _, intent := range intentList.Items {
		if !namespaceList.Has(intent.Namespace) {
			return ctx.JSON(http.StatusNotFound, map[string]string{
				"err": "intent not found",
			})
		}

		return ctx.JSON(200, i.formatIntentToAPI(intent))
	}

	return ctx.JSON(http.StatusInternalServerError, map[string]string{
		"err": "Intent is expected to be found",
	})
}

func (i *Intent) GetAvailableResources(ctx echo.Context) error {
	var peeringCandidates fluidosv1alpha1.PeeringCandidateList

	if err := i.Client.List(ctx.Request().Context(), &peeringCandidates); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   err.Error(),
			"context": "cannot retrieve PeeringCandidates",
		})
	}

	availableGPUs := make([]api.AvailableGPU, 0, len(peeringCandidates.Items))

	for _, pc := range peeringCandidates.Items {
		var obj map[string]interface{}
		if err := json.Unmarshal(pc.Spec.Flavor.Spec.FlavorType.TypeData.Raw, &obj); err != nil {
			continue
		}

		costPerHour, _, _ := unstructured.NestedFloat64(obj, "characteristics", "gpu", "hourly_rate")
		count, _, _ := unstructured.NestedInt64(obj, "characteristics", "gpu", "count")
		memory, _, _ := unstructured.NestedString(obj, "characteristics", "gpu", "memory")
		model, _, _ := unstructured.NestedString(obj, "characteristics", "gpu", "model")
		provider, _, _ := unstructured.NestedString(obj, "characteristics", "gpu", "provider")
		region, _, _ := unstructured.NestedString(obj, "characteristics", "gpu", "region")
		availableGPUs = append(availableGPUs, api.AvailableGPU{
			CostPerHour: ptr.To(strconv.FormatFloat(costPerHour, 'g', -1, 64) + " EUR"),
			Count:       ptr.To(int(count)),
			Location:    &region,
			Memory:      &memory,
			Model:       &model,
			Provider:    &provider,
		})
	}

	return ctx.JSON(200, api.AvailableResourcesResponse{
		AvailableGpus: &availableGPUs,
	})
}
