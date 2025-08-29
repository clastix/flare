// Copyright 2025 Clastix Labs
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"

	"github.com/spf13/pflag"
	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/clastix/flare-internal/internal/controllers"
	"github.com/clastix/flare-internal/internal/scheme"
)

func main() {
	setupLog := ctrl.Log.WithName("setup")

	var enableLeaderElection bool
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
		EncoderConfigOptions: append([]zap.EncoderConfigOption{}, func(config *zapcore.EncoderConfig) {
			config.EncodeTime = zapcore.ISO8601TimeEncoder
		}),
		Level: zapcore.InfoLevel,
	}

	var goFlagSet flag.FlagSet
	opts.BindFlags(&goFlagSet)
	pflag.CommandLine.AddGoFlagSet(&goFlagSet)

	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	scheme, schemeErr := scheme.New()
	if schemeErr != nil {
		setupLog.Error(schemeErr, "failed to initialize scheme")
		os.Exit(1)
	}

	ctx := ctrl.SetupSignalHandler()

	mgr, mgrErr := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: ":8081",
		},
		LeaderElection:   enableLeaderElection,
		LeaderElectionID: "flare.clastix.io",
	})
	if mgrErr != nil {
		setupLog.Error(mgrErr, "unable to create manager")
		os.Exit(1)
	}

	if err := (&controllers.IntentReconciler{Client: mgr.GetClient()}).SetupWithManager(mgr); err != nil {
		setupLog.Error(mgrErr, "unable to setup controllers.IntentReconciler")
		os.Exit(1)
	}

	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
}
