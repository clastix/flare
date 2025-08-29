// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	fluidosreservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	fluidosconsts "github.com/fluidos-project/node/pkg/utils/consts"
	"github.com/fluidos-project/node/pkg/utils/flags"
	nodeoffloadingv1beta1 "github.com/liqotech/liqo/apis/offloading/v1beta1"
	liqoconsts "github.com/liqotech/liqo/pkg/consts"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	flarev1alpha1 "github.com/clastix/flare-internal/api/v1alpha1"
)

func (i *IntentReconciler) TrackNamespaceOffloadingUnknown(ctx context.Context, intent *flarev1alpha1.Intent) error {
	meta.SetStatusCondition(&intent.Status.Conditions, metav1.Condition{
		Type:    flarev1alpha1.IntentStatusTypeOffloading,
		Status:  metav1.ConditionUnknown,
		Reason:  "NamespaceOffloadingCreation",
		Message: "NamespaceOffloading is about to start",
	})

	return i.Client.Status().Update(ctx, intent)
}

//+kubebuilder:rbac:groups=reservation.fluidos.eu,resources=reservations;contracts,verbs=get;list;watch
//+kubebuilder:rbac:groups=offloading.liqo.io,resources=namespaceoffloadings,verbs=get;list;watch;create;update

func (i *IntentReconciler) HandleNamespaceOffloadingPhase(ctx context.Context, intent *flarev1alpha1.Intent) error {
	var nsOffloading nodeoffloadingv1beta1.NamespaceOffloading
	if err := i.Client.Get(ctx, types.NamespacedName{Name: liqoconsts.DefaultNamespaceOffloadingName, Namespace: intent.Namespace}, &nsOffloading); err != nil {
		if apierrors.IsNotFound(err) {
			return i.CreateNamespaceOffloading(ctx, intent)
		}

		return errors.Wrap(err, "Cannot handle NamespaceOffloading")
	}

	condition := meta.FindStatusCondition(intent.Status.Conditions, flarev1alpha1.IntentStatusTypeOffloading).DeepCopy()

	switch nsOffloading.Status.OffloadingPhase {
	case nodeoffloadingv1beta1.ReadyOffloadingPhaseType:
		condition.Status = metav1.ConditionTrue
		condition.Reason = "Completed"
		condition.Message = ""
	case nodeoffloadingv1beta1.NoClusterSelectedOffloadingPhaseType:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "NoClusterSelected"
		condition.Message = ""
	case nodeoffloadingv1beta1.InProgressOffloadingPhaseType:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "InProgress"
		condition.Message = ""
	case nodeoffloadingv1beta1.SomeFailedOffloadingPhaseType:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "SomeFailed"
		condition.Message = ""
	case nodeoffloadingv1beta1.AllFailedOffloadingPhaseType:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "AllFailed"
		condition.Message = ""
	case nodeoffloadingv1beta1.TerminatingOffloadingPhaseType:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "Terminating"
		condition.Message = ""
	}

	condition.ObservedGeneration = nsOffloading.Generation

	return UpdateStatusCondition(ctx, i.Client, intent, *condition)

}

func (i *IntentReconciler) CreateNamespaceOffloading(ctx context.Context, intent *flarev1alpha1.Intent) error {
	condition := meta.FindStatusCondition(intent.Status.Conditions, flarev1alpha1.IntentStatusTypeOffloading).DeepCopy()

	var reservation fluidosreservationv1alpha1.Reservation
	if err := i.Client.Get(ctx, types.NamespacedName{Name: "reservation-" + intent.Namespace, Namespace: flags.FluidosNamespace}, &reservation); err != nil {
		condition.Reason = "ReservationNotFound"
		condition.Message = err.Error()

		return UpdateStatusCondition(ctx, i.Client, intent, *condition)
	}

	if reservation.Status.Contract.Name == "" {
		condition.Reason = "MissingContractReference"
		condition.Message = "Missing Name or Namespace in Contract status"

		return UpdateStatusCondition(ctx, i.Client, intent, *condition)
	}

	var contract fluidosreservationv1alpha1.Contract
	if err := i.Client.Get(ctx, types.NamespacedName{Name: reservation.Status.Contract.Name, Namespace: reservation.Status.Contract.Namespace}, &contract); err != nil {
		condition.Reason = "ContractNotFound"
		condition.Message = err.Error()

		return UpdateStatusCondition(ctx, i.Client, intent, *condition)
	}

	if contract.Spec.PeeringTargetCredentials.ClusterID == "" {
		condition.Reason = "MissingPeeringTargetCredentials"
		condition.Message = "LiqoID field is empty"

		return UpdateStatusCondition(ctx, i.Client, intent, *condition)
	}

	var nsOffloading nodeoffloadingv1beta1.NamespaceOffloading
	nsOffloading.Name = liqoconsts.DefaultNamespaceOffloadingName
	nsOffloading.Namespace = intent.Namespace

	or, err := controllerutil.CreateOrUpdate(ctx, i.Client, &nsOffloading, func() error {
		nsOffloading.Spec.NamespaceMappingStrategy = nodeoffloadingv1beta1.DefaultNameMappingStrategyType
		nsOffloading.Spec.PodOffloadingStrategy = nodeoffloadingv1beta1.RemotePodOffloadingStrategyType
		nsOffloading.Spec.ClusterSelector.NodeSelectorTerms = []corev1.NodeSelectorTerm{
			{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      fluidosconsts.LiqoRemoteClusterIDLabel,
						Operator: corev1.NodeSelectorOpIn,
						Values: []string{
							contract.Spec.PeeringTargetCredentials.ClusterID,
						},
					},
				},
			},
		}

		return nil
	})
	switch {
	case err != nil:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "NamespaceOffloadingCreationFailed"
		condition.Message = err.Error()
	case or == controllerutil.OperationResultCreated, or == controllerutil.OperationResultNone:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "NamespaceOffloadingCreationCompleted"
		condition.Message = "NamespaceOffloading has been created, waiting for its solving."
	case or == controllerutil.OperationResultUpdated:
		condition.Status = metav1.ConditionFalse
		condition.Reason = "NamespaceOffloadingCreationCompleted"
		condition.Message = "NamespaceOffloading has been updated, waiting for its solving."
	}

	condition.ObservedGeneration = nsOffloading.Generation

	return UpdateStatusCondition(ctx, i.Client, intent, *condition)
}
