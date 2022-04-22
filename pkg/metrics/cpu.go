package metrics

import (
	"fmt"
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

	if ok := currentReplicasGaugeVec.DeleteLabelValues(m.Cluster, m.Hpa, m.App, m.AppCode, m.ProjectCode); !ok {
		fmt.Printf("unable to delete currentReplicasGaugeVec, metrics: %+v\n", m)
	}
	if ok := minReplicasGaugeVec.DeleteLabelValues(m.Cluster, m.Hpa, m.App, m.AppCode, m.ProjectCode); !ok {
		fmt.Printf("unable to delete minReplicasGaugeVec, metrics: %+v\n", m)
	}
	if ok := maxReplicasGaugeVec.DeleteLabelValues(m.Cluster, m.Hpa, m.App, m.AppCode, m.ProjectCode); !ok {
		fmt.Printf("unable to delete maxReplicasGaugeVec, metrics: %+v\n", m)
	}
}
