package controller

import (
	"context"
	"sync"
	"time"

	"k8s.io/api/autoscaling/v2beta2"

	"github.com/symcn/api"
	"github.com/symcn/hparecord/pkg/kube"
	symcnclient "github.com/symcn/pkg/clustermanager/client"
	"github.com/symcn/pkg/clustermanager/configuration"
	"github.com/symcn/pkg/clustermanager/handler"
	"github.com/symcn/pkg/clustermanager/predicate"
	"github.com/symcn/pkg/clustermanager/workqueue"
	"k8s.io/utils/trace"
)

type Controller struct {
	ctx context.Context

	api.MultiMingleClient
	sync.Mutex

	cpuMetricsClient *cpuMetricsClient
}

func New(ctx context.Context, mcc *symcnclient.MultiClientConfig) (*Controller, error) {
	mcc.ClusterCfgManager = configuration.NewClusterCfgManagerWithCM(
		kube.ManagerPlaneClusterClient.GetKubeInterface(),
		"sym-admin",
		map[string]string{"ClusterOwner": "sym-admin"},
		"kubeconfig.yaml",
		"status",
	)

	cc, err := symcnclient.Complete(mcc)
	if err != nil {
		return nil, err
	}

	mc, err := cc.New()
	if err != nil {
		return nil, err
	}

	cpuMetricsClient, err := newCpuMetricsClient()
	if err != nil {
		return nil, err
	}

	ctrl := &Controller{
		ctx:               ctx,
		MultiMingleClient: mc,
		cpuMetricsClient:  cpuMetricsClient,
	}
	ctrl.registryBeforeAfterHandler()

	return ctrl, nil
}

func (ctrl *Controller) Start() error {
	return ctrl.MultiMingleClient.Start(ctrl.ctx)
}

func (ctrl *Controller) registryBeforeAfterHandler() {
	go startHTTPPrometheus(ctrl.ctx)
	initFilterLabelList()
	ctrl.RegistryBeforAfterHandler(func(ctx context.Context, cli api.MingleClient) error {
		queue, err := workqueue.Completed(workqueue.NewEventQueueConfig(cli.GetClusterCfgInfo().GetName(), ctrl)).NewQueue()
		if err != nil {
			return err
		}

		go queue.Start(ctx)

		cli.AddResourceEventHandler(
			&v2beta2.HorizontalPodAutoscaler{},
			handler.NewResourceEventHandler(
				queue,
				handler.NewEventResourceHandler(),
				predicate.NamespacePredicate("*"),
			),
		)

		return nil
	})
}

func (ctrl *Controller) OnAdd(qname string, obj interface{}) (requeue api.NeedRequeue, after time.Duration, err error) {
	instance := obj.(*v2beta2.HorizontalPodAutoscaler)

	tr := trace.New("hpa-event-collector",
		trace.Field{Key: "cluster", Value: qname},
		trace.Field{Key: "namespace", Value: instance.Namespace},
		trace.Field{Key: "name", Value: instance.Name},
	)
	defer tr.LogIfLong(time.Millisecond * 100)

	if err := ctrl.handleMetrics(qname, instance); err != nil {
		return api.Requeue, time.Second * 5, err
	}
	tr.Step("handleMetrics")

	return api.Done, 0, nil
}

func (ctrl *Controller) OnUpdate(qname string, oldObj, newObj interface{}) (requeue api.NeedRequeue, after time.Duration, err error) {
	instance := newObj.(*v2beta2.HorizontalPodAutoscaler)

	tr := trace.New("hpa-event-collector",
		trace.Field{Key: "cluster", Value: qname},
		trace.Field{Key: "namespace", Value: instance.Namespace},
		trace.Field{Key: "name", Value: instance.Name},
	)
	defer tr.LogIfLong(time.Millisecond * 100)

	if err := ctrl.handleMetrics(qname, instance); err != nil {
		return api.Requeue, time.Second * 5, err
	}
	tr.Step("handleMetrics")

	return api.Done, 0, nil
}

func (ctrl *Controller) OnDelete(qname string, obj interface{}) (requeue api.NeedRequeue, after time.Duration, err error) {
	instance := obj.(*v2beta2.HorizontalPodAutoscaler)

	tr := trace.New("hpa-event-collector",
		trace.Field{Key: "cluster", Value: qname},
		trace.Field{Key: "namespace", Value: instance.Namespace},
		trace.Field{Key: "name", Value: instance.Name},
	)
	defer tr.LogIfLong(time.Millisecond * 100)

	if err := ctrl.deleteMetrics(qname, instance); err != nil {
		return api.Requeue, time.Second * 5, err
	}
	tr.Step("handleMetrics")

	return api.Done, 0, nil
}
