// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flarev1alpha1 "github.com/clastix/flare-internal/api/v1alpha1"
)

func (i *IntentReconciler) TrackDeployUnknown(ctx context.Context, intent *flarev1alpha1.Intent) error {
	meta.SetStatusCondition(&intent.Status.Conditions, metav1.Condition{
		Type:    flarev1alpha1.IntentStatusTypeDeploy,
		Status:  metav1.ConditionUnknown,
		Reason:  "ResourceDeployment",
		Message: "Resource deployment is about to start",
	})

	return i.Client.Status().Update(ctx, intent)
}
