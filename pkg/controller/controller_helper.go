package controller

import (
	"github.com/symcn/hparecord/pkg/metrics"
	"k8s.io/api/autoscaling/v2beta2"
)

const (
	// todo support qps
	cpuName = "cpu"
)

func handleMetrics(cluster string, hpa *v2beta2.HorizontalPodAutoscaler) error {
	for _, metric := range hpa.Spec.Metrics {
		metadata := metrics.NewMetadata(
			cluster,
			hpa.GetName(),
			hpa.GetLabels()["app"],
			hpa.GetLabels()["appCode"],
			hpa.GetLabels()["projectCode"],
			hpa.Status.CurrentReplicas,
			*hpa.Spec.MinReplicas,
			hpa.Spec.MaxReplicas,
		)
		// todo support qps
		switch metric.Type {
		case v2beta2.ResourceMetricSourceType:
			switch metric.Resource.Name {
			case cpuName:
				err := handleCpuMetrics(metadata, metric, hpa.Status)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func handleCpuMetrics(metadata *metrics.Metadata, metric v2beta2.MetricSpec, status v2beta2.HorizontalPodAutoscalerStatus) error {
	var (
		targetCpuValue  int32
		currentCpuValue int32
	)
	targetCpuValue = *metric.Resource.Target.AverageUtilization
	for _, m := range status.CurrentMetrics {
		if m.Type == v2beta2.ResourceMetricSourceType {
			if m.Resource.Name == cpuName {
				currentCpuValue = *m.Resource.Current.AverageUtilization
			}
		}
	}

	cpuMetric := metrics.NewCpuMetric(metadata, targetCpuValue, currentCpuValue)
	return cpuMetric.SetMetrics()
}
