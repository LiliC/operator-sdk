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

	"github.com/operator-framework/operator-sdk/pkg/scaffold/input"
)

const MetricsFile = "metrics.go"

type Metrics struct {
	input.Input

	// Resource defines the inputs for the new custom resource definition
	Resource *Resource
}

func (s *Metrics) GetInput() (input.Input, error) {

	if s.Path == "" {
		s.Path = filepath.Join(MetricsDir, MetricsFile)
	}
	// Do not overwrite this file if it exists.
	s.IfExistsAction = input.Skip
	s.TemplateBody = metricsPkgTemplate
	return s.Input, nil
}

const metricsPkgTemplate = `
package metrics

import (
	kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/client-go/rest"
	ksmetrics "k8s.io/kube-state-metrics/pkg/metrics"
)

var resource = "{{ .Resource.APIVersion }}"
var kind = "{{ .Resource.Kind }}"

var (
    MetricFamilies = []ksmetrics.FamilyGenerator{
        ksmetrics.FamilyGenerator{
            Name: "{{ .Resource.LowerKind }}_info",
            Type: ksmetrics.MetricTypeGauge,
            Help: "Information about the {{ .Resource.Kind }} operator replica.",
            GenerateFunc: func(obj interface{}) ksmetrics.Family {
                crd := obj.(*unstructured.Unstructured)

                return ksmetrics.Family{
                    &ksmetrics.Metric{
                        Name:        "{{ .Resource.LowerKind }}_info",
                        Value:       1,
                        LabelKeys:   []string{"namespace", "{{ .Resource.LowerKind }}"},
                        LabelValues: []string{crd.GetNamespace(), crd.GetName()},
                    },
                }
            },
        },
    }
)

func ServeOperatorSpecificMetrics(cfg *rest.Config) {
    uc := kubemetrics.NewForConfig(cfg)

	// TODO: replace namespaces? 
    c := kubemetrics.NewCollector(uc, []string{"default"}, resource, kind, MetricFamilies)

    //prometheus.MustRegister(c)
    kubemetrics.ServeMetrics(c)
}
`
