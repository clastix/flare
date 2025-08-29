// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package indexer

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	flarev1alpha1 "github.com/clastix/flare-internal/api/v1alpha1"
)

type IntentUID struct{}

func (i IntentUID) Object() client.Object {
	return &flarev1alpha1.Intent{}
}

func (i IntentUID) Field() string {
	return "metadata.uid"
}

func (i IntentUID) Func() client.IndexerFunc {
	return func(object client.Object) []string {
		return []string{string(object.GetUID())}
	}
}
