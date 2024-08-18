/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"crypto/tls"
	_ "embed"
	"flag"
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"aedificans.com/k8s-ssm-param-injector/pkg/injector"
	"aedificans.com/k8s-ssm-param-injector/pkg/utils"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	// +kubebuilder:scaffold:imports
)

//go:generate sh -c "printf %s $(git rev-parse --short HEAD) > commit.txt"
//go:embed commit.txt
var Commit string

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var awsRegion string
	var enableHTTP2 bool
	var enableLeaderElection bool
	var metricsAddr string
	var probeAddr string
	var secureMetrics bool
	var webhookPort int
	flag.StringVar(&awsRegion, "aws-region", utils.GetEnvString("AWS_REGION", "us-east-1"),
		"The AWS region for the SSM client to create a session in for the service.")
	flag.BoolVar(&enableHTTP2, "enable-http2", utils.GetEnvBool("ENABLE_HTTP2", false),
		"If set, HTTP/2 will be enabled for the metrics and webhook servers.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", utils.GetEnvString("HEALTH_PROBE_BIND_ADDRESS", ":8081"),
		"The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", utils.GetEnvBool("LEADER_ELECT", false),
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", utils.GetEnvString("METRICS_BIND_ADDRESS", "0"),
		"The address the metrics endpoint binds to. Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.BoolVar(&secureMetrics, "metrics-secure", utils.GetEnvBool("METRICS_SECURE", true),
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.IntVar(&webhookPort, "webhook-address", utils.GetEnvInt("WEBHOOK_PORT", 8443),
		"The port of the webhook server for the mutating webhook.")
	opts := zap.Options{
		Encoder: zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			MessageKey:   "msg",
			LevelKey:     "level",
			TimeKey:      "time",
			NameKey:      "name",
			CallerKey:    "caller",
			FunctionKey:  "function",
			EncodeCaller: zapcore.FullCallerEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
		}),
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)).
		WithValues("app", "ssm-param-injector", "commit", Commit))

	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	tlsOpts := []func(*tls.Config){}
	tlsOpts = append(tlsOpts, disableHTTP2)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:   metricsAddr,
			SecureServing: secureMetrics,
			TLSOpts:       tlsOpts,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "57458571.aedificans.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(awsRegion))
	if err != nil {
		setupLog.Error(err, "failed to load aws config")
		panic(err)
	}

	ssmClient := ssm.NewFromConfig(cfg)

	webhookServer := webhook.NewServer(webhook.Options{
		CertDir: "ssl",
		Port:    webhookPort,
		TLSOpts: tlsOpts,
	})
	webhookServer.Register("/mutate", &webhook.Admission{
		Handler: &injector.SSMParameterInjector{
			SsmClient: ssmClient,
			Decoder:   admission.NewDecoder(scheme)}})
	mgr.Add(webhookServer)

	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
