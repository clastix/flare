// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/clastix/flare-internal/internal/api"
)

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;delete

type Token struct {
	Helper Helper
	Client client.Client
}

func (t *Token) ListTokens(ctx echo.Context) error {
	user := ctx.Get("user").(authenticationv1.UserInfo)

	tnt, notFoundErr := t.Helper.RetrieveCapsuleTenant(ctx.Request().Context(), user)
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

	var secretList corev1.SecretList
	if err := t.Client.List(ctx.Request().Context(), &secretList, client.InNamespace("tenants"), client.MatchingLabels{"tenant": tnt.Name}); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   err.Error(),
			"context": "cannot list Secret",
		})
	}

	var tokenList []api.Token
	for _, secret := range secretList.Items {
		tokenList = append(tokenList, api.Token{
			CreatedAt: ptr.To(secret.CreationTimestamp.Time),
			Name:      ptr.To(strings.ReplaceAll(secret.Name, secret.GenerateName, "")),
			TokenId:   ptr.To(string(secret.UID)),
		})
	}

	return ctx.JSON(200, api.ListTokensResponse{
		Tokens: &tokenList,
	})
}

func (t *Token) CreateToken(ctx echo.Context) error {
	user := ctx.Get("user").(authenticationv1.UserInfo)

	tnt, notFoundErr := t.Helper.RetrieveCapsuleTenant(ctx.Request().Context(), user)
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

	var secret corev1.Secret
	secret.GenerateName = tnt.Name + "-"
	secret.Namespace = "tenants"
	secret.Labels = map[string]string{
		"tenant": tnt.Name,
	}
	secret.Annotations = map[string]string{
		corev1.ServiceAccountNameKey: tnt.Name,
	}
	secret.Type = corev1.SecretTypeServiceAccountToken

	if err := t.Client.Create(ctx.Request().Context(), &secret); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   err.Error(),
			"context": "cannot create ServiceAccount",
		})
	}

	err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		return err != nil
	}, func() error {
		if err := t.Client.Get(ctx.Request().Context(), client.ObjectKeyFromObject(&secret), &secret); err != nil {
			return err
		}

		if secret.Data == nil {
			return fmt.Errorf("missing data")
		}

		if secret.Data[corev1.ServiceAccountTokenKey] == nil {
			return fmt.Errorf("missing token")
		}

		return nil
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   err.Error(),
			"context": "cannot create Token",
		})
	}

	token, _, tokenErr := jwt.NewParser().ParseUnverified(string(secret.Data[corev1.ServiceAccountTokenKey]), jwt.MapClaims{})
	if tokenErr != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   tokenErr.Error(),
			"context": "cannot parse JWT",
		})
	}

	return ctx.JSON(200, api.CreateTokenResponse{
		CreatedAt:   ptr.To(secret.CreationTimestamp.Time),
		ExpiresAt:   nil,
		Name:        ptr.To(strings.ReplaceAll(secret.Name, secret.GenerateName, "")),
		Permissions: nil,
		Token:       ptr.To(token.Raw),
		TokenId:     ptr.To(string(secret.UID)),
	})
}

func (t *Token) RevokeToken(ctx echo.Context, tokenId string) error {
	user := ctx.Get("user").(authenticationv1.UserInfo)

	tnt, notFoundErr := t.Helper.RetrieveCapsuleTenant(ctx.Request().Context(), user)
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

	var secretList corev1.SecretList
	if err := t.Client.List(ctx.Request().Context(), &secretList, client.InNamespace("tenants"), client.MatchingLabels{"tenant": tnt.Name}); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error":   err.Error(),
			"context": "cannot list Secret",
		})
	}

	for _, secret := range secretList.Items {
		if string(secret.UID) == tokenId {
			if err := t.Client.Delete(ctx.Request().Context(), &secret); err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{
					"error":   err.Error(),
					"context": "cannot delete Token",
				})
			}

			return ctx.JSON(200, api.RevokeTokenResponse{
				Message:   nil,
				RevokedAt: ptr.To(time.Now()),
				TokenId:   &tokenId,
			})
		}
	}

	return ctx.JSON(http.StatusNotFound, map[string]string{
		"error": "token not found",
	})
}
