package controller

import (
	"github.com/symcn/api"
	"github.com/symcn/pkg/metrics"
)

type cpuMetricsClient struct {
	api.Metrics
}

func newCpuMetricsClient() (*cpuMetricsClient, error) {
	cpuMetrics, err := metrics.NewMetrics("cpu", nil)
	if err != nil {
		return nil, err
	}
	return &cpuMetricsClient{Metrics: cpuMetrics}, nil
}

func (cm *cpuMetricsClient) setPromMetrics(label promLabels, value value) {
	cm.GaugeWithLabels("target_value", label).Set(value.TargetValue)
	cm.GaugeWithLabels("current_value", label).Set(value.CurrentValue)
	cm.GaugeWithLabels("current_replicas", label).Set(value.CurrentReplicas)
	cm.GaugeWithLabels("max_replicas", label).Set(value.MaxReplicas)
	cm.GaugeWithLabels("min_replicas", label).Set(value.MinReplicas)
}

func (cm *cpuMetricsClient) deletePromMetrics(label promLabels) {
	cm.DeleteWithLabels("target_value", label)
	cm.DeleteWithLabels("current_value", label)
	cm.DeleteWithLabels("current_replicas", label)
	cm.DeleteWithLabels("max_replicas", label)
	cm.DeleteWithLabels("min_replicas", label)
}
