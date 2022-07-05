package controller

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/symcn/pkg/metrics"
	"k8s.io/klog/v2"
)

var (
	FilterLabels    string
	filterLabelList []string

	clusterLabel = "cluster"
)

type promLabels map[string]string

func initFilterLabelList() {
	filterLabelList = strings.Split(FilterLabels, ",")
}

func newPromLabels(cluster string, labels map[string]string) promLabels {
	if FilterLabels == "" {
		return promLabels{
			clusterLabel: cluster,
		}
	}

	result := make(promLabels, len(filterLabelList))
	for _, label := range filterLabelList {
		newLabel := strings.ReplaceAll(label, "-", "_")
		// prometheus metrics label count must immutable
		result[newLabel] = labels[label]
	}
	result[clusterLabel] = cluster

	return result
}

func startHTTPPrometheus(ctx context.Context) {
	server := &http.Server{
		Addr: ":8080",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprint(rw, "ok")
	})

	metrics.RegisterHTTPHandler(func(pattern string, handler http.Handler) {
		mux.Handle(pattern, handler)
	})

	server.Handler = mux
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if !strings.EqualFold(err.Error(), "http: Server closed") {
				klog.Error(err)
			}
		}
		klog.Info("http shutdown")
	}()
	<-ctx.Done()
	server.Shutdown(context.Background())
}
