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
	"path/filepath"

	"github.com/operator-framework/operator-sdk/internal/pkg/scaffold/input"
)

const (
	metricsFile = "metrics.go"
	metricsDir  = PkgDir + filePathSep + "metrics"
)

type Metrics struct {
	input.Input

	// Resource defines the inputs for the new custom resource definition
	Resource *Resource
}

func (s *Metrics) GetInput() (input.Input, error) {
	if s.Path == "" {
		s.Path = filepath.Join(metricsDir, metricsFile)
	}
	s.TemplateBody = metricsPkgTemplate
	return s.Input, nil
}

const metricsPkgTemplate = `package metrics

import (
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	ksmetric "k8s.io/kube-state-metrics/pkg/metric"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	log        = logf.Log.WithName("metrics")
	resource   = "{{ .Resource.APIVersion }}"
	kind       = "{{ .Resource.Kind }}"
	metricName = "{{ .Resource.LowerKind }}_info"

	MetricFamilies = []ksmetric.FamilyGenerator{
        ksmetric.FamilyGenerator{
            Name: metricName,
            Type: ksmetric.Gauge,
            Help: "Information about the {{ .Resource.Kind }} operator replica.",
            GenerateFunc: func(obj interface{}) *ksmetric.Family {
                crd := obj.(*unstructured.Unstructured)

                return &ksmetric.Family{
					Metrics: []*ksmetric.Metric{
						{
							Value:       1,
							LabelKeys:   []string{"namespace", "{{ .Resource.LowerKind }}"},
							LabelValues: []string{crd.GetNamespace(), crd.GetName()},
						},
					},
				}
			},
		},
	}
)

func ServeOperatorSpecificMetrics(cfg *rest.Config, ns, host string, port int32) error {
	err := ksmetric.ServeOperatorSpecificMetrics(cfg, ns, host, port, MetricFamilies, resource, kind)
	if err != nil {
		return err
	}
}
`
