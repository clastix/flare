// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	flarev1alpha1 "github.com/clastix/flare-internal/api/v1alpha1"
)

func UpdateStatusCondition(ctx context.Context, clt client.Client, intent *flarev1alpha1.Intent, condition metav1.Condition) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := clt.Get(ctx, client.ObjectKeyFromObject(intent), intent); err != nil {
			return err
		}

		meta.SetStatusCondition(&intent.Status.Conditions, condition)

		return clt.Status().Update(ctx, intent)
	})
}
