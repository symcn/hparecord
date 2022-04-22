package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	labelNames = []string{
		"cluster",
		"hpa",
		"app",
		"appCode",
		"projectCode",
	}

	targetCpuValueVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "target_cpu_value",
		Help: "Gauge number of target value",
	}, labelNames)

	currentCpuValueVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "current_cpu_value",
		Help: "Gauge number of current value",
	}, labelNames)

	currentReplicasGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "current_replicas",
		Help: "Gauge number of current replicas",
	}, labelNames)

	minReplicasGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "min_replicas",
		Help: "Gauge number of min replicas",
	}, labelNames)

	maxReplicasGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "max_replicas",
		Help: "Gauge number of max replicas",
	}, labelNames)

	collectors = []prometheus.Collector{
		targetCpuValueVec,
		currentCpuValueVec,
		currentReplicasGaugeVec,
		minReplicasGaugeVec,
		maxReplicasGaugeVec,
	}
)
