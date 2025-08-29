// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	fluidosnodev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	flarev1alpha1 "github.com/clastix/flare-internal/api/v1alpha1"
)

//+kubebuilder:rbac:groups=nodecore.fluidos.eu,resources=solvers,verbs=get;list;watch;create

func (i *IntentReconciler) HandleSolverPhase(ctx context.Context, intent *flarev1alpha1.Intent) error {
	var solver fluidosnodev1alpha1.Solver
	if err := i.Client.Get(ctx, types.NamespacedName{Name: intent.Namespace, Namespace: flags.FluidosNamespace}, &solver); err != nil {
		if apierrors.IsNotFound(err) {
			return i.CreateSolver(ctx, intent)
		}

		return errors.Wrap(err, "cannot handle Solver")
	}

	condition := meta.FindStatusCondition(intent.Status.Conditions, flarev1alpha1.IntentStatusTypeSolver).DeepCopy()

	switch solver.Status.SolverPhase.Phase {
	case fluidosnodev1alpha1.PhaseSolved:
		condition.Status = metav1.ConditionTrue
		condition.Reason = "SolverSolved"
	case fluidosnodev1alpha1.PhaseFailed:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "SolverFailed"
	case fluidosnodev1alpha1.PhaseRunning:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "SolverRunning"
	case fluidosnodev1alpha1.PhaseAllocating:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "SolverAllocating"
	case fluidosnodev1alpha1.PhaseIdle:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "SolverIdle"
	case fluidosnodev1alpha1.PhaseTimeout:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "SolverTimeout"
	case fluidosnodev1alpha1.PhaseActive:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "SolverActive"
	case fluidosnodev1alpha1.PhasePending:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "SolverPending"
	case fluidosnodev1alpha1.PhaseInactive:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "SolverInactive"
	}

	condition.ObservedGeneration = solver.Generation
	condition.Message = solver.Status.SolverPhase.Message

	return UpdateStatusCondition(ctx, i.Client, intent, *condition)
}

func (i *IntentReconciler) CreateSolver(ctx context.Context, intent *flarev1alpha1.Intent) error {
	condition := meta.FindStatusCondition(intent.Status.Conditions, flarev1alpha1.IntentStatusTypeSolver).DeepCopy()

	var solver fluidosnodev1alpha1.Solver
	solver.Name = intent.Namespace
	solver.Namespace = flags.FluidosNamespace

	res, err := controllerutil.CreateOrUpdate(ctx, i.Client, &solver, func() error {
		solver.Spec.IntentID = intent.Namespace
		solver.Spec.FindCandidate = true
		solver.Spec.ReserveAndBuy = true
		solver.Spec.EstablishPeering = true

		sliceSelector := fluidosnodev1alpha1.K8SliceSelector{
			GPUFilters: make([]fluidosnodev1alpha1.GPUFieldSelector, 0),
		}

		if intent.Spec.Workload.Resources.GPU.MultiGPUEfficiency > 0 {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "multi_gpu_efficiency",
				Selector: fluidosnodev1alpha1.ResourceRangeSelectorName,
				Data:     runtime.RawExtension{},
			}

			selector.Data.Raw, _ = json.Marshal(fluidosnodev1alpha1.NumberRangeSelector{
				Min: &intent.Spec.Workload.Resources.GPU.MultiGPUEfficiency,
			})

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if intent.Spec.Workload.Resources.GPU.Topology != nil && *intent.Spec.Workload.Resources.GPU.Topology != "" {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "topology",
				Selector: fluidosnodev1alpha1.ResourceMatchSelectorName,
				Data:     runtime.RawExtension{},
			}

			selector.Data.Raw, _ = json.Marshal(fluidosnodev1alpha1.StringMatchSelector{
				Value: *intent.Spec.Workload.Resources.GPU.Topology,
			})

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if intent.Spec.Workload.Resources.GPU.FP32TFlops > 0 {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "fp32_tflops",
				Selector: fluidosnodev1alpha1.ResourceRangeSelectorName,
				Data:     runtime.RawExtension{},
			}

			selector.Data.Raw, _ = json.Marshal(fluidosnodev1alpha1.NumberRangeSelector{
				Min: &intent.Spec.Workload.Resources.GPU.FP32TFlops,
			})

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if intent.Spec.Workload.Resources.GPU.Dedicated != nil {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "dedicated",
				Selector: fluidosnodev1alpha1.BooleanFilterSelectorName,
				Data:     runtime.RawExtension{},
			}

			selector.Data.Raw, _ = json.Marshal(*intent.Spec.Workload.Resources.GPU.Dedicated)

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if intent.Spec.Workload.Resources.GPU.MultiInstance != nil {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "multi_instance",
				Selector: fluidosnodev1alpha1.BooleanFilterSelectorName,
				Data:     runtime.RawExtension{},
			}

			selector.Data.Raw, _ = json.Marshal(*intent.Spec.Workload.Resources.GPU.MultiInstance)

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if intent.Spec.Workload.Resources.GPU.Interruptible != nil {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "interruptible",
				Selector: fluidosnodev1alpha1.BooleanFilterSelectorName,
				Data:     runtime.RawExtension{},
			}

			selector.Data.Raw, _ = json.Marshal(*intent.Spec.Workload.Resources.GPU.Interruptible)

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if intent.Spec.Workload.Resources.GPU.Interconnect != "" {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "interconnect",
				Selector: fluidosnodev1alpha1.ResourceMatchSelectorName,
				Data:     runtime.RawExtension{},
			}

			selector.Data.Raw, _ = json.Marshal(fluidosnodev1alpha1.StringMatchSelector{
				Value: intent.Spec.Workload.Resources.GPU.Interconnect,
			})

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if intent.Spec.Workload.Resources.GPU.Shared != nil {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "shared",
				Selector: fluidosnodev1alpha1.BooleanFilterSelectorName,
				Data:     runtime.RawExtension{},
			}

			selector.Data.Raw, _ = json.Marshal(*intent.Spec.Workload.Resources.GPU.Shared)

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if intent.Spec.Workload.Resources.GPU.Tier != "Any" {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "tier",
				Selector: fluidosnodev1alpha1.ResourceMatchSelectorName,
				Data:     runtime.RawExtension{},
			}

			selector.Data.Raw, _ = json.Marshal(fluidosnodev1alpha1.StringMatchSelector{
				Value: intent.Spec.Workload.Resources.GPU.Tier,
			})

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if intent.Spec.Workload.Resources.GPU.Architecture != "Any" {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "architecture",
				Selector: fluidosnodev1alpha1.ResourceMatchSelectorName,
				Data:     runtime.RawExtension{},
			}

			selector.Data.Raw, _ = json.Marshal(fluidosnodev1alpha1.StringMatchSelector{
				Value: intent.Spec.Workload.Resources.GPU.Architecture,
			})

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if intent.Spec.Workload.Resources.GPU.ComputeCapability != "Any" {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "compute_capability",
				Selector: fluidosnodev1alpha1.ResourceMatchSelectorName,
				Data:     runtime.RawExtension{},
			}

			selector.Data.Raw, _ = json.Marshal(fluidosnodev1alpha1.StringMatchSelector{
				Value: intent.Spec.Workload.Resources.GPU.ComputeCapability,
			})

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if !intent.Spec.Workload.Resources.GPU.ClockSpeedMin.IsZero() {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "clock_speed",
				Selector: fluidosnodev1alpha1.ResourceRangeSelectorName,
				Data:     runtime.RawExtension{},
			}

			selector.Data.Raw, _ = json.Marshal(fluidosnodev1alpha1.ResourceRangeSelector{
				Min: &intent.Spec.Workload.Resources.GPU.ClockSpeedMin,
			})

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if intent.Spec.Workload.Resources.GPU.CoresMin > 0 || intent.Spec.Workload.Resources.GPU.CoresMax > 0 {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "cores",
				Selector: fluidosnodev1alpha1.ResourceRangeSelectorName,
				Data:     runtime.RawExtension{},
			}

			var qtySelector fluidosnodev1alpha1.NumberRangeSelector
			if intent.Spec.Workload.Resources.GPU.CoresMin > 0 {
				qtySelector.Min = ptr.To(float64(intent.Spec.Workload.Resources.GPU.CoresMin))
			}

			if intent.Spec.Workload.Resources.GPU.CoresMax > 0 {
				qtySelector.Max = ptr.To(float64(intent.Spec.Workload.Resources.GPU.CoresMax))
			}

			selector.Data.Raw, _ = json.Marshal(qtySelector)

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if !intent.Spec.Workload.Resources.GPU.MemoryMin.IsZero() || !intent.Spec.Workload.Resources.GPU.MemoryMax.IsZero() {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "memory",
				Selector: fluidosnodev1alpha1.ResourceRangeSelectorName,
				Data:     runtime.RawExtension{},
			}

			var qtySelector fluidosnodev1alpha1.ResourceRangeSelector
			if !intent.Spec.Workload.Resources.GPU.MemoryMin.IsZero() {
				qtySelector.Min = &intent.Spec.Workload.Resources.GPU.MemoryMin
			}

			if !intent.Spec.Workload.Resources.GPU.MemoryMax.IsZero() {
				qtySelector.Max = &intent.Spec.Workload.Resources.GPU.MemoryMax
			}

			selector.Data.Raw, _ = json.Marshal(qtySelector)

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if intent.Spec.Workload.Resources.GPU.Count > 0 {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "count",
				Selector: fluidosnodev1alpha1.NumberMatchSelectorName,
				Data:     runtime.RawExtension{},
			}

			selector.Data.Raw, _ = json.Marshal(fluidosnodev1alpha1.NumberMatchSelector{
				Value: float64(intent.Spec.Workload.Resources.GPU.Count),
			})

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if intent.Spec.Workload.Resources.GPU.Model != "" {
			selector := fluidosnodev1alpha1.GPUFieldSelector{
				Field:    "model",
				Selector: fluidosnodev1alpha1.StringFilterSelectorName,
				Data:     runtime.RawExtension{},
			}

			selector.Data.Raw, _ = json.Marshal(fluidosnodev1alpha1.StringMatchSelector{
				Value: intent.Spec.Workload.Resources.GPU.Model,
			})

			sliceSelector.GPUFilters = append(sliceSelector.GPUFilters, selector)
		}

		if !intent.Spec.Workload.Resources.Memory.IsZero() {
			sliceSelector.MemoryFilter = &fluidosnodev1alpha1.ResourceQuantityFilter{
				Name: fluidosnodev1alpha1.TypeRangeFilter,
				Data: runtime.RawExtension{},
			}

			sliceSelector.MemoryFilter.Data.Raw, _ = json.Marshal(fluidosnodev1alpha1.ResourceRangeSelector{
				Min: &intent.Spec.Workload.Resources.Memory,
			})
		}

		if !intent.Spec.Workload.Resources.CPU.IsZero() {
			sliceSelector.CPUFilter = &fluidosnodev1alpha1.ResourceQuantityFilter{
				Name: fluidosnodev1alpha1.TypeRangeFilter,
				Data: runtime.RawExtension{},
			}

			sliceSelector.CPUFilter.Data.Raw, _ = json.Marshal(fluidosnodev1alpha1.ResourceRangeSelector{
				Min: &intent.Spec.Workload.Resources.CPU,
			})
		}

		solver.Spec.Selector = &fluidosnodev1alpha1.Selector{
			FlavorType: fluidosnodev1alpha1.TypeK8Slice,
			Filters:    &runtime.RawExtension{},
		}

		solver.Spec.Selector.Filters.Raw, _ = json.Marshal(sliceSelector)

		return nil
	})
	switch {
	case err != nil:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "SolverCreationFailed"
		condition.Message = err.Error()
	case res == controllerutil.OperationResultCreated, res == controllerutil.OperationResultNone:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "SolverCreationCompleted"
		condition.Message = "Solver has been created, waiting for its solving."
	case res == controllerutil.OperationResultUpdated:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "SolverCreationCompleted"
		condition.Message = "Solver has been updated, waiting for its solving."
	}

	condition.ObservedGeneration = solver.Generation

	return UpdateStatusCondition(ctx, i.Client, intent, *condition)
}

func (i *IntentReconciler) TrackSolverUnknown(ctx context.Context, intent *flarev1alpha1.Intent) error {
	meta.SetStatusCondition(&intent.Status.Conditions, metav1.Condition{
		Type:    flarev1alpha1.IntentStatusTypeSolver,
		Status:  metav1.ConditionUnknown,
		Reason:  "IntentCreated",
		Message: "Intent is going to be created",
	})

	return i.Client.Status().Update(ctx, intent)
}
