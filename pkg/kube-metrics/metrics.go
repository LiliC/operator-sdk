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
	ksm "k8s.io/kube-state-metrics/pkg/metric"
)

var log = logf.Log.WithName("metrics")

// ServeOperatorSpecificMetrics generates CR specific metrics and starts serving those metrics.
func ServeOperatorSpecificMetrics(cfg *rest.Config, ns, host string, port int32, MetricFamilies *[]ksm.FamilyGenerator) error {
	uc := NewForConfig(cfg)
	c, err := NewCollector(uc, namespaces, resource, kind, MetricFamilies)
	if err != nil {
		if err == k8sutil.ErrNoNamespace {
			log.Info("Skipping operator specific metrics; not running in a cluster.")
			return nil
		}
		return err
	}

	go ServeMetrics(c, host, port)
	return nil
}
