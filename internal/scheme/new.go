// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package scheme

import (
	fluidosadvertisementv1alpha1 "github.com/fluidos-project/node/apis/advertisement/v1alpha1"
	fluidosnodesv1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	fluidosreservationv1alpha1 "github.com/fluidos-project/node/apis/reservation/v1alpha1"
	liqooffloadingv1beta1 "github.com/liqotech/liqo/apis/offloading/v1beta1"
	"github.com/pkg/errors"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	flarev1alpha1 "github.com/clastix/flare-internal/api/v1alpha1"
)

func New() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()

	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, errors.Wrap(err, "unable to register clientgoscheme Scheme")
	}

	if err := flarev1alpha1.AddToScheme(scheme); err != nil {
		return nil, errors.Wrap(err, "unable to register flarev1alpha1 Scheme")
	}

	if err := fluidosnodesv1alpha1.AddToScheme(scheme); err != nil {
		return nil, errors.Wrap(err, "unable to register fluidosnodesv1alpha1 Scheme")
	}

	if err := fluidosadvertisementv1alpha1.AddToScheme(scheme); err != nil {
		return nil, errors.Wrap(err, "unable to register fluidosadvertisementv1alpha1 Scheme")
	}

	if err := fluidosreservationv1alpha1.AddToScheme(scheme); err != nil {
		return nil, errors.Wrap(err, "unable to register fluidosreservationv1alpha1 Scheme")
	}

	if err := capsulev1beta2.AddToScheme(scheme); err != nil {
		return nil, errors.Wrap(err, "unable to register capsulev1beta2 Scheme")
	}

	if err := liqooffloadingv1beta1.AddToScheme(scheme); err != nil {
		return nil, errors.Wrap(err, "unable to register liqooffloadingv1beta1 Scheme")
	}

	return scheme, nil
}
