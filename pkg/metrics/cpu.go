package metrics

import (
	"k8s.io/klog/v2"
)

type CpuMetric struct {
	*Metadata

	TargetCpuValue  int32
	CurrentCpuValue int32
}

func NewCpuMetric(metadata *Metadata, targetCpuValue int32, currentCpuValue int32) *CpuMetric {
	return &CpuMetric{Metadata: metadata, TargetCpuValue: targetCpuValue, CurrentCpuValue: currentCpuValue}
}

func (m *CpuMetric) SetMetrics() error {
	if err := m.setBaseMetrics(); err != nil {
		return err
	}

	gauge, err := targetCpuValueVec.GetMetricWithLabelValues(m.Cluster, m.Hpa, m.App, m.AppCode, m.ProjectCode)
	if err != nil {
		return err
	}
	gauge.Set(float64(m.TargetCpuValue))

	gauge, err = currentCpuValueVec.GetMetricWithLabelValues(m.Cluster, m.Hpa, m.App, m.AppCode, m.ProjectCode)
	if err != nil {
		return err
	}
	gauge.Set(float64(m.CurrentCpuValue))

	return nil
}

func (m *CpuMetric) deleteMetrics() {
	m.deleteBaseMetrics()

	if ok := targetCpuValueVec.DeleteLabelValues(m.Cluster, m.Hpa, m.App, m.AppCode, m.ProjectCode); !ok {
		klog.Warningf("unable to delete currentReplicasGaugeVec, metrics: %+v", m)
	}
	if ok := currentCpuValueVec.DeleteLabelValues(m.Cluster, m.Hpa, m.App, m.AppCode, m.ProjectCode); !ok {
		klog.Warningf("unable to delete minReplicasGaugeVec, metrics: %+v", m)
	}
}
