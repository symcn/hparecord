package controller

import (
	"k8s.io/api/autoscaling/v2beta2"
	"k8s.io/klog/v2"
)

const (
	// todo support qps
	cpuName  = "cpu"
	appLabel = "app"
)

func (ctrl *Controller) handleMetrics(cluster string, hpa *v2beta2.HorizontalPodAutoscaler) error {
	hpaName := hpa.GetName()
	labels := hpa.GetLabels()

	var app string
	if len(labels) > 0 {
		app = labels[appLabel]
	}
	if app == "" {
		klog.Warningf("hpa: %s does not include label(app)", hpaName)
		return nil
	}

	var minReplicas int32
	if hpa.Spec.MinReplicas != nil {
		minReplicas = *hpa.Spec.MinReplicas
	}

	promLabels := newPromLabels(cluster, labels)

	var found bool
	for _, metric := range hpa.Spec.Metrics {
		// todo support qps
		switch metric.Type {
		case v2beta2.ResourceMetricSourceType:
			switch metric.Resource.Name {
			case cpuName:
				found = true
				targetCpuValue, currentCpuValue := calCpuMetricValue(metric, hpa.Status)
				value := newValue(targetCpuValue, currentCpuValue, hpa.Status.CurrentReplicas, hpa.Spec.MaxReplicas, minReplicas)
				ctrl.cpuMetricsClient.setPromMetrics(promLabels, value)
			}
		}
	}
	if !found {
		klog.Warningf("hpa: %s has no supported metrics", hpaName)
		return nil
	}
	return nil
}

func (ctrl *Controller) deleteMetrics(cluster string, hpa *v2beta2.HorizontalPodAutoscaler) error {
	hpaName := hpa.GetName()
	labels := hpa.GetLabels()

	var app string
	if len(labels) > 0 {
		app = labels[appLabel]
	}
	if app == "" {
		klog.Warningf("hpa: %s does not include label(app)", hpaName)
		return nil
	}

	promLabels := newPromLabels(cluster, labels)

	var found bool
	for _, metric := range hpa.Spec.Metrics {
		// todo support qps
		switch metric.Type {
		case v2beta2.ResourceMetricSourceType:
			switch metric.Resource.Name {
			case cpuName:
				found = true
				ctrl.cpuMetricsClient.deletePromMetrics(promLabels)
			}
		}
	}
	if !found {
		klog.Warningf("hpa: %s has no supported metrics", hpaName)
		return nil
	}
	return nil
}

func calCpuMetricValue(metric v2beta2.MetricSpec, status v2beta2.HorizontalPodAutoscalerStatus) (targetCpuValue, currentCpuValue int32) {
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
	return targetCpuValue, currentCpuValue
}
