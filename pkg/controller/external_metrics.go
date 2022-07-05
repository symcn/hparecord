package controller

import (
	"github.com/symcn/api"
	"github.com/symcn/pkg/metrics"
)

// externalMetricsClient correspond keda external type
type externalMetricsClient struct {
	api.Metrics
}

func newExternalMetricsClient(kind string) (*externalMetricsClient, error) {
	m, err := metrics.NewMetrics(kind, nil)
	if err != nil {
		return nil, err
	}
	return &externalMetricsClient{Metrics: m}, nil
}

func (c *externalMetricsClient) setPromMetrics(label promLabels, value metricsValue) {
	c.GaugeWithLabels("target_value", label).Set(value.TargetValue)
	c.GaugeWithLabels("current_value", label).Set(value.CurrentValue)
}

func (c *externalMetricsClient) deletePromMetrics(label promLabels) {
	c.DeleteWithLabels("target_value", label)
	c.DeleteWithLabels("current_value", label)
}

type metricsValue struct {
	TargetValue  float64
	CurrentValue float64
}

func newMetricsValue(targetValue, currentValue int64) metricsValue {
	return metricsValue{
		TargetValue:  float64(targetValue),
		CurrentValue: float64(currentValue),
	}
}
