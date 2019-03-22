// Copyright 2019 The Operator-SDK Authors
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
	"testing"

	"github.com/operator-framework/operator-sdk/internal/util/diffutil"
)

func TestMetrics(t *testing.T) {
	r, err := NewResource(appApiVersion, appKind)
	if err != nil {
		t.Fatal(err)
	}
	s, buf := setupScaffoldAndWriter()
	err = s.Execute(appConfig, &Metrics{Resource: r})
	if err != nil {
		t.Fatalf("Failed to execute the scaffold: (%v)", err)
	}

	if metricsExp != buf.String() {
		diffs := diffutil.Diff(metricsExp, buf.String())
		t.Fatalf("Expected vs actual differs.\n%v", diffs)
	}
}

const metricsExp = `package metrics

import (
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	ksmetric "k8s.io/kube-state-metrics/pkg/metric"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("metrics")
var resource = "app.example.com/v1alpha1"
var kind = "AppService"
var metricName = "appservice_info"
var (
	MetricFamilies = []ksmetric.FamilyGenerator{
		ksmetric.FamilyGenerator{
			Name: metricName,
			Type: ksmetric.Gauge,
			Help: "Information about the AppService operator replica.",
			GenerateFunc: func(obj interface{}) *ksmetric.Family {
				crd := obj.(*unstructured.Unstructured)

				return &ksmetric.Family{
					Metrics: []*ksmetric.Metric{
						{
							Value:       1,
							LabelKeys:   []string{"namespace", "appservice"},
							LabelValues: []string{crd.GetNamespace(), crd.GetName()},
						},
					},
				}
			},
		},
	}
)

func ServeOperatorSpecificMetrics(cfg *rest.Config, host string, port int32) error {
	uc := kubemetrics.NewForConfig(cfg)
	// By default the current namespace will be detected and used to create metrics.
	// Add to the namespaces to include any other namespaces.
	namespaces := []string{}

	c, err := kubemetrics.NewCollector(uc, namespaces, resource, kind, MetricFamilies)
	if err != nil {
		if err == k8sutil.ErrNoNamespace {
			log.Info("Skipping operator specific metrics; not running in a cluster.")
			return nil
		}
		return err
	}

	go kubemetrics.ServeMetrics(c, host, port)
	return nil
}
`
