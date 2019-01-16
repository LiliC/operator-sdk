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
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	kcoll "k8s.io/kube-state-metrics/pkg/collector"
	"k8s.io/kube-state-metrics/pkg/metrics"
	metricsstore "k8s.io/kube-state-metrics/pkg/metrics_store"
)

// NewCollector returns a collection of metrics in the namespaces provided, per the api/kind resource.
// The metrics are registered in the custom generateStore function that needs to be defined.
// Note: If namespaces are empty, current namespace all namespaces will be taken.
func NewCollector(uc *Client,
	namespaces []string,
	api string,
	kind string,
	generateStore []metrics.FamilyGenerator) (collectors []*kcoll.Collector) {
	fmt.Println("new collector")
	// TODO: what if namespaces are empty.
	// fetch all namespaces instead

	// add current namespace to the list of ns
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		fmt.Printf("failed to get ns: (%v)", err)
	}

	namespaces = append(namespaces, namespace)

	for _, ns := range namespaces {
		dclient, err := uc.ClientFor(api, kind, ns)
		if err != nil {
			fmt.Println(err)
			return
		}
		composedMetricGenFuncs := kcoll.ComposeMetricGenFuncs(generateStore)
		headers := kcoll.ExtractMetricFamilyHeaders(generateStore)
		store := metricsstore.NewMetricsStore(headers, composedMetricGenFuncs)
		reflectorPerNamespace(context.TODO(), dclient, &unstructured.Unstructured{}, store, ns)
		collector := kcoll.NewCollector(store)
		collectors = append(collectors, collector)
	}

	return
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
