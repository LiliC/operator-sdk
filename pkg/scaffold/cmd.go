// Copyright 2018 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scaffold

import (
	"path/filepath"

	"github.com/operator-framework/operator-sdk/pkg/scaffold/input"
)

const CmdFile = "main.go"

type Cmd struct {
	input.Input
}

func (s *Cmd) GetInput() (input.Input, error) {
	if s.Path == "" {
		s.Path = filepath.Join(ManagerDir, CmdFile)
	}
	s.TemplateBody = cmdTmpl
	return s.Input, nil
}

const cmdTmpl = `package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"{{ .Repo }}/pkg/apis"
	"{{ .Repo }}/pkg/controller"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/ready"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ksmetrics "k8s.io/kube-state-metrics/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var log = logf.Log.WithName("cmd")

var (
    MemcachedMetricFamilies = []ksmetrics.FamilyGenerator{
        ksmetrics.FamilyGenerator{
            Name: "memcached_service_info",
            Type: ksmetrics.MetricTypeGauge,
            Help: "Information about the operator replica.",
            GenerateFunc: func(obj interface{}) ksmetrics.Family {
                crd := obj.(*unstructured.Unstructured)

                return ksmetrics.Family{
                    &ksmetrics.Metric{
                        Name:        "memcached_service_info",
                        Value:       1,
                        LabelKeys:   []string{"namespace", "memcached"},
                        LabelValues: []string{crd.GetNamespace(), crd.GetName()},
                    },
                }
            },
        },
    }
)

func printVersion() {
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("operator-sdk Version: %v", sdkVersion.Version))
}

func main() {
	flag.Parse()

	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(logf.ZapLogger(false))

	printVersion()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Error(err, "failed to get watch namespace")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Become the leader before proceeding
	leader.Become(context.TODO(), "{{ .ProjectName }}-lock")

    // operator-specific-metrics aka kube-metrics
    uc := metrics.NewForConfig(cfg)
    resource := "metric.example.com/v1alpha1"
    kind := "MetricService"

    c := kubemetrics.NewCollector(uc, []string{"default"}, resource, kind, MemcachedMetricFamilies)

    //prometheus.MustRegister(c)
    kubemetrics.ServeMetrics(c)

	r := ready.NewFileReady()
	err = r.Set()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	defer r.Unset()

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{Namespace: namespace})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "manager exited non-zero")
		os.Exit(1)
	}
}
`
