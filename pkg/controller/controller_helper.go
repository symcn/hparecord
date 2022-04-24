package controller

import (
	"github.com/symcn/hparecord/pkg/metrics"
	"k8s.io/api/autoscaling/v2beta2"
	"k8s.io/klog/v2"
)

const (
	// todo support qps
	cpuName = "cpu"
)

func handleMetrics(cluster string, hpa *v2beta2.HorizontalPodAutoscaler) error {
	hpaName := hpa.GetName()
	app := hpa.GetLabels()["app"]
	appCode := hpa.GetLabels()["appCode"]
	projectCode := hpa.GetLabels()["projectCode"]

	if app == "" || appCode == "" || projectCode == "" {
		klog.Warningf("hpa: %s not included app appcode projectCode label", hpaName)
		return nil
	}

	var minReplicas int32
	if hpa.Spec.MinReplicas != nil {
		minReplicas = *hpa.Spec.MinReplicas
	}

	for _, metric := range hpa.Spec.Metrics {
		metadata := metrics.NewMetadata(
			cluster,
			hpaName,
			app,
			appCode,
			projectCode,
			hpa.Status.CurrentReplicas,
			minReplicas,
			hpa.Spec.MaxReplicas,
		)
		// todo support qps
		switch metric.Type {
		case v2beta2.ResourceMetricSourceType:
			switch metric.Resource.Name {
			case cpuName:
				if err := handleCpuMetrics(metadata, metric, hpa.Status); err != nil {
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
	if metric.Resource.Target.AverageUtilization != nil {
		targetCpuValue = *metric.Resource.Target.AverageUtilization
	}
	for _, m := range status.CurrentMetrics {
		if m.Type == v2beta2.ResourceMetricSourceType {
			if m.Resource.Name == cpuName {
				if m.Resource.Current.AverageUtilization != nil {
					currentCpuValue = *m.Resource.Current.AverageUtilization
				}
			}
		}
	}

	cpuMetric := metrics.NewCpuMetric(metadata, targetCpuValue, currentCpuValue)
	return cpuMetric.SetMetrics()
}
