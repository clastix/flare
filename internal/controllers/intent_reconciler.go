// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"strings"

	fluidosnodev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/liqotech/liqo/apis/offloading/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	flarev1alpha1 "github.com/clastix/flare-internal/api/v1alpha1"
)

type IntentReconciler struct {
	Client client.Client
}

//+kubebuilder:rbac:groups=flare.clastix.io,resources=intents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=flare.clastix.io,resources=intents/status,verbs=get;update

func (i *IntentReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx)

	var intent flarev1alpha1.Intent
	if err := i.Client.Get(ctx, request.NamespacedName, &intent); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("flarev1alpha.Intent may have been deleted")

			return reconcile.Result{}, nil
		}

		logger.Error(err, "cannot retrieve flarev1alpha.Intent")

		return reconcile.Result{}, err
	}

	if intent.DeletionTimestamp != nil {
		logger.Info("skipping reconciliation for object marked for deletion")

		return reconcile.Result{}, nil
	}

	logger.Info("handling Solver phase")

	solverCondition := meta.FindStatusCondition(intent.Status.Conditions, flarev1alpha1.IntentStatusTypeSolver)
	switch {
	case solverCondition == nil:
		logger.Info("Solver phase is unknown")

		if err := i.TrackSolverUnknown(ctx, &intent); err != nil {
			logger.Error(err, "cannot handle Intent unknown status")

			return reconcile.Result{}, err
		}

		logger.Info("Solver unknown phase has been completed")

		return reconcile.Result{}, nil
	case solverCondition.Status == metav1.ConditionUnknown:
		logger.Info("creating Solver")

		if err := i.CreateSolver(ctx, &intent); err != nil {
			logger.Error(err, "cannot handle Intent creation status")

			return reconcile.Result{}, err
		}

		logger.Info("Solver creation has been completed")

		return reconcile.Result{}, nil
	case solverCondition.Status == metav1.ConditionFalse:
		logger.Info("handling Solver")

		if err := i.HandleSolverPhase(ctx, &intent); err != nil {
			logger.Error(err, "cannot handle Solver phase")

			return reconcile.Result{}, err
		}

		logger.Info("Solver handling has been completed")

		return reconcile.Result{}, nil
	case solverCondition.Status == metav1.ConditionTrue:
		logger.Info("reconciling Solver")

		if err := i.HandleSolverPhase(ctx, &intent); err != nil {
			logger.Error(err, "cannot handle Solver phase")

			return reconcile.Result{}, err
		}

		logger.Info("Solver reconciliation has been completed")
	}

	logger.Info("handling NamespaceOffloading phase")

	nsCondition := meta.FindStatusCondition(intent.Status.Conditions, flarev1alpha1.IntentStatusTypeOffloading)
	switch {
	case nsCondition == nil:
		logger.Info("NamespaceOffloading phase is unknown")

		if err := i.TrackNamespaceOffloadingUnknown(ctx, &intent); err != nil {
			logger.Error(err, "cannot handle NamespaceOffloading unknown status")

			return reconcile.Result{}, err
		}

		logger.Info("NamespaceOffloading unknown phase has been completed")

		return reconcile.Result{}, nil
	case nsCondition.Status == metav1.ConditionUnknown:
		logger.Info("creating NamespaceOffloading")

		if err := i.CreateNamespaceOffloading(ctx, &intent); err != nil {
			logger.Error(err, "cannot handle NamespaceOffloading creation status")

			return reconcile.Result{}, err
		}

		logger.Info("NamespaceOffloading creation has been completed")

		return reconcile.Result{}, nil
	case nsCondition.Status == metav1.ConditionFalse:
		logger.Info("handling NamespaceOffloading")

		if err := i.HandleNamespaceOffloadingPhase(ctx, &intent); err != nil {
			logger.Error(err, "cannot handle NamespaceOffloading phase")

			return reconcile.Result{}, err
		}

		logger.Info("NamespaceOffloading handling has been completed")

		return reconcile.Result{}, nil
	case nsCondition.Status == metav1.ConditionTrue:
		logger.Info("reconciling NamespaceOffloading")

		if err := i.HandleNamespaceOffloadingPhase(ctx, &intent); err != nil {
			logger.Error(err, "cannot handle NamespaceOffloading phase")

			return reconcile.Result{}, err
		}

		logger.Info("NamespaceOffloading reconciliation has been completed")
	}

	logger.Info("handling resource deployment phase")

	deployCondition := meta.FindStatusCondition(intent.Status.Conditions, flarev1alpha1.IntentStatusTypeDeploy)
	switch {
	case deployCondition == nil:
		logger.Info("Deploy phase is unknown")

		if err := i.TrackDeployUnknown(ctx, &intent); err != nil {
			logger.Error(err, "cannot handle Deploy unknown status")

			return reconcile.Result{}, err
		}

		logger.Info("Deploy unknown phase has been completed")

		return reconcile.Result{}, nil
	case deployCondition.Status == metav1.ConditionUnknown, deployCondition.Status == metav1.ConditionFalse:
		logger.Info("creating Deploy")

		if err := i.HandleKubernetesObjects(ctx, &intent); err != nil {
			logger.Error(err, "cannot handle Kubernetes objects deployment")

			return reconcile.Result{}, err
		}

		logger.Info("Deploy creation has been completed")

		return reconcile.Result{}, nil
	case deployCondition.Status == metav1.ConditionTrue:
		logger.Info("reconciling Deploy")

		if err := i.HandleKubernetesObjects(ctx, &intent); err != nil {
			logger.Error(err, "cannot handle Kubernetes objects deployment")

			return reconcile.Result{}, err
		}

		logger.Info("Deploy reconciliation has been completed")
	}

	logger.Info("Intent has been reconciled")

	return reconcile.Result{}, nil
}

func (i *IntentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&flarev1alpha1.Intent{}).
		Watches(&v1beta1.NamespaceOffloading{}, handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			parts := strings.Split(obj.GetNamespace(), "-")

			if len(parts) != 2 {
				return nil
			}

			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Namespace: obj.GetNamespace(),
						Name:      parts[1],
					},
				},
			}
		})).
		Watches(&fluidosnodev1alpha1.Solver{}, handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			parts := strings.Split(obj.GetName(), "-")

			if len(parts) != 2 {
				return nil
			}

			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Namespace: obj.GetName(),
						Name:      parts[1],
					},
				},
			}
		})).
		Complete(i)
}
