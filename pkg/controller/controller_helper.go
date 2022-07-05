package controller

import (
	"regexp"
	"strings"

	"k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

const (
	appLabel = "app"

	cpuName = "cpu"
)

var (
	MetricsKinds   string
	metricsKindSet = map[string]struct{}{}
)

func initFilterMetricsKindList() {
	for _, v := range strings.Split(MetricsKinds, ",") {
		strings.ReplaceAll(v, "-", "_")
		metricsKindSet[strings.ToLower(v)] = struct{}{}
	}
}

// filterExternalMetricsKind if external metrics not in metricsKindSet, will not set prometheus metrics
func filterExternalMetricsKind(MetricsType string) bool {
	_, exist := metricsKindSet[MetricsType]
	return !exist
}

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
			err := setResourceMetrics(metric.Resource.Name, metric, hpa.Status, promLabels)
			if err != nil {
				return err
			}
			found = true
		case v2beta2.ExternalMetricSourceType:
			metricsKind := convertMetricsKind(metric.External.Metric.Name)
			if filterExternalMetricsKind(metricsKind) {
				klog.Warningf("not supported metrics Kind: %s", metricsKind)
				continue
			}

			if err := setExternalMetrics(metricsKind, metric, hpa.Status, promLabels); err != nil {
				return err
			}
			found = true
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
			if err := deleteMetrics(string(metric.Resource.Name), promLabels); err != nil {
				return err
			}
			found = true
		case v2beta2.ExternalMetricSourceType:
			metricsKind := convertMetricsKind(metric.External.Metric.Name)
			if filterExternalMetricsKind(metricsKind) {
				klog.Warningf("not supported metrics Kind: %s", metricsKind)
				continue
			}

			if err := deleteMetrics(metricsKind, promLabels); err != nil {
				return err
			}
			found = true
		}
	}
	if !found {
		klog.Warningf("hpa: %s has no supported metrics", hpaName)
		return nil
	}
	return nil
}

// setResourceMetrics calculate resource metric and set prometheus metrics, cpu, memory, and so on
func setResourceMetrics(resource corev1.ResourceName, metric v2beta2.MetricSpec, status v2beta2.HorizontalPodAutoscalerStatus, labels promLabels) error {
	targetCpuValue, currentCpuValue := calResourceMetricValue(resource, metric, status)
	value := newMetricsValue(targetCpuValue, currentCpuValue)

	client, err := newMetricsClient(string(resource))
	if err != nil {
		return err
	}

	client.setPromMetrics(labels, value)
	return nil
}

// setExternalMetrics calculate external metric and set prometheus metrics, qps, mq, and so on
func setExternalMetrics(metricsKind string, metric v2beta2.MetricSpec, status v2beta2.HorizontalPodAutoscalerStatus, labels promLabels) error {
	targetCpuValue, currentCpuValue := calExternalMetricValue(metricsKind, metric, status)
	value := newMetricsValue(targetCpuValue, currentCpuValue)

	client, err := newMetricsClient(metricsKind)
	if err != nil {
		return err
	}

	client.setPromMetrics(labels, value)
	return nil
}

// deleteMetrics delete prometheus metrics
func deleteMetrics(metricsKind string, labels promLabels) error {
	client, err := newMetricsClient(metricsKind)
	if err != nil {
		return err
	}

	client.deletePromMetrics(labels)
	return nil
}

func calResourceMetricValue(resource corev1.ResourceName, metric v2beta2.MetricSpec, status v2beta2.HorizontalPodAutoscalerStatus) (targetValue, currentValue int64) {
	if metric.Resource.Target.AverageUtilization != nil {
		targetValue = int64(*metric.Resource.Target.AverageUtilization)
	}
	for _, m := range status.CurrentMetrics {
		if m.Type == v2beta2.ResourceMetricSourceType {
			if m.Resource.Name == resource {
				if m.Resource.Current.AverageUtilization != nil {
					currentValue = int64(*m.Resource.Current.AverageUtilization)
				}
			}
		}
	}
	return targetValue, currentValue
}

func calExternalMetricValue(metricsKind string, metric v2beta2.MetricSpec, status v2beta2.HorizontalPodAutoscalerStatus) (targetValue, currentValue int64) {
	if metric.External.Target.AverageValue != nil {
		targetValue = metric.External.Target.AverageValue.Value()
	}
	for _, m := range status.CurrentMetrics {
		if m.Type == v2beta2.ExternalMetricSourceType {
			if convertMetricsKind(m.External.Metric.Name) == metricsKind {
				if m.External.Current.AverageValue != nil {
					currentValue = m.External.Current.AverageValue.Value()
				}
			}
		}
	}
	return targetValue, currentValue
}

// name: s0-QPS, keda generate it
func convertMetricsKind(name string) string {
	reg, _ := regexp.Compile("^s\\d+-(.*)")

	subMatch := reg.FindStringSubmatch(name)
	if len(subMatch) > 1 {
		kind := subMatch[1]
		strings.ReplaceAll(kind, "-", "_")
		return strings.ToLower(kind)
	}
	return ""
}
