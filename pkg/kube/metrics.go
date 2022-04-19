package kube

import (
	"encoding/json"
	"sort"
	"time"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/klog/v2"
)

var (
	defaultCacheDay = 15
)

type AggragageMetrics struct {
	List []SingleDayMetrics `json:"list"`
}

type SingleDayMetrics struct {
	Day               string                         `json:"day"`
	EnhanceHpaMetrics map[string][]EnhanceHpaMetrics `json:"enhanceHpaMetrics"`
}

type EnhanceHpaMetrics struct {
	Metrics []autoscalingv1.MetricStatus `json:"metrics"`
	Time    time.Time                    `json:"time"`
	Message string                       `json:"message"`
}

func BuildAggragageData(source string, cluster string, metrics string, message string) string {
	sourceAggragageData := AggragageMetrics{}
	if len(source) != 0 {
		json.Unmarshal([]byte(source), &sourceAggragageData)
	}
	m := []autoscalingv1.MetricStatus{}
	err := json.Unmarshal([]byte(metrics), &m)
	if err != nil {
		klog.Errorf("json.unmarshal autoscalingv1.MetricsStatus(%s) failed:%+v", metrics, err)
		return ""
	}

	var (
		now   = time.Now().Local()
		day   = now.Format("2006-01-02")
		isAdd = false
	)
	for _, single := range sourceAggragageData.List {
		if single.Day == day {
			if _, ok := single.EnhanceHpaMetrics[cluster]; !ok {
				single.EnhanceHpaMetrics[cluster] = []EnhanceHpaMetrics{}
			}
			single.EnhanceHpaMetrics[cluster] = append(single.EnhanceHpaMetrics[cluster], EnhanceHpaMetrics{
				Metrics: m,
				Time:    now,
				Message: message,
			})
			isAdd = true
			break
		}
	}
	if isAdd {
		b, _ := json.Marshal(sourceAggragageData)
		return string(b)
	}

	// not add, mean new day
	sourceAggragageData.List = append(sourceAggragageData.List, SingleDayMetrics{
		Day: day,
		EnhanceHpaMetrics: map[string][]EnhanceHpaMetrics{
			cluster: {
				{
					Metrics: m,
					Time:    now,
					Message: message,
				},
			},
		},
	})
	if len(sourceAggragageData.List) > defaultCacheDay {
		sort.Slice(sourceAggragageData.List, func(i, j int) bool {
			return sourceAggragageData.List[i].Day < sourceAggragageData.List[j].Day
		})

		sourceAggragageData.List = sourceAggragageData.List[1:]
	}
	b, _ := json.Marshal(sourceAggragageData)
	return string(b)
}
