// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	capsuleindexer "github.com/projectcapsule/capsule/pkg/indexer"
	"github.com/projectcapsule/capsule/pkg/indexer/tenant"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrllogger "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/clastix/flare-internal/internal/api"
	"github.com/clastix/flare-internal/internal/handlers"
	"github.com/clastix/flare-internal/internal/indexer"
	"github.com/clastix/flare-internal/internal/middlewares"
	"github.com/clastix/flare-internal/internal/scheme"
)

func main() {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetLevel(log.INFO)
	e.Logger.SetPrefix("webserver")
	e.Logger.(*log.Logger).DisableColor()

	k8sScheme, schemeErr := scheme.New()
	if schemeErr != nil {
		e.Logger.Fatalf("cannot initialize Kubernetes scheme, %s", schemeErr.Error())
	}

	ctrl.SetLogger(logr.New(ctrllogger.NullLogSink{}))

	mgr, mgrErr := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: k8sScheme,
		Metrics: server.Options{
			BindAddress: "0",
		},
		LeaderElection: false,
	})
	if mgrErr != nil {
		e.Logger.Fatalf("cannot initialize manager, %s", mgrErr.Error())
	}

	ctx := ctrl.SetupSignalHandler()

	tenantOwnerRefIndexer, intentUIDIndexer := tenant.OwnerReference{}, indexer.IntentUID{}

	for _, index := range []capsuleindexer.CustomIndexer{tenantOwnerRefIndexer, intentUIDIndexer} {
		if err := mgr.GetFieldIndexer().IndexField(ctx, index.Object(), index.Field(), index.Func()); err != nil {
			e.Logger.Fatalf("cannot initialize indexer, %s", err.Error())
		}
	}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Header.Get(echo.HeaderContentType) != echo.MIMEApplicationJSON {
				return c.JSON(http.StatusUnsupportedMediaType, map[string]string{
					"error": "Content-Type must be application/json",
				})
			}
			return next(c)
		}
	})
	e.Use(middleware.Logger())
	e.Use(middlewares.JWTAuthenticationMiddleware(mgr.GetClient()))

	helper := handlers.Helper{
		Client:                mgr.GetClient(),
		TenantOwnerRefIndexer: tenantOwnerRefIndexer,
	}

	api.RegisterHandlers(e, &handlers.Server{
		Intent: handlers.Intent{
			Helper:           helper,
			Client:           mgr.GetClient(),
			IntentUIDIndexer: intentUIDIndexer,
		},
		Token: handlers.Token{
			Helper: helper,
			Client: mgr.GetClient(),
		},
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: e,
	}

	go func() {
		e.Logger.Infof("server started at %s", srv.Addr)
		if err := e.StartServer(srv); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal("shutting down server")
		}
	}()

	if err := mgr.Start(ctx); err != nil {
		e.Logger.Errorf("cannot stop manager: %s", err.Error())
	}

	e.Logger.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		e.Logger.Fatal(err)
	}

	e.Logger.Info("shut down server completed")
}
