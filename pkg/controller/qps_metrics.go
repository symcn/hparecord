package controller

import (
	"github.com/symcn/api"
	"github.com/symcn/pkg/metrics"
)

type qpsMetricsClient struct {
	api.Metrics
}

func newQpsMetricsClient() (*qpsMetricsClient, error) {
	qpsMetrics, err := metrics.NewMetrics(qpsName, nil)
	if err != nil {
		return nil, err
	}
	return &qpsMetricsClient{Metrics: qpsMetrics}, nil
}

func (c *qpsMetricsClient) setPromMetrics(label promLabels, value metricsValue) {
	c.GaugeWithLabels("target_value", label).Set(value.TargetValue)
	c.GaugeWithLabels("current_value", label).Set(value.CurrentValue)
}

func (c *qpsMetricsClient) deletePromMetrics(label promLabels) {
	c.DeleteWithLabels("target_value", label)
	c.DeleteWithLabels("current_value", label)
}
