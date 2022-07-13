package controller

import (
	"github.com/symcn/api"
	"github.com/symcn/pkg/metrics"
)

type hpaMetricsClient struct {
	api.Metrics
}

const hpaName = "hpa"

func newHpaMetricsClient() (*hpaMetricsClient, error) {
	hpaMetrics, err := metrics.NewMetrics(hpaName, nil)
	if err != nil {
		return nil, err
	}
	return &hpaMetricsClient{Metrics: hpaMetrics}, nil
}

func (c *hpaMetricsClient) setPromMetrics(label promLabels, value hpaValue) {
	c.GaugeWithLabels("current_replicas", label).Set(value.CurrentReplicas)
	c.GaugeWithLabels("max_replicas", label).Set(value.MaxReplicas)
	c.GaugeWithLabels("min_replicas", label).Set(value.MinReplicas)
}

func (c *hpaMetricsClient) deletePromMetrics(label promLabels) {
	c.DeleteWithLabels("current_replicas", label)
	c.DeleteWithLabels("max_replicas", label)
	c.DeleteWithLabels("min_replicas", label)
}

type hpaValue struct {
	CurrentReplicas float64
	MaxReplicas     float64
	MinReplicas     float64
}

func newHpaValue(currentReplicas, maxReplicas, minReplicas int32) hpaValue {
	return hpaValue{
		CurrentReplicas: float64(currentReplicas),
		MaxReplicas:     float64(maxReplicas),
		MinReplicas:     float64(minReplicas),
	}
}
