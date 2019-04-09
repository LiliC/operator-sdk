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

package kubemetrics

import (
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"

	"k8s.io/client-go/rest"
	kcollector "k8s.io/kube-state-metrics/pkg/collector"
	ksm "k8s.io/kube-state-metrics/pkg/metric"
)

var log = logf.Log.WithName("kubemetrics")

type KubeMetrics struct {
	Resource       string
	Kind           string
	MetricFamilies *[]ksm.FamilyGenerator
}

// ServeCRMetrics generates CR specific metrics.
// It starts serving collections of those metrics on given host and port.
func ServeCRMetrics(cfg *rest.Config, ns []string, kubeMetrics []KubeMetrics, host string, port int32) error {
	// Get unstructured client.
	uc := NewClientForConfig(cfg)

	// TODO:
	// gather all the different resource/kind and metric families from the different CR
	// loop through them and create new collectors for all of them.

	var collectors [][]*kcollector.Collector
	for km := range kubeMetrics {
		// Generate collector based on the resource, kind and the metric family.
		c, err := NewCollector(uc, ns, km.Resource, km.Kind, km.MetricFamilies)
		if err != nil {
			if err == k8sutil.ErrNoNamespace {
				log.Info("Skipping operator specific metrics; not running in a cluster.")
				return nil
			}
			return err
		}
		collectors = append(collectors, c)
	}
	// Start serving metrics.
	go ServeMetrics(collectors, host, port)

	return nil
}

// NewKubeMetrics creates and populates a new KubeMetrics struct.
func NewKubeMetrics(resource, kind string, mf *[]ksm.FamilyGenerator) KubeMetrics {
	return KubeMetrics{
		Resource:        resource,
		Kind:            kind,
		MetricsFamilies: mf,
	}
}
