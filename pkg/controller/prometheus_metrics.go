package controller

import (
	"context"
	"fmt"
	"github.com/symcn/pkg/metrics"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
)

type labelMap map[string]string

func newLabelMap(cluster, hpa, app string) labelMap {
	return map[string]string{
		"cluster": cluster,
		"hpa":     hpa,
		"app":     app,
	}
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
