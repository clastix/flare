// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"fmt"
	"net/http"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"github.com/projectcapsule/capsule/pkg/indexer"
	authenticationv1 "k8s.io/api/authentication/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Helper struct {
	Client                client.Client
	TenantOwnerRefIndexer indexer.CustomIndexer
}

//+kubebuilder:rbac:groups=capsule.clastix.io,resources=tenants,verbs=get;list;watch

func (i *Helper) RetrieveCapsuleTenant(ctx context.Context, user authenticationv1.UserInfo) (*capsulev1beta2.Tenant, error) {
	var tntList capsulev1beta2.TenantList
	if err := i.Client.List(ctx, &tntList, client.MatchingFields{i.TenantOwnerRefIndexer.Field(): fmt.Sprintf("ServiceAccount:%s", user.Username)}); err != nil {
		return nil, err
	}

	if len(tntList.Items) == 0 {
		return nil, &apierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound, Code: http.StatusNotFound}}
	}

	return tntList.Items[0].DeepCopy(), nil
}
