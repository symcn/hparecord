package controller

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/symcn/pkg/metrics"
	"k8s.io/klog/v2"
)

var FilterLabels string
var filterLabelList []string

type promLabels map[string]string

func initFilterLabelList() {
	filterLabelList = strings.Split(FilterLabels, ",")
}

func newPromLabels(cluster string, labels map[string]string) promLabels {
	if FilterLabels == "" {
		return promLabels{
			"cluster": cluster,
		}
	}

	result := make(promLabels, len(filterLabelList))
	for _, label := range filterLabelList {
		newLabel := strings.ReplaceAll(label, "-", "_")
		v, ok := labels[label]
		if !ok {
			klog.V(2).Infof("labels %v does not contains label: %s", labels, label)
		}
		result[newLabel] = v
	}
	result["cluster"] = cluster

	return result
}

type value struct {
	TargetValue     float64
	CurrentValue    float64
	CurrentReplicas float64
	MaxReplicas     float64
	MinReplicas     float64
}

func newValue(targetValue, currentValue, currentReplicas, maxReplicas, minReplicas int32) value {
	return value{
		TargetValue:     float64(targetValue),
		CurrentValue:    float64(currentValue),
		CurrentReplicas: float64(currentReplicas),
		MaxReplicas:     float64(maxReplicas),
		MinReplicas:     float64(minReplicas),
	}
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
