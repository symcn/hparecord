package metrics

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
)

type Metadata struct {
	Cluster string
	Hpa     string

	App         string
	AppCode     string
	ProjectCode string

	CurrentReplicas int32
	MinReplicas     int32
	MaxReplicas     int32
}

func NewMetadata(cluster, hpa, app, appCode, projectCode string, currentReplicas int32, minReplicas int32, maxReplicas int32) *Metadata {
	return &Metadata{
		Cluster:         cluster,
		Hpa:             hpa,
		App:             app,
		AppCode:         appCode,
		ProjectCode:     projectCode,
		CurrentReplicas: currentReplicas,
		MinReplicas:     minReplicas,
		MaxReplicas:     maxReplicas,
	}
}

func (m *Metadata) setBaseMetrics() error {
	gauge, err := currentReplicasGaugeVec.GetMetricWithLabelValues(m.Cluster, m.Hpa, m.App, m.AppCode, m.ProjectCode)
	if err != nil {
		return err
	}
	gauge.Set(float64(m.CurrentReplicas))

	gauge, err = maxReplicasGaugeVec.GetMetricWithLabelValues(m.Cluster, m.Hpa, m.App, m.AppCode, m.ProjectCode)
	if err != nil {
		return err
	}
	gauge.Set(float64(m.MaxReplicas))

	gauge, err = minReplicasGaugeVec.GetMetricWithLabelValues(m.Cluster, m.Hpa, m.App, m.AppCode, m.ProjectCode)
	if err != nil {
		return err
	}
	gauge.Set(float64(m.MinReplicas))

	return nil
}

func (m *Metadata) deleteBaseMetrics() {
	if ok := currentReplicasGaugeVec.DeleteLabelValues(m.Cluster, m.Hpa, m.App, m.AppCode, m.ProjectCode); !ok {
		klog.Warningf("unable to delete currentReplicasGaugeVec, metrics: %+v\n", m)
	}
	if ok := minReplicasGaugeVec.DeleteLabelValues(m.Cluster, m.Hpa, m.App, m.AppCode, m.ProjectCode); !ok {
		klog.Warningf("unable to delete minReplicasGaugeVec, metrics: %+v\n", m)
	}
	if ok := maxReplicasGaugeVec.DeleteLabelValues(m.Cluster, m.Hpa, m.App, m.AppCode, m.ProjectCode); !ok {
		klog.Warningf("unable to delete maxReplicasGaugeVec, metrics: %+v\n", m)
	}
}

func Start(ctx context.Context) {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", 8080),
	}
	mux := http.NewServeMux()

	r := prometheus.NewRegistry()
	r.MustRegister(collectors...)

	mux.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))

	server.Handler = mux
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if !strings.EqualFold(err.Error(), "http: Server closed") {
				klog.Error(err)
				return
			}
		}
		klog.Info("http shutdown")
	}()
	<-ctx.Done()
	server.Shutdown(context.Background())
}
