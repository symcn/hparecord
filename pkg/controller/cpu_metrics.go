package controller

import (
	"github.com/symcn/api"
	"github.com/symcn/pkg/metrics"
)

type cpuMetricsClient struct {
	api.Metrics
}

func newCpuMetricsClient() (*cpuMetricsClient, error) {
	cpuMetrics, err := metrics.NewMetrics(cpuName, nil)
	if err != nil {
		return nil, err
	}
	return &cpuMetricsClient{Metrics: cpuMetrics}, nil
}

func (c *cpuMetricsClient) setPromMetrics(label promLabels, value metricsValue) {
	c.GaugeWithLabels("target_value", label).Set(value.TargetValue)
	c.GaugeWithLabels("current_value", label).Set(value.CurrentValue)
}

func (c *cpuMetricsClient) deletePromMetrics(label promLabels) {
	c.DeleteWithLabels("target_value", label)
	c.DeleteWithLabels("current_value", label)
}
