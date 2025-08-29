// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	v1 "k8s.io/api/authentication/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=get;list;watch;create

func JWTAuthenticationMiddleware(client client.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing or invalid token"})
			}

			var tokenReview v1.TokenReview
			tokenReview.Spec.Token = strings.TrimPrefix(auth, "Bearer ")
			if err := client.Create(c.Request().Context(), &tokenReview); err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
			}

			if !tokenReview.Status.Authenticated {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthenticated user"})
			}

			if tokenReview.Status.Error != "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "TokenReview returned the following error: " + tokenReview.Status.Error})
			}

			if parts := strings.Split(tokenReview.Status.User.Username, ":"); parts[0] != "system" && parts[1] != "serviceaccount" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "only ServiceAccount can access APIs",
				})
			}

			c.Set("user", tokenReview.Status.User)

			return next(c)
		}
	}
}
