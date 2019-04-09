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
	"context"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	kcollector "k8s.io/kube-state-metrics/pkg/collector"
	"k8s.io/kube-state-metrics/pkg/metric"
	metricsstore "k8s.io/kube-state-metrics/pkg/metrics_store"
)

// NewCollectors returns collections of metrics in the namespaces provided, per the api/kind resource.
// The metrics are registered in the custom generateStore function that needs to be defined.
// Current operators namespace will be added to the passed namespaces if it is not present.
func NewCollectors(uc *Client,
	namespaces []string,
	api string,
	kind string,
	generateStore []metric.FamilyGenerator) ([]*kcollector.Collector, error) {
	var collectors []*kcollector.Collector
	// Detect in which namespace the operator is deployed in.
	ns, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return nil, err
	}
	namespaces = append(namespaces, ns)
	namespaces = deduplicateNamespaces(namespaces)

	// Get collector per namespace.
	for _, ns := range namespaces {
		dclient, err := uc.ClientFor(api, kind, ns)
		if err != nil {
			return nil, err
		}
		composedMetricGenFuncs := metric.ComposeMetricGenFuncs(generateStore)
		headers := metric.ExtractMetricFamilyHeaders(generateStore)
		store := metricsstore.NewMetricsStore(headers, composedMetricGenFuncs)
		reflectorPerNamespace(context.TODO(), dclient, &unstructured.Unstructured{}, store, ns)
		collector := kcollector.NewCollector(store)
		collectors = append(collectors, collector)
	}

	return collectors, nil
}

func deduplicateNamespaces(ns []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range ns {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func reflectorPerNamespace(
	ctx context.Context,
	dynamicInterface dynamic.NamespaceableResourceInterface,
	expectedType interface{},
	store cache.Store,
	ns string,
) {
	lw := listWatchFunc(dynamicInterface, ns)
	reflector := cache.NewReflector(&lw, expectedType, store, 0)
	go reflector.Run(ctx.Done())
}

func listWatchFunc(dynamicInterface dynamic.NamespaceableResourceInterface, namespace string) cache.ListWatch {
	return cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return dynamicInterface.Namespace(namespace).List(opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return dynamicInterface.Namespace(namespace).Watch(opts)
		},
	}
}
