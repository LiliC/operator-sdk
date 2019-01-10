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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	kcoll "k8s.io/kube-state-metrics/pkg/collectors"
	"k8s.io/kube-state-metrics/pkg/metrics"
	metricsstore "k8s.io/kube-state-metrics/pkg/metrics_store"
)

// NewCollector returns a collection of metrics in the namespaces provided, per the api/kind resource.
// The metrics are registered in the custom generateStore function that needs to be defined.
// Note: If namespaces are empty, all namespaces will be fetched.
func NewCollector(uc *Client,
	namespaces []string,
	api string,
	kind string,
	generateStore []metrics.FamilyGenerator) (collectors []*kcoll.Collector) {
	fmt.Println("new collector")
	// TODO: what if namespaces are empty.
	// fetch all namespaces instead
	for _, ns := range namespaces {
		dclient, err := uc.ClientFor(api, kind, ns)
		if err != nil {
			// TODO: log instead
			fmt.Println(err)
			return
		}
		fmt.Println("before headers")
		filteredMetricFamilies := filterMetricFamilies(generateStore)
		composedMetricGenFuncs := composeMetricGenFuncs(filteredMetricFamilies)
		headers := extractMetricFamilyHeaders(filteredMetricFamilies)
		store := metricsstore.NewMetricsStore(headers, composedMetricGenFuncs)
		reflectorPerNamespace(context.TODO(), dclient, &unstructured.Unstructured{}, store, ns)
		collector := kcoll.NewCollector(store)
		fmt.Printf("%#+v", collector)
		collectors = append(collectors, collector)
	}

	return
}

func extractMetricFamilyHeaders(families []metrics.FamilyGenerator) []string {
	headers := make([]string, len(families))

	for i, f := range families {
		header := strings.Builder{}

		header.WriteString("# HELP ")
		header.WriteString(f.Name)
		header.WriteByte(' ')
		header.WriteString(f.Help)
		header.WriteByte('\n')
		header.WriteString("# TYPE ")
		header.WriteString(f.Name)
		header.WriteByte(' ')
		header.WriteString(string(f.Type))

		headers[i] = header.String()
	}

	return headers
}

func filterMetricFamilies(families []metrics.FamilyGenerator) []metrics.FamilyGenerator {
	filtered := []metrics.FamilyGenerator{}

	for _, f := range families {
		filtered = append(filtered, f)
	}
	fmt.Println("filtered")
	fmt.Printf("%#+v", filtered)
	return filtered
}

func composeMetricGenFuncs(families []metrics.FamilyGenerator) func(obj interface{}) []metricsstore.FamilyStringer {
	funcs := []func(obj interface{}) metrics.Family{}

	for _, f := range families {
		funcs = append(funcs, f.GenerateFunc)
	}

	return func(obj interface{}) []metricsstore.FamilyStringer {
		families := make([]metricsstore.FamilyStringer, len(funcs))

		for i, f := range funcs {
			families[i] = f(obj)
		}

		return families
	}
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
