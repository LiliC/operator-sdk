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

package kubemetrics

import (
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	kcoll "k8s.io/kube-state-metrics/pkg/collectors"
)

type MetricHandler struct {
	c []*kcoll.Collector
}

func ServeMetrics(collectors []*kcoll.Collector) {
	listenAddress := net.JoinHostPort("0.0.0.0", "8080")
	mux := http.NewServeMux()
	mux.Handle("/metrics", &MetricHandler{collectors})
	mux.HandleFunc("/healtz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})
	// Add index
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Kube Metrics Server</title></head>
             <body>
             <h1>Kube Metrics</h1>
			 <ul>
             <li><a href='` + "/metrics" + `'>metrics</a></li>
             <li><a href='` + "/healthz" + `'>healthz</a></li>
			 </ul>
             </body>
             </html>`))
	})

	fmt.Println(http.ListenAndServe(listenAddress, mux))
}

func (m *MetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resHeader := w.Header()
	var writer io.Writer = w

	resHeader.Set("Content-Type", `text/plain; version=`+"0.0.4")

	// Gzip response if requested. Taken from
	// github.com/prometheus/client_golang/prometheus/promhttp.decorateWriter.
	reqHeader := r.Header.Get("Accept-Encoding")
	parts := strings.Split(reqHeader, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "gzip" || strings.HasPrefix(part, "gzip;") {
			writer = gzip.NewWriter(writer)
			resHeader.Set("Content-Encoding", "gzip")
		}
	}

	for _, c := range m.c {
		c.Collect(writer)
	}

	// In case we gziped the response, we have to close the writer.
	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}
