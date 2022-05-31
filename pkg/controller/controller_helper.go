package controller

import (
	"strings"

	"k8s.io/api/autoscaling/v2beta2"
	"k8s.io/klog/v2"
)

const (
	cpuName  = "cpu"
	qpsName  = "qps"
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

	// set hpa base metrics
	value := newHpaValue(hpa.Status.CurrentReplicas, hpa.Spec.MaxReplicas, minReplicas)
	ctrl.hpaMetricsClient.setPromMetrics(promLabels, value)

	var found bool
	for _, metric := range hpa.Spec.Metrics {
		switch metric.Type {
		case v2beta2.ResourceMetricSourceType:
			switch metric.Resource.Name {
			case cpuName:
				found = true
				targetCpuValue, currentCpuValue := calCpuMetricValue(metric, hpa.Status)
				value := newMetricsValue(targetCpuValue, currentCpuValue)
				ctrl.cpuMetricsClient.setPromMetrics(promLabels, value)
			}
		case v2beta2.ExternalMetricSourceType:
			metricsType := externalMetricsType(metric.External.Metric.Name)
			switch metricsType {
			case qpsName:
				found = true
				targetCpuValue, currentCpuValue := calQpsMetricValue(metric, hpa.Status)
				value := newMetricsValue(targetCpuValue, currentCpuValue)
				ctrl.qpsMetricsClient.setPromMetrics(promLabels, value)
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

	// delete hpa base metrics
	ctrl.hpaMetricsClient.deletePromMetrics(promLabels)

	var found bool
	for _, metric := range hpa.Spec.Metrics {
		switch metric.Type {
		case v2beta2.ResourceMetricSourceType:
			switch metric.Resource.Name {
			case cpuName:
				found = true
				ctrl.cpuMetricsClient.deletePromMetrics(promLabels)
			}
		case v2beta2.ExternalMetricSourceType:
			metricsType := externalMetricsType(metric.External.Metric.Name)
			switch metricsType {
			case qpsName:
				found = true
				ctrl.qpsMetricsClient.deletePromMetrics(promLabels)
			}
		}
	}
	if !found {
		klog.Warningf("hpa: %s has no supported metrics", hpaName)
		return nil
	}
	return nil
}

func calCpuMetricValue(metric v2beta2.MetricSpec, status v2beta2.HorizontalPodAutoscalerStatus) (targetCpuValue, currentCpuValue int64) {
	if metric.Resource.Target.AverageUtilization != nil {
		targetCpuValue = int64(*metric.Resource.Target.AverageUtilization)
	}
	for _, m := range status.CurrentMetrics {
		if m.Type == v2beta2.ResourceMetricSourceType {
			if m.Resource.Name == cpuName {
				if m.Resource.Current.AverageUtilization != nil {
					currentCpuValue = int64(*m.Resource.Current.AverageUtilization)
				}
			}
		}
	}
	return targetCpuValue, currentCpuValue
}

func calQpsMetricValue(metric v2beta2.MetricSpec, status v2beta2.HorizontalPodAutoscalerStatus) (targetQpsValue, currentQpsValue int64) {
	if metric.External.Target.AverageValue != nil {
		targetQpsValue = metric.External.Target.AverageValue.Value()
	}
	for _, m := range status.CurrentMetrics {
		if m.Type == v2beta2.ExternalMetricSourceType {
			if externalMetricsType(m.External.Metric.Name) == qpsName {
				if m.External.Current.AverageValue != nil {
					currentQpsValue = m.External.Current.AverageValue.Value()
				}
			}
		}
	}
	return targetQpsValue, currentQpsValue
}

func externalMetricsType(name string) string {
	// use upper case
	if strings.Contains(name, strings.ToUpper(qpsName)) {
		return qpsName
	}
	return ""
}
