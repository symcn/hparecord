package controller

import (
	"github.com/symcn/api"
	"github.com/symcn/pkg/metrics"
)

type qpmMetricsClient struct {
	api.Metrics
}

func newQpmMetricsClient() (*qpmMetricsClient, error) {
	qpmMetrics, err := metrics.NewMetrics(qpmName, nil)
	if err != nil {
		return nil, err
	}
	return &qpmMetricsClient{Metrics: qpmMetrics}, nil
}

func (c *qpmMetricsClient) setPromMetrics(label promLabels, value metricsValue) {
	c.GaugeWithLabels("target_value", label).Set(value.TargetValue)
	c.GaugeWithLabels("current_value", label).Set(value.CurrentValue)
}

func (c *qpmMetricsClient) deletePromMetrics(label promLabels) {
	c.DeleteWithLabels("target_value", label)
	c.DeleteWithLabels("current_value", label)
}
