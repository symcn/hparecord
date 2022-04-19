package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/symcn/hparecord/pkg/kube"
	"github.com/symcn/pkg/metrics"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func startHTTPServer(ctx context.Context) {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", 8080),
	}
	mux := http.NewServeMux()
	metrics.RegisterHTTPHandler(func(pattern string, handler http.Handler) {
		mux.Handle(pattern, handler)
	})
	mux.HandleFunc("/record", recordFormatServer)
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		fmt.Fprint(rw, "ok")
	})
	server.Handler = mux
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if !strings.EqualFold(err.Error(), "http: Server closed") {
				klog.Error(err)
				return
			}
		}
		klog.Info("http shutdown")
	}()
	<-ctx.Done()
	server.Shutdown(context.Background())
}

func recordFormatServer(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	namespace := r.FormValue("namespace")
	if name == "" {
		w.Write([]byte("not found name"))
		return
	}
	if namespace == "" {
		w.Write([]byte("not found namespace"))
		return
	}
	cm, err := getConfigMap(namespace, name)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	am := kube.AggragageMetrics{}
	err = json.Unmarshal([]byte(cm.Data[configmapDataName]), &am)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Not found hpa record or other err: %+v", err)))
		return
	}

	// indexTmpl.Execute(os.Stdout, formatOutput(am))
	indexTmpl.Execute(w, formatOutput(am))
	return
}

func getConfigMap(namespace, name string) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	err := kube.ManagerPlaneClusterClient.Get(types.NamespacedName{Namespace: namespace, Name: name}, cm)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("not found %s/%s hpa record.", namespace, name)
		}
		return nil, err
	}
	return cm, nil
}

func formatOutput(am kube.AggragageMetrics) string {
	htmlOutput := ""
	sort.Slice(am.List, func(i, j int) bool {
		return am.List[i].Day > am.List[j].Day
	})

	for _, aggragageMetrics := range am.List {
		htmlOutput += "<tr>"

		singleClusterDataLen := 0
		for _, enhanceMetrics := range aggragageMetrics.EnhanceHpaMetrics {
			singleClusterDataLen += len(enhanceMetrics)
		}
		htmlOutput += fmt.Sprintf(`<td rowspan="%d">%s</td>`, singleClusterDataLen, aggragageMetrics.Day)

		for cluster, enhanceMetrics := range aggragageMetrics.EnhanceHpaMetrics {
			htmlOutput += fmt.Sprintf(`<td rowspan="%d">%s</td>`, len(enhanceMetrics), cluster)
			sort.Slice(enhanceMetrics, func(i, j int) bool {
				return enhanceMetrics[i].Time.After(enhanceMetrics[j].Time)
			})

			for i, enhanceMetric := range enhanceMetrics {
				if i > 0 {
					htmlOutput += "</tr><tr>"
				}
				if len(enhanceMetric.Metrics) < 1 {
					continue
				}
				htmlOutput += fmt.Sprintf("<td>%s</td><td>%s</td><td>%s</td><td>%d</td><td>%s</td>",
					enhanceMetric.Time.Local().Format("2006-01-02 15:04:05"),
					messageToEvent(enhanceMetric.Message),
					enhanceMetric.Message,
					*enhanceMetric.Metrics[0].Resource.CurrentAverageUtilization,
					enhanceMetric.Metrics[0].Resource.CurrentAverageValue.String(),
				)
			}
			htmlOutput += "</tr>"
		}

		htmlOutput += "</tr>"
	}
	return htmlOutput
}

func messageToEvent(message string) string {
	if strings.Contains(message, "below target") {
		return "缩容"
	}
	if strings.Contains(message, "above target") {
		return "扩容"
	}
	return "Unknow"
}
