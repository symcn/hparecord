package controller

import (
	"fmt"
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

	var app string
	if len(hpa.GetLabels()) > 0 {
		app = hpa.GetLabels()[appLabel]
	}
	if app == "" {
		klog.Warningf("hpa: %s does not include label(app)", hpaName)
		return nil
	}

	var minReplicas int32
	if hpa.Spec.MinReplicas != nil {
		minReplicas = *hpa.Spec.MinReplicas
	}

	label := newLabelMap(cluster, hpaName, app)

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
				ctrl.cpuMetricsClient.setPromMetrics(label, value)
			}
		}
	}
	if !found {
		return fmt.Errorf("hpa: %s has no supported metrics", hpaName)
	}
	return nil
}

func (ctrl *Controller) deleteMetrics(cluster string, hpa *v2beta2.HorizontalPodAutoscaler) error {
	hpaName := hpa.GetName()

	var app string
	if len(hpa.GetLabels()) > 0 {
		app = hpa.GetLabels()[appLabel]
	}
	if app == "" {
		klog.Warningf("hpa: %s does not include label(app)", hpaName)
		return nil
	}

	label := newLabelMap(cluster, hpaName, app)

	var found bool
	for _, metric := range hpa.Spec.Metrics {
		// todo support qps
		switch metric.Type {
		case v2beta2.ResourceMetricSourceType:
			switch metric.Resource.Name {
			case cpuName:
				found = true
				ctrl.cpuMetricsClient.deletePromMetrics(label)
			}
		}
	}
	if !found {
		return fmt.Errorf("hpa: %s has no supported metrics", hpaName)
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
